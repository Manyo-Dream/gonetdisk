package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manyodream/gonetdisk/configs"
	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/model"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/util"
	"gorm.io/gorm"
)

type FileService struct {
	userRepo  *repository.UserRepo
	fileRepo  *repository.FileRepo
	jwtManger *util.JWTManager
	config    *configs.Config
}

func NewFileService(userRepo *repository.UserRepo, fileRepo *repository.FileRepo, jwtManger *util.JWTManager, config *configs.Config) *FileService {
	return &FileService{
		userRepo:  userRepo,
		fileRepo:  fileRepo,
		jwtManger: jwtManger,
		config:    config,
	}
}

func (s *FileService) UploadPhyFileAndBindFile(email string, parentID uint64, fileHeader *multipart.FileHeader) (*dto.FileUploadResponse, error) {
	// 1.验证文件信息
	validResult, err := s.validFile(fileHeader)
	if err != nil {
		return nil, BadRequest(fmt.Sprintf("验证文件信息失败: %s", err))
	}

	// 2.获取用户信息
	userInfo, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound(fmt.Sprintf("用户不存在: %s", email))
		}
		// 数据库连接错误、查询错误等应该返回500
		return nil, Internal(fmt.Sprintf("查询用户信息失败: %s", err))
	}

	if userInfo.Total_Space > 0 && userInfo.Used_Space+uint64(validResult.FileSize) > userInfo.Total_Space {
		return nil, Conflict(fmt.Sprintf("空间不足：%s", err))
	}

	// 验证父文件夹是否存在
	if parentID != 0 {
		_, err := s.fileRepo.GetParentFolderByParentID(userInfo.ID, parentID)
		if err != nil {
			return nil, NotFound(fmt.Sprintf("父文件夹不存在: %s", err))
		}
	}

	// 3.保存到临时文件夹
	tempPath, err := s.saveToTemp(fileHeader)
	if err != nil {
		return nil, Internal(fmt.Sprintf("保存到临时文件夹失败: %s", err))
	}
	defer func() {
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		} else if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("临时文件已清除: %s", tempPath)
		} else {
			fmt.Printf("检查临时文件状态失败: %v", err)
		}
	}()

	defer func() {
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// 4.计算文件哈希
	fileHash, err := s.calculateFileHash(tempPath)
	if err != nil {
		return nil, Internal(fmt.Sprintf("计算文件哈希失败: %s", err))
	}

	// 5.哈希查重
	hashResult, err := s.fileRepo.HashDeduplication(fileHash)

	if err == nil {
		// 6.文件已存在
		respFileName := validResult.FileName

		// 更新物理文件引用数
		err = s.fileRepo.IncrPhyFileRefCount(hashResult.ID, 1)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新物理文件引用数失败: %s", err))
		}

		// 查询原始文件名是否与用户文件中的文件名冲突
		_, err := s.fileRepo.GetUserFileByFileName(userInfo.ID, parentID, validResult.FileName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound(fmt.Sprintf("查询用户文件记录失败: %s", err))
		}

		if err == nil {
			respFileName = fmt.Sprintf("%s_%d%s",
				strings.TrimSuffix(validResult.FileName, filepath.Ext(validResult.FileName)),
				time.Now().UnixNano(),
				filepath.Ext(validResult.FileName))
		}

		// 创建用户文件记录
		userFile := &model.UserFile{
			UserID:     userInfo.ID,
			PhysicalID: hashResult.ID,
			ParentID:   parentID,
			FileName:   respFileName,
			FileExt:    strings.ToLower(filepath.Ext(respFileName)),
			FileSize:   hashResult.FileSize,
			IsDir:      false,
		}

		err = s.fileRepo.CreateUserFile(userFile)
		if err != nil {
			return nil, Internal(fmt.Sprintf("创建用户文件记录失败: %s", err))
		}

		// 构建pathStack
		var pathStack string
		if parentID == 0 {
			pathStack = fmt.Sprintf("/0/%d", userFile.ID)
		} else {
			parentFolder, err := s.fileRepo.GetUserFolderByID(userInfo.ID, parentID)
			if err != nil {
				return nil, Internal(fmt.Sprintf("父目录不存在或不是当前用户目录: %s", err))
			}
			pathStack = parentFolder.PathStack + "/" + strconv.FormatUint(userFile.ID, 10)
		}

		// 更新用户文件表
		err = s.fileRepo.UpdateUserFilePath(userFile.ID, pathStack)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户文件表失败: %s", err))
		}

		// 更新用户已使用空间
		err = s.userRepo.IncrUserSpace(userInfo.ID, validResult.FileSize)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户已使用空间失败: %s", err))
		}

		// 返回上传响应
		return &dto.FileUploadResponse{
			UserFileID: userFile.ID,
			FileName:   respFileName,
			FileExt:    strings.ToLower(filepath.Ext(respFileName)),
			FIleSize:   validResult.FileSize,
			ParentID:   parentID,
		}, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 7.文件不存在
		// 保存到正式目录
		filePath, err := s.promoteToLocal(tempPath, validResult.FileName)
		if err != nil {
			return nil, Internal(fmt.Sprintf("保存到正式目录失败: %s", err))
		}

		// 创建物理文件记录
		phyFile := &model.PhysicalFile{
			FileHash:    fileHash,
			FileName:    validResult.FileName,
			FileExt:     strings.ToLower(filepath.Ext(validResult.FileName)),
			FileSize:    validResult.FileSize,
			FilePath:    filePath,
			StorageType: "local",
			RefCount:    1,
		}

		err = s.fileRepo.CreatePhyFile(phyFile)
		if err != nil {
			return nil, Internal(fmt.Sprintf("创建物理文件记录失败: %s", err))
		}

		// 查询原始文件名是否与用户文件中的文件名冲突
		respFileName := validResult.FileName

		_, err = s.fileRepo.GetUserFileByFileName(userInfo.ID, parentID, validResult.FileName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Internal(fmt.Sprintf("查询用户文件记录失败: %s", err))
		} else if err == nil {
			ext := filepath.Ext(validResult.FileName)
			name := strings.TrimSuffix(validResult.FileName, ext)

			respFileName = fmt.Sprintf("%s_%d%s",
				name,
				time.Now().UnixNano(),
				ext)
		}

		// 创建用户文件记录
		userFile := &model.UserFile{
			UserID:     userInfo.ID,
			PhysicalID: phyFile.ID,
			ParentID:   parentID,
			FileName:   respFileName,
			FileExt:    strings.ToLower(filepath.Ext(respFileName)),
			FileSize:   validResult.FileSize,
			IsDir:      false,
		}

		err = s.fileRepo.CreateUserFile(userFile)
		if err != nil {
			return nil, Internal(fmt.Sprintf("创建用户文件记录失败: %s", err))
		}

		// 构建pathStack
		var pathStack string
		if parentID == 0 {
			pathStack = fmt.Sprintf("/0/%d", userFile.ID)
		} else {
			parentFolder, err := s.fileRepo.GetUserFolderByID(userInfo.ID, parentID)
			if err != nil {
				return nil, Internal(fmt.Sprintf("父目录不存在或不是当前用户目录: %s", err))
			}
			pathStack = parentFolder.PathStack + "/" + strconv.FormatUint(userFile.ID, 10)
		}

		// 更新用户文件表
		err = s.fileRepo.UpdateUserFilePath(userFile.ID, pathStack)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户文件表失败: %s", err))
		}

		// 更新用户已使用空间
		err = s.userRepo.IncrUserSpace(userInfo.ID, validResult.FileSize)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户已使用空间失败: %s", err))
		}

		return &dto.FileUploadResponse{
			UserFileID: userFile.ID,
			FileName:   respFileName,
			FileExt:    strings.ToLower(filepath.Ext(respFileName)),
			FIleSize:   validResult.FileSize,
			ParentID:   parentID,
		}, nil
	} else {
		return nil, Internal(fmt.Sprintf("哈希查重失败: %s", err))
	}
}

func (s *FileService) DownloadUserFile(userID, userFileID uint64) (*dto.FileDownloadMeta, *os.File, error) {
	userFile, phyFile, err := s.fileRepo.GetFileByDownloadReq(userID, userFileID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, NotFound("文件不存在")
	} else if err != nil {
		return nil, nil, Internal(fmt.Sprintf("查询文件失败: %s", err))
	}

	if phyFile.StorageType != "local" {
		return nil, nil, NotImplemented("不支持的存储类型: " + phyFile.StorageType)
	}

	if phyFile.FilePath == "" {
		return nil, nil, NotFound("文件路径为空")
	}

	file, err := os.Open(phyFile.FilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, NotFound("文件在磁盘上不存在")
		}
		return nil, nil, Internal(fmt.Sprintf("读取文件失败: %s", err))
	}

	return &dto.FileDownloadMeta{
		FileName:    userFile.FileName,
		StorageType: phyFile.StorageType,
		FileExt:     userFile.FileExt,
		FileSize:    phyFile.FileSize,
	}, file, nil
}

func (s *FileService) GetUserFileList(userID, parentID uint64, page, pageSize int, sortBy, orderBy string) (*dto.FileListResponse, error) {
	// 参数默认值处理
	switch {
	case page < 1:
		page = 1
	}

	switch {
	case pageSize < 1:
		pageSize = 5
	case pageSize > 100:
		pageSize = 100
	}

	switch sortBy {
	case "":
		sortBy = "updated_at"
	case "file_name", "file_size", "created_at", "updated_at":
	default:
		sortBy = "updated_at"
	}

	switch orderBy {
	case "":
		orderBy = "desc"
	case "asc", "desc":
	default:
		orderBy = "desc"
	}

	// 调用 Repository
	userFileList, total, err := s.fileRepo.GetUserFileList(userID, parentID, page, pageSize, sortBy, orderBy)
	if err != nil {
		return nil, Internal(fmt.Sprintf("获取文件列表失败: %s", err))
	}

	// 返回响应
	return &dto.FileListResponse{
		List:     userFileList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *FileService) GetTrashList(userID uint64, page, pageSize int) (*dto.TrashListResponse, error) {
	// 参数默认值处理
	switch {
	case page < 1:
		page = 1
	}

	switch {
	case pageSize < 1:
		pageSize = 5
	case pageSize > 100:
		pageSize = 100
	}

	// 调用 Repository
	trashFileList, total, err := s.fileRepo.GetTrashFileList(userID, page, pageSize)
	if err != nil {
		return nil, Internal(fmt.Sprintf("获取文件列表失败: %s", err))
	}

	return &dto.TrashListResponse{
		List:     trashFileList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *FileService) MoveFileToTrash(userID, userFileID uint64) (*dto.TrashDeleteResponse, error) {
	// 验证文件合法性
	userFile, err := s.fileRepo.GetUserFileByIDAny(userID, userFileID)
	if err != nil {
		return nil, Internal(fmt.Sprintf("获取用户文件失败: %s", err))
	}
	if userFile.IsDir {
		return nil, Internal("该项是文件夹，请调用文件夹接口")
	}

	// 根据deleted_at进行相应操作
	if !userFile.DeletedAt.Valid {
		// 调用repo移入回收站
		err := s.fileRepo.SoftDeleteUserItem(userID, userFileID)
		if err != nil {
			return nil, Internal(fmt.Sprintf("移入回收站失败: %s", err))
		}

		// 用户空间更新
		err = s.userRepo.DecrUserSpace(userID, userFile.FileSize)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户空间失败: %s", err))
		}

		// 返回响应
		return &dto.TrashDeleteResponse{Message: "文件成功移入回收站"}, nil
	} else {
		return nil, Conflict("文件已在回收站")
	}
}

// RestoreFile 将指定文件从回收站还原到原目录。
// 如果还原时发现同目录下存在同名文件，将自动生成带编号的新名称（格式：xxx(1)、xxx(2)、...），
// 直到找到不冲突的名称为止。如果(9999)仍无法找到可用名称，则使用时间戳后缀。
//
// 参数：
//   - userID: 用户ID
//   - userFileID: 待还原文件的ID
//
// 返回值：
//   - 还原成功时返回包含成功信息的响应结构体
//   - 还原失败时返回错误信息
func (s *FileService) RestoreFile(userID, userFileID uint64) (*dto.TrashRestoreResponse, error) {
	// 第一步：验证用户文件是否存在于回收站
	userFile, err := s.fileRepo.GetUserFileByIDAny(userID, userFileID)
	if err != nil {
		return nil, Internal(err.Error())
	}
	if !userFile.DeletedAt.Valid {
		return nil, Conflict("文件不在回收站，无法还原")
	}

	// 第二步：检查目标目录是否已存在同名文件（仅检查未删除的文件）
	// 如果存在同名文件，需要生成一个不冲突的新名称
	if s.isFileNameExistsInFolder(userID, userFile.ParentID, userFile.FileName, userFileID) {
		newName, err := s.generateUniqueFileName(userID, userFile.ParentID, userFile.FileName, userFile.FileExt, userFileID)
		if err != nil {
			return nil, err
		}
		userFile.FileName = newName
		err = s.fileRepo.UpdateUserFile(userFile)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户文件失败: %s", err))
		}
	}

	// 第三步：执行文件还原操作（从回收站状态恢复为正常状态）
	err = s.fileRepo.RestoreUserFile(userID, userFileID)
	if err != nil {
		return nil, Internal(fmt.Sprintf("还原用户文件失败: %s", err))
	}

	// 第四步：更新用户已使用空间（还原文件需要重新计入用户空间配额）
	err = s.userRepo.IncrUserSpace(userID, userFile.FileSize)
	if err != nil {
		return nil, Internal(fmt.Sprintf("更新用户空间失败: %s", err))
	}

	// 还原成功，返回成功信息
	return &dto.TrashRestoreResponse{Message: "文件成功还原"}, nil
}

// isFileNameExistsInFolder 检查指定文件夹中是否存在指定的文件名（排除指定文件自身）。
// 用于判断重命名时目标名称是否与现有文件冲突。
//
// 参数说明：
//   - userID: 用户ID，用于权限验证
//   - parentID: 父文件夹ID，0表示根目录
//   - fileName: 要检查的文件名
//   - excludeFileID: 排除的文件ID（通常是当前正在操作的文件，避免与自身比较）
//
// 返回值：
//   - true 表示文件名在目标文件夹中已存在（且不是排除的文件）
//   - false 表示可以使用该文件名
func (s *FileService) isFileNameExistsInFolder(userID, parentID uint64, fileName string, excludeFileID uint64) bool {
	result, err := s.fileRepo.GetUserFileByFileName(userID, parentID, fileName)
	if err != nil {
		return false
	}
	// 找到同名文件且该文件未在回收站中（DeletedAt.Valid == false），且不是自身
	return result != nil && !result.DeletedAt.Valid && result.ID != excludeFileID
}

// generateUniqueFileName 生成一个在目标文件夹中不冲突的文件名。
// 命名策略遵循 Windows 风格：
//   - 优先尝试追加编号 (1)、(2)、(3)... 直到 (9999)
//   - 如果所有编号都被占用，使用时间戳后缀作为兜底方案
//
// 参数说明：
//   - userID: 用户ID
//   - parentID: 父文件夹ID，0表示根目录
//   - baseName: 文件基础名称（不包含扩展名）
//   - ext: 文件扩展名（包括前导点号，例如 ".txt"）
//   - excludeFileID: 排除的文件ID（避免与自身比较）
//
// 返回值：
//   - 成功时返回唯一的文件名
//   - 失败时返回错误信息
func (s *FileService) generateUniqueFileName(userID, parentID uint64, baseName, ext string, excludeFileID uint64) (string, error) {
	// 定义最大尝试次数，避免无限循环
	const maxAttempts = 9999

	// 编号模式：(n) 其中 n 是从1开始的整数
	// 匹配形如 "xxx"、"xxx(1)"、"xxx(2)" 等模式
	// 正则捕获：baseName = "xxx" 或 "xxx(n)" 中的 "xxx" 部分
	pattern := regexp.MustCompile(`^(.+?)(?:\((\d+)\))?$`)
	matches := pattern.FindStringSubmatch(baseName)

	var namePart string
	if matches != nil {
		// 成功匹配，namePart 是文件名的主干部分（不含编号）
		namePart = matches[1]
	} else {
		// 匹配失败（例如基础名称为空），直接使用原名称
		namePart = baseName
	}

	// 遍历尝试所有可能的编号
	for i := 1; i <= maxAttempts; i++ {
		candidateName := fmt.Sprintf("%s(%d)%s", namePart, i, ext)
		// 检查该名称是否与目标文件夹中的现有文件冲突
		if !s.isFileNameExistsInFolder(userID, parentID, candidateName, excludeFileID) {
			return candidateName, nil
		}
	}

	// 兜底方案：当编号 (1)~(9999) 都已被占用时，使用时间戳确保唯一性
	// 时间戳精确到纳秒，冲突概率极低
	uniqueName := fmt.Sprintf("%s_%d%s", namePart, time.Now().UnixNano(), ext)
	return uniqueName, nil
}

func (s *FileService) validFile(fileHeader *multipart.FileHeader) (*model.PhysicalFile, error) {
	// 1. 获取基本信息
	fileName := fileHeader.Filename
	fileSize := fileHeader.Size

	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	// 2. 验证文件名与后缀
	err := s.VolidtateFileName(name)
	if err != nil {
		return nil, err
	}

	// 3. 验证文件大小
	// 100MB = 100 * 1024 * 1024
	maxFileSize := s.config.Upload.MaxFileSizeMB * 1024 * 1024
	if fileSize > maxFileSize {
		return nil, errors.New("上传文件过大(超过 100MB)")
	} else if fileSize <= 0 {
		return nil, errors.New("上传文件为空")
	}

	// 4. 全部验证通过，构造并返回成功响应
	return &model.PhysicalFile{
		FileName: fileName,
		FileSize: fileSize,
	}, nil
}

func (s *FileService) VolidtateFileName(fileName string) error {
	if strings.TrimSpace(fileName) == "" {
		return errors.New("文件名不能为空")
	}

	if len(fileName) > 255 {
		return errors.New("文件名过长")
	}

	illegal := `[<>:"/\\|?*]`
	matched, _ := regexp.MatchString(illegal, fileName)
	if matched {
		return errors.New("文件名包含非法字符")
	}

	if fileName == "." || fileName == ".." {
		return errors.New("文件夹名不能为 '.' 或 '..'")
	}

	return nil
}

func (s *FileService) saveToTemp(fileHeader *multipart.FileHeader) (string, error) {
	baseDir := s.config.Storage.TempDir

	subDir := time.Now().Format("2006/01/02")
	finalDir := filepath.Join(baseDir, subDir)

	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return "", err
	}

	targetName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
	finalPath := filepath.Join(finalDir, targetName)

	out, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	return finalPath, nil
}

func (s *FileService) promoteToLocal(tempPath string, originalName string) (string, error) {
	baseDir := s.config.Storage.UploadDir

	subDir := time.Now().Format("2006/01/02")
	finalDir := filepath.Join(baseDir, subDir)

	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %s", err)
	}

	finalName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), originalName)
	finalPath := filepath.Join(finalDir, finalName)

	// 优先 Rename（同分区零拷贝）
	if err := os.Rename(tempPath, finalPath); err != nil {
		// 跨分区回退到拷贝
		src, err := os.Open(tempPath)
		if err != nil {
			return "", fmt.Errorf("打开临时文件失败: %s", err)
		}
		defer src.Close()

		out, err := os.Create(finalPath)
		if err != nil {
			return "", fmt.Errorf("创建目标文件失败: %s", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			os.Remove(finalPath)
			return "", fmt.Errorf("拷贝文件失败: %s", err)
		}
		// 拷贝成功后删除临时文件
		if err := os.Remove(tempPath); err != nil {
			fmt.Printf("警告：删除临时文件失败 %s: %v\n", tempPath, err)
		}
	}

	return finalPath, nil
}

func (s *FileService) calculateFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

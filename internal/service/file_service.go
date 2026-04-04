package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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
	storage   *configs.StorageConfig
	upload    *configs.UploadConfig
}

func NewFileService(userRepo *repository.UserRepo, fileRepo *repository.FileRepo, jwtManger *util.JWTManager, storage *configs.StorageConfig, upload *configs.UploadConfig) *FileService {
	return &FileService{
		userRepo:  userRepo,
		fileRepo:  fileRepo,
		jwtManger: jwtManger,
		storage:   storage,
		upload:    upload,
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
		return nil, NotFound(fmt.Sprintf("获取用户信息失败: %s", err))
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
	// defer func() {
	// 	if _, err := os.Stat(tempPath); err == nil {
	// 		os.Remove(tempPath)
	// 	} else if errors.Is(err, os.ErrNotExist) {
	// 		fmt.Printf("临时文件已不存在: %s", tempPath)
	// 	} else {
	// 		fmt.Printf("检查临时文件失败: %v", err)
	// 	}
	// }()

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
	hashResult, err := s.fileRepo.HashDeduplication(s.fileRepo.DB, fileHash)

	if err == nil {
		// 6.文件已存在
		respFileName := validResult.FileName

		// 更新物理文件引用数
		err = s.fileRepo.IncrPhyFileRefCount(s.fileRepo.DB, hashResult.ID, 1)
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
			PhysicalID: &hashResult.ID,
			ParentID:   parentID,
			FileName:   respFileName,
			FileExt:    strings.ToLower(filepath.Ext(respFileName)),
			IsDir:      false,
		}

		err = s.fileRepo.CreateUserFile(s.fileRepo.DB, userFile)
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
		err = s.fileRepo.UpdateUserSpace(s.fileRepo.DB, userInfo.ID, validResult.FileSize)
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

		err = s.fileRepo.CreatePhyFile(s.fileRepo.DB, phyFile)
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
			PhysicalID: &phyFile.ID,
			ParentID:   parentID,
			FileName:   respFileName,
			FileExt:    strings.ToLower(filepath.Ext(respFileName)),
			IsDir:      false,
		}

		err = s.fileRepo.CreateUserFile(s.fileRepo.DB, userFile)
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
		err = s.fileRepo.UpdateUserSpace(s.fileRepo.DB, userInfo.ID, validResult.FileSize)
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
		return nil, nil, Internal("不支持的存储类型: " + phyFile.StorageType)
	}

	if phyFile.FilePath == "" {
		return nil, nil, NotFound("文件路径为空")
	}

	file, err := os.Open(phyFile.FilePath)
	if err != nil {
		return nil, nil, Internal(fmt.Sprintf("文件获取失败: %s", err))
	}

	return &dto.FileDownloadMeta{
		FileName:    userFile.FileName,
		StorageType: phyFile.StorageType,
		FileExt:     userFile.FileExt,
		FileSize:    phyFile.FileSize,
	}, file, nil
}

func (s *FileService) validFile(fileHeader *multipart.FileHeader) (*model.PhysicalFile, error) {
	// 1. 获取基本信息
	fileName := fileHeader.Filename
	fileSize := fileHeader.Size

	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	// 2. 验证文件名与后缀
	if name == "" || len(ext) <= 1 {
		return nil, errors.New("文件名或扩展名违规")
	}

	// 3. 验证文件大小
	// 100MB = 100 * 1024 * 1024
	maxFileSize := s.upload.MaxFileSizeMB * 1024 * 1024
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

func (s *FileService) saveToTemp(fileHeader *multipart.FileHeader) (string, error) {
	baseDir := s.storage.TempDir

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
	baseDir := s.storage.UploadDir

	subDir := time.Now().Format("2006/01/02")
	finalDir := filepath.Join(baseDir, subDir)

	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	finalName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), originalName)
	finalPath := filepath.Join(finalDir, finalName)

	// 优先 Rename（同分区零拷贝）
	if err := os.Rename(tempPath, finalPath); err != nil {
		// 跨分区回退到拷贝
		src, err := os.Open(tempPath)
		if err != nil {
			return "", fmt.Errorf("打开临时文件失败: %w", err)
		}
		defer src.Close()

		out, err := os.Create(finalPath)
		if err != nil {
			return "", fmt.Errorf("创建目标文件失败: %w", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			os.Remove(finalPath)
			return "", fmt.Errorf("拷贝文件失败: %w", err)
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

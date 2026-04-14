package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/model"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/util"
	"gorm.io/gorm"
)

type FolderService struct {
	userRepo  *repository.UserRepo
	fileRepo  *repository.FileRepo
	jwtManger *util.JWTManager
}

func NewFolderService(userRepo *repository.UserRepo, fileRepo *repository.FileRepo, jwtManger *util.JWTManager) *FolderService {
	return &FolderService{userRepo: userRepo, fileRepo: fileRepo, jwtManger: jwtManger}
}

func (fds *FolderService) CreateFolder(email, folderName string, parentID uint64) (*dto.FolderResponse, error) {
	// 参数校验
	if err := fds.VolidtateFolderName(folderName); err != nil {
		return nil, BadRequest(fmt.Sprintf("校验FolderName失败: %s", err.Error()))
	}

	// 获取userID
	userInfo, err := fds.userRepo.GetByEmail(email)
	if err != nil {
		return nil, NotFound(fmt.Sprintf("获取UserID失败: %s", err.Error()))
	}

	// 查询parentID对应的父文件夹是否存在(根目录0、其他目录分情况查询)
	// 不存在 -> 返回错误
	if parentID != 0 {
		_, err := fds.fileRepo.GetParentFolderByParentID(userInfo.ID, parentID)
		if err != nil {
			return nil, NotFound(fmt.Sprintf("父目录不存在或不是当前用户目录: %s", err.Error()))
		}
	}

	// 查询用户文件表是否存在同名文件夹
	// 存在 -> 增加时间后缀
	_, err = fds.fileRepo.GetUserFileByFolderName(userInfo.ID, parentID, folderName)
	if err == nil {
		folderName = fmt.Sprintf("%s_%d", folderName, time.Now().UnixNano())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, Conflict(fmt.Sprintf("检查文件夹重名失败: %s", err.Error()))
	}

	// 创建用户文件表
	userFolder := &model.UserFile{
		UserID:     userInfo.ID,
		PhysicalID: 0,
		ParentID:   parentID,
		FileName:   folderName,
		IsDir:      true,
	}
	err = fds.fileRepo.CreateUserFile(userFolder)
	if err != nil {
		return nil, Internal(fmt.Sprintf("创建用户文件夹失败: %s", err))
	}

	// 构建pathStack = parentID + ID(根目录0、其他目录分情况构建)
	var pathStack string
	if parentID == 0 {
		pathStack = fmt.Sprintf("/0/%d", userFolder.ID)
	} else {
		parentFolder, err := fds.fileRepo.GetUserFolderByID(userInfo.ID, parentID)
		if err != nil {
			return nil, NotFound(fmt.Sprintf("父目录不存在或不是当前用户目录: %s", err.Error()))
		}
		pathStack = parentFolder.PathStack + "/" + strconv.FormatUint(userFolder.ID, 10)
	}

	// 更新用户文件表
	err = fds.fileRepo.UpdateUserFilePath(userFolder.ID, pathStack)
	if err != nil {
		return nil, Internal(fmt.Sprintf("更新用户文件表失败: %s", err.Error()))
	}

	// 返回响应
	return &dto.FolderResponse{
		FolderName: userFolder.FileName,
		ParentID:   userFolder.ParentID,
		FolderID:   userFolder.ID,
	}, nil
}

func (fds *FolderService) VolidtateFolderName(folderName string) error {
	if strings.TrimSpace(folderName) == "" {
		return errors.New("文件夹名不能为空")
	}

	if len(folderName) > 255 {
		return errors.New("文件夹名过长")
	}

	illegal := `[<>:"/\\|?*]`
	matched, _ := regexp.MatchString(illegal, folderName)
	if matched {
		return errors.New("文件夹名包含非法字符")
	}

	if folderName == "." || folderName == ".." {
		return errors.New("文件夹名不能为 '.' 或 '..'")
	}

	return nil
}

func (fds *FolderService) MoveFolderToTrash(userID, userFileID uint64) (*dto.TrashDeleteResponse, error) {
	userFile, err := fds.fileRepo.GetUserFileByIDAny(userID, userFileID)
	if err != nil {
		return nil, Internal(fmt.Sprintf("获取用户文件夹失败: %s", err))
	}
	if !userFile.IsDir {
		return nil, Internal("该项是文件，请调用文件接口")
	}
	if !userFile.DeletedAt.Time.IsZero() {
		return nil, Conflict("文件夹已在回收站")
	}

	err = fds.softDeleteFolderRecursive(userID, userFileID)
	if err != nil {
		return nil, err
	}

	return &dto.TrashDeleteResponse{Message: "文件夹成功移入回收站"}, nil
}

func (fds *FolderService) softDeleteFolderRecursive(userID, folderID uint64) error {
	children, err := fds.fileRepo.GetChildrenFiles(userID, folderID)
	if err != nil {
		return Internal(fmt.Sprintf("获取子文件失败: %s", err))
	}

	for _, child := range children {
		if child.IsDir {
			err = fds.softDeleteFolderRecursive(userID, child.ID)
			if err != nil {
				return err
			}
		} else {
			err = fds.fileRepo.SoftDeleteUserItem(userID, child.ID)
			if err != nil {
				return Internal(fmt.Sprintf("移入回收站失败: %s", err))
			}
			if child.FileSize > 0 {
				err = fds.userRepo.DecrUserSpace(userID, child.FileSize)
				if err != nil {
					return Internal(fmt.Sprintf("更新用户空间失败: %s", err))
				}
			}
		}
	}

	err = fds.fileRepo.SoftDeleteUserItem(userID, folderID)
	if err != nil {
		return Internal(fmt.Sprintf("移入回收站失败: %s", err))
	}

	return nil
}

// RestoreFolder 递归还原文件夹及其所有子项（文件或子文件夹）。
// 如果还原时发现同目录下存在同名文件夹，将自动生成带编号的新名称（格式：xxx(1)、xxx(2)、...），
// 直到找到不冲突的名称为止。
//
// 参数：
//   - userID: 用户ID
//   - folderID: 待还原的文件夹ID
//
// 还原流程：
//  1. 验证文件夹是否在回收站中
//  2. 检查目标目录是否已存在同名文件夹，如存在则生成唯一名称
//  3. 将文件夹从回收站还原为正常状态
//  4. 递归还原所有子文件（文件和子文件夹）
//
// 返回值：
//   - 还原成功时返回 nil
//   - 还原失败时返回错误信息
func (fds *FolderService) RestoreFolder(userID, folderID uint64) (*dto.TrashRestoreResponse, error) {
	// 第一步：查询并验证目标文件夹是否存在于回收站
	userFolder, err := fds.fileRepo.GetUserFileByIDAny(userID, folderID)
	if err != nil {
		return nil, Internal(fmt.Sprintf("查询用户文件夹失败: %s", err))
	}
	if !userFolder.DeletedAt.Valid {
		return nil, Conflict("文件夹不在回收站")
	}

	// 第二步：检查目标目录是否已存在同名文件夹（仅检查未删除的文件夹）
	// 如果存在同名，需要生成一个不冲突的新名称
	if fds.isFolderNameExistsInFolder(userID, userFolder.ParentID, userFolder.FileName, folderID) {
		newName, err := fds.generateUniqueFolderName(userID, userFolder.ParentID, userFolder.FileName, folderID)
		if err != nil {
			return nil, Internal(fmt.Sprintf("生成文件名失败: %s", err))
		}
		userFolder.FileName = newName
		err = fds.fileRepo.UpdateUserFile(userFolder)
		if err != nil {
			return nil, Internal(fmt.Sprintf("更新用户文件夹失败: %s", err))
		}
	}

	// 第三步：执行文件夹还原操作（从回收站状态恢复为正常状态）
	err = fds.fileRepo.RestoreUserFile(userID, folderID)
	if err != nil {
		return nil, Internal(fmt.Sprintf("还原文件夹失败: %s", err))
	}

	// 第四步：递归还原所有子项（文件或子文件夹）
	children, err := fds.fileRepo.GetTrashChildrenFiles(userID, folderID)
	if err != nil {
		return nil, Internal(fmt.Sprintf("获取子文件失败: %s", err))
	}

	for _, child := range children {
		if child.IsDir {
			// 子项是文件夹，递归还原
			_, err = fds.RestoreFolder(userID, child.ID)
			if err != nil {
				return nil, Internal(fmt.Sprintf("还原文件夹失败: %s", err))
			}
		} else {
			// 子项是文件，直接还原
			err = fds.fileRepo.RestoreUserFile(userID, child.ID)
			if err != nil {
				return nil, Internal(fmt.Sprintf("还原文件失败: %s", err))
			}
		}
	}

	return &dto.TrashRestoreResponse{Message: "文件夹成功还原"}, nil
}

// isFolderNameExistsInFolder 检查指定文件夹中是否存在指定的文件夹名称（排除指定文件夹自身）。
// 用于判断重命名时目标名称是否与现有文件夹冲突。
//
// 参数说明：
//   - userID: 用户ID，用于权限验证
//   - parentID: 父文件夹ID，0表示根目录
//   - folderName: 要检查的文件夹名称
//   - excludeFolderID: 排除的文件夹ID（通常是当前正在操作的文件，避免与自身比较）
//
// 返回值：
//   - true 表示文件夹名称在目标目录中已存在（且不是排除的文件）
//   - false 表示可以使用该文件夹名称
func (fds *FolderService) isFolderNameExistsInFolder(userID, parentID uint64, folderName string, excludeFolderID uint64) bool {
	result, err := fds.fileRepo.GetUserFileByFileName(userID, parentID, folderName)
	if err != nil {
		return false
	}
	// 找到同名文件夹且该文件夹未在回收站中（DeletedAt.Valid == false），且不是自身
	return result != nil && !result.DeletedAt.Valid && result.ID != excludeFolderID
}

// generateUniqueFolderName 生成一个在目标文件夹中不冲突的文件夹名称。
// 命名策略遵循 Windows 风格：
//   - 优先尝试追加编号 (1)、(2)、(3)... 直到 (9999)
//   - 如果所有编号都被占用，使用时间戳后缀作为兜底方案
//
// 参数说明：
//   - userID: 用户ID
//   - parentID: 父文件夹ID，0表示根目录
//   - baseName: 文件夹基础名称
//   - excludeFolderID: 排除的文件夹ID（避免与自身比较）
//
// 返回值：
//   - 成功时返回唯一的文件夹名称
//   - 失败时返回错误信息
func (fds *FolderService) generateUniqueFolderName(userID, parentID uint64, baseName string, excludeFolderID uint64) (string, error) {
	// 定义最大尝试次数，避免无限循环
	const maxAttempts = 9999

	// 编号模式：(n) 其中 n 是从1开始的整数
	// 匹配形如 "xxx"、"xxx(1)"、"xxx(2)" 等模式
	// 正则捕获：namePart = "xxx" 或 "xxx(n)" 中的 "xxx" 部分
	pattern := regexp.MustCompile(`^(.+?)(?:\((\d+)\))?$`)
	matches := pattern.FindStringSubmatch(baseName)

	var namePart string
	if matches != nil {
		// 成功匹配，namePart 是文件夹名称的主干部分（不含编号）
		namePart = matches[1]
	} else {
		// 匹配失败（例如基础名称为空），直接使用原名称
		namePart = baseName
	}

	// 遍历尝试所有可能的编号
	for i := 1; i <= maxAttempts; i++ {
		candidateName := fmt.Sprintf("%s(%d)", namePart, i)
		// 检查该名称是否与目标目录中的现有文件夹冲突
		if !fds.isFolderNameExistsInFolder(userID, parentID, candidateName, excludeFolderID) {
			return candidateName, nil
		}
	}

	// 兜底方案：当编号 (1)~(9999) 都已被占用时，使用时间戳确保唯一性
	// 时间戳精确到纳秒，冲突概率极低
	uniqueName := fmt.Sprintf("%s_%d", namePart, time.Now().UnixNano())
	return uniqueName, nil
}

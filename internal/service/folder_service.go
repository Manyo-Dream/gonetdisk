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
		PhysicalID: nil,
		ParentID:   parentID,
		FileName:   folderName,
		IsDir:      true,
	}
	err = fds.fileRepo.CreateUserFile(fds.fileRepo.DB, userFolder)
	if err != nil {
		return nil, Internal(fmt.Sprintf("创建用户文件夹失败: %s", err))
	}

	// 构建pathStack = parentID + ID(根目录0、其他目录分情况构建)
	var pathStack string
	if parentID == 0 {
		pathStack = fmt.Sprintf("/0/%d", userFolder.ID)
	} else {
		parentFolder, err := fds.fileRepo.GetUserFileByID(userInfo.ID, parentID)
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
	return &dto.FolderResponse{FolderName: userFolder.FileName, ParentID: userFolder.ParentID}, nil
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

package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

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
}

func NewFileService(userRepo *repository.UserRepo, fileRepo *repository.FileRepo, jwtManger *util.JWTManager) *FileService {
	return &FileService{userRepo: userRepo, fileRepo: fileRepo, jwtManger: jwtManger}
}

func (s *FileService) UploadPhyFileAndBindFile(email string, fileHeader *multipart.FileHeader) (*dto.FileUploadResponse, error) {
	// 验证文件
	validResult, err := s.validFile(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("验证文件失败: %w", err)
	}

	// 获取用户信息
	userInfo, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("未找到用户: %w", err)
	}

	// 创建文件流用于计算 Hash
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 计算文件哈希
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("文件哈希计算失败: %w", err)
	}
	fileHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// 定义变量存储事务结果
    var (
        physicalID uint64
        finalPath  string
        finalName  string
    )

	// 开启事务
	err = s.fileRepo.DB.Transaction(func(tx *gorm.DB) error {
		// 查询哈希去重
		result, err := s.fileRepo.HashDeduplication(tx, fileHash)
		if err == nil {
			// 文件已存在
			physicalID = result.ID
			finalPath = result.FilePath
			finalName = result.FileName
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// 文件不存在
			file.Seek(0, 0)
			filePath, err := s.saveToLocal(fileHeader)
			if err != nil {
				return err
			}

			// 创建文件结构体
			phyFile := &model.PhysicalFile{
				FileName: validResult.FileName,
				FileSize: validResult.FileSize,
				FileHash: fileHash,
				FilePath: filePath,
			}

			// 插入物理文件表
			err = s.fileRepo.CreatePhyFile(tx, phyFile)
			if err != nil {
				return err
			}

			// 获取数据
			physicalID = phyFile.ID
			finalPath = filePath
			finalName = phyFile.FileName
		} else {
			return err
		}

		userFile := &model.UserFile{
			UserID:     userInfo.ID,
			PhysicalID: physicalID,
			FileName:   finalName,
			FileExt:    filepath.Ext(finalPath),
		}
		// 查询用户表是否存在同名文件

		// 有则将文件名增加时间后缀

		// 创建用户表记录

		// 否则，直接创建用户表记录
		if err := s.fileRepo.CreateUserFile(tx, userFile); err != nil {
			return fmt.Errorf("failed to bind user file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.FileUploadResponse{
		DownloadURL: finalPath,
		FileName:    finalName,
	}, nil
}

func (s *FileService) saveToLocal(fileHeader *multipart.FileHeader) (string, error) {
	baseDir := "./storage/uploads"

	subDir := time.Now().Format("2006/01/02")
	finalDir := filepath.Join(baseDir, subDir)

	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return "", err
	}

	targetName := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
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

func (s *FileService) validFile(fileHeader *multipart.FileHeader) (*model.PhysicalFile, error) {
	// 1. 获取基本信息
	fileName := fileHeader.Filename
	fileSize := fileHeader.Size
	ext := filepath.Ext(fileName)

	// 2. 验证文件名与后缀
	if fileName == "" {
		return nil, errors.New("文件名为空")
	}
	if ext == "" {
		return nil, errors.New("无法识别文件后缀")
	}
	// 建议：使用 strings.ToLower 处理大小写不敏感的情况
	if strings.ToLower(ext) != ".jpg" {
		return nil, errors.New("只支持 jpg 格式")
	}

	// 3. 验证文件大小
	// 10MB = 10 * 1024 * 1024
	const maxFileSize = 10 * 1024 * 1024
	if fileSize > maxFileSize {
		return nil, errors.New("上传文件过大(超过 10MB)")
	} else if fileSize <= 0 {
		return nil, errors.New("上传文件为空")
	}

	// 4. 全部验证通过，构造并返回成功响应
	return &model.PhysicalFile{
		FileName: fileName,
		FileSize: fileSize,
	}, nil
}

package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/model"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/util"
	"gorm.io/gorm"
)

type FileService struct {
	fileRepo  *repository.FileRepo
	jwtManger *util.JWTManager
}

func NewFileService(fileRepo *repository.FileRepo, jwtManger *util.JWTManager) *FileService {
	return &FileService{fileRepo: fileRepo, jwtManger: jwtManger}
}

func (s *FileService) UploadPhyFile(fileHeader *multipart.FileHeader) (*dto.FileUploadResponse, error) {
	// 创建文件源
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	// 验证文件扩展名
	fileName := fileHeader.Filename
	if fileName == "" {
		return nil, errors.New("文件名为空")
	} else if filepath.Ext(fileName) == "" {
		return nil, errors.New("无法识别文件后缀")
	} else if filepath.Ext(fileName) != ".jpg" {
		return nil, errors.New("只支持 jpg 格式")
	}

	// 验证文件大小
	fileSize := fileHeader.Size
	if fileSize > 1024*1024*10 {
		return nil, errors.New("上传文件过大")
	} else if fileSize <= 0 {
		return nil, errors.New("上传文件为空")
	}

	// 计算文件哈希
	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, errors.Join(err)
	}
	fileHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// 查询哈希去重
	result, err := s.fileRepo.HashDeduplication(fileHash)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.New("重置文件指针失败")
	}

	// 文件已存在
	if result != nil {
		return &dto.FileUploadResponse{
			DownloadURL: result.FilePath,
			FileName:    result.FileName,
		}, nil
	}

	// 文件不存在
	// 保存文件
	filePath, err := s.saveToLocal(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("文件写入磁盘失败: %w", err)
	}

	// 更新物理文件表
	phyFile := &model.PhyscialFile{
		FileHash: fileHash,
		FileName: fileName,
		FileSize: fileSize,
		FilePath: filePath,
	}
	err = s.fileRepo.CreatePhyFile(phyFile)
	if err != nil {
		return nil, err
	}

	return &dto.FileUploadResponse{
		DownloadURL: phyFile.FilePath,
		FileName:    phyFile.FileName,
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

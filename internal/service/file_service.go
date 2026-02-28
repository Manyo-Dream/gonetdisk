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
	// 1.验证文件信息
	validResult, err := s.validFile(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("验证文件信息失败: %w", err)
	}

	// 2.获取用户信息
	userInfo, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("未找到用户: %w", err)
	}

	// 4.保存到临时文件夹
	tempPath, err := s.saveToTemp(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("保存到临时文件夹失败: %w", err)
	}
	defer func() {
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// 5.计算文件哈希
	fileHash, err := s.calculateFileHash(tempPath)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 6.事务处理
	// 6.1.定义变量存储事务结果
	var (
		physicalID uint64
		finalName  string
		isNew      bool
	)

	// 6.2.开启事务
	err = s.fileRepo.DB.Transaction(func(tx *gorm.DB) error {
		// 6.2.1.查询哈希去重
		result, err := s.fileRepo.HashDeduplication(tx, fileHash)
		if err == nil {
			// 文件已存在
			physicalID = result.ID
			finalName = result.FileName
			isNew = false
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// 文件不存在
			isNew = true

			phyFile := &model.PhysicalFile{
				FileName: validResult.FileName,
				FileSize: validResult.FileSize,
				FileHash: fileHash,
				FilePath: "",
			}

			// 6.2.2.插入物理文件表
			err = s.fileRepo.CreatePhyFile(tx, phyFile)
			if err != nil {
				return fmt.Errorf("创建物理文件记录失败: %w", err)
			}

			// 6.2.3.获取数据
			physicalID = phyFile.ID
			finalName = phyFile.FileName
		} else {
			return fmt.Errorf("哈希查询失败: %w", err)
		}

		// 6.2.4.查询用户表是否存在同名文件
		count, err := s.fileRepo.GetUserFileNumByFileName(userInfo.ID, finalName)
		if err != nil {
			return fmt.Errorf("查询同名文件失败: %w", err)
		}
		if count > 0 {
			// 有同名文件，将文件名增加时间后缀，photo.jpg → photo_20260220_150405.jpg
			ext := filepath.Ext(finalName)
			nameWithoutExt := strings.TrimSuffix(finalName, ext)
			timeSuffix := time.Now().Format("_20060102_150405")
			finalName = nameWithoutExt + timeSuffix + ext
		}

		// 6.2.5.创建用户表记录
		userFile := &model.UserFile{
			UserID:     userInfo.ID,
			PhysicalID: physicalID,
			FileName:   finalName,
			FileExt:    filepath.Ext(validResult.FileName),
		}

		if err := s.fileRepo.CreateUserFile(tx, userFile); err != nil {
			return fmt.Errorf("绑定用户文件失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 7.是新文件，移动到本地文件夹
	var finalPath string

    if isNew {
        // 新文件：临时文件转正式存储
        localPath, err := s.promoteToLocal(tempPath, fileHeader.Filename)
        if err != nil {
            // TODO: 记录日志，后续补偿任务处理
            return nil, fmt.Errorf("转存正式文件失败: %w", err)
        }

        // 回填物理文件的存储路径
        if err := s.fileRepo.UpdatePhyFilePath(physicalID, localPath); err != nil {
            return nil, fmt.Errorf("更新文件路径失败: %w", err)
        }

		finalPath = localPath
    } else {
        // 秒传：临时文件由 defer 清理，拿已有文件的路径
        existing, _ := s.fileRepo.GetPhyFileByID(physicalID)
        if existing != nil {
            finalPath = existing.FilePath
        }
    }

    return &dto.FileUploadResponse{
        DownloadURL: finalPath,
        FileName:    finalName,
    }, nil
}

func (s *FileService) saveToTemp(fileHeader *multipart.FileHeader) (string, error) {
	baseDir := "./storage/temp"

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

func (s *FileService) promoteToLocal(tempPath string, originalName string) (string, error) {
    baseDir := "./storage/uploads"

    subDir := time.Now().Format("2006/01/02")
    finalDir := filepath.Join(baseDir, subDir)

    if err := os.MkdirAll(finalDir, 0755); err != nil {
        return "", err
    }

    targetName := fmt.Sprintf("%d_%s", time.Now().Unix(), originalName)
    localPath := filepath.Join(finalDir, targetName)

    // 优先 Rename（同分区零拷贝）
    if err := os.Rename(tempPath, localPath); err != nil {
        // 跨分区回退到拷贝
        src, err := os.Open(tempPath)
        if err != nil {
            return "", err
        }
        defer src.Close()

        out, err := os.Create(localPath)
        if err != nil {
            return "", err
        }
        defer out.Close()

        if _, err := io.Copy(out, src); err != nil {
            return "", err
        }
        // 拷贝成功后删除临时文件
        os.Remove(tempPath)
    }

    return localPath, nil
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
	// if strings.ToLower(ext) != ".jpg" {
	// 	return nil, errors.New("只支持 jpg 格式")
	// }

	// 3. 验证文件大小
	// 10MB = 10 * 1024 * 1024
	const maxFileSize = 1000 * 1024 * 1024
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

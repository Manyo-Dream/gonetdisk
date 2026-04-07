package repository

import (
	"fmt"

	"github.com/manyodream/gonetdisk/internal/model"
	"gorm.io/gorm"
)

type FileRepo struct {
	DB *gorm.DB
}

func NewFileRepo(DB *gorm.DB) *FileRepo {
	return &FileRepo{DB: DB}
}

func (r *FileRepo) CreatePhyFile(DB *gorm.DB, phyfile *model.PhysicalFile) error {
	return DB.Create(phyfile).Error
}

func (r *FileRepo) GetPhyFileByFileName(filename string) (*model.PhysicalFile, error) {
	var phyfile model.PhysicalFile

	err := r.DB.Model(&model.PhysicalFile{}).Where("file_name = ?", filename).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, nil
}

func (r *FileRepo) GetPhyFileByID(id uint64) (*model.PhysicalFile, error) {
	var phyFile model.PhysicalFile
	err := r.DB.Where("id = ?", id).First(&phyFile).Error
	return &phyFile, err
}

func (r *FileRepo) HashDeduplication(DB *gorm.DB, fileHash string) (*model.PhysicalFile, error) {
	var phyfile model.PhysicalFile

	err := DB.Model(&model.PhysicalFile{}).Where("file_hash = ?", fileHash).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, err
}

func (r *FileRepo) CreateUserFile(DB *gorm.DB, userfile *model.UserFile) error {
	return DB.Create(userfile).Error
}

func (r *FileRepo) GetUserFileByFolderName(userID, parentID uint64, folderName string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Where("user_id = ? AND parent_id = ? AND file_name = ? AND is_dir = ?", userID, parentID, folderName, true).First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetParentFolderByParentID(userID, parentID uint64) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Where("user_id = ? AND id = ? AND is_dir = ?", userID, parentID, true).First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetUserFolderByID(userID, parentID uint64) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Where("user_id = ? AND id = ? AND is_dir = ?", userID, parentID, true).First(&userfile).Error
	if err != nil {
		return nil, err
	}
	return &userfile, nil
}

func (r *FileRepo) GetUserFileByID(userID, userFileID uint64) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Model(&model.UserFile{}).Where("user_id = ? AND id = ? AND is_dir = ?", userID, userFileID, false).First(&userfile).Error
	if err != nil {
		return nil, err
	}
	return &userfile, nil
}

func (r *FileRepo) GetParentIDyFolderName(userID uint64, folderName string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Where("user_id = ? AND file_name = ? AND is_dir = ?", userID, folderName, true).First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetUserFileByFileName(userID, parentID uint64, filename string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Model(&model.UserFile{}).
		Where("user_id = ? AND parent_id = ? AND file_name = ? AND is_dir = ?", userID, parentID, filename, false).
		First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetUserFileList(userID, parentID uint64, page, pageSize int, sortBy, orderBy string) ([]model.UserFile, int64, error) {
	// 构建查询
	query := r.DB.Model(&model.UserFile{}).
		Where("user_id = ? AND parent_id = ?", userID, parentID)

	// 统计总数
	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取文件总数失败: %w", err)
	}

	if total == 0 {
		return []model.UserFile{}, 0, nil
	}

	// 分页查询
	var items []model.UserFile

	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("is_dir DESC, %s %s", sortBy, orderBy)

	err := query.Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, fmt.Errorf("构建文件列表失败: %w", err)
	}

	return items, total, nil
}

func (r *FileRepo) UpdatePhyFilePath(id uint64, filePath string) error {
	return r.DB.Model(&model.PhysicalFile{}).Where("id = ?", id).Update("file_path", filePath).Error
}

func (r *FileRepo) GetSpace(userid string) (uint64, error) {
	var user model.User
	err := r.DB.Where("user_id = ?", userid).First(&user).Error
	return user.Total_Space, err
}

func (r *FileRepo) IncrPhyFileRefCount(DB *gorm.DB, id uint64, delta int) error {
	return DB.Model(&model.PhysicalFile{}).
		Where("id = ?", id).
		UpdateColumn("ref_count", gorm.Expr("COALESCE(ref_count, 0) + ?", delta)).
		Error
}

func (r *FileRepo) UpdateUserSpace(DB *gorm.DB, userID uint64, fileSize int64) error {
	return DB.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("used_space", gorm.Expr("COALESCE(used_space, 0) + ?", fileSize)).
		Error
}

func (r *FileRepo) GetPhyFileByUserFile(userID, userFileID uint64) (*uint64, error) {
	var result struct {
		PhysicalID *uint64 `gorm:"column:physical_id"`
	}

	err := r.DB.Model(&model.UserFile{}).
		Select("physical_id").
		Where("id = ? AND user_id = ? AND is_dir = ?", userFileID, userID, false).
		Take(&result).Error
	if err != nil {
		return nil, err
	}

	if result.PhysicalID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return result.PhysicalID, nil
}

func (r *FileRepo) GetFileByDownloadReq(userID, userFileID uint64) (*model.UserFile, *model.PhysicalFile, error) {
	var userFile model.UserFile

	err := r.DB.
		Preload("PhysicalFile"). //预载PhysicalFile表，通过相连外键进行查询
		Where("id = ? AND user_id = ? AND is_dir = ?", userFileID, userID, false).
		Take(&userFile).Error
	if err != nil {
		return nil, nil, err
	}

	if userFile.PhysicalID == nil || userFile.PhysicalFile == nil {
		return nil, nil, gorm.ErrRecordNotFound
	}

	return &userFile, userFile.PhysicalFile, nil
}

func (r *FileRepo) UpdateUserFilePath(id uint64, pathStack string) error {
	return r.DB.Model(&model.UserFile{}).Where("id = ?", id).Update("path_stack", pathStack).Error
}

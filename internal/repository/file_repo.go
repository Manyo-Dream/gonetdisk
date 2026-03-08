package repository

import (
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

func (r *FileRepo) CreateUserFile(DB *gorm.DB, userfile *model.UserFile) error {
	return DB.Create(userfile).Error
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

func (r *FileRepo) UpdatePhyFilePath(id uint64, filePath string) error {
	return r.DB.Model(&model.PhysicalFile{}).Where("id = ?", id).Update("file_path", filePath).Error
}

func (r *FileRepo) GetUserFileByFileName(userID uint64, filename string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.DB.Model(&model.UserFile{}).Where("user_id = ? AND file_name = ?", userID, filename).First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) HashDeduplication(DB *gorm.DB, fileHash string) (*model.PhysicalFile, error) {
	var phyfile model.PhysicalFile

	err := DB.Model(&model.PhysicalFile{}).Where("file_hash = ?", fileHash).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, err
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

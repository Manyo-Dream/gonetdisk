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

	err := r.DB.Where("file_name = ?", filename).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, nil
}

func (r *FileRepo) HashDeduplication(db *gorm.DB, fileHash string) (*model.PhysicalFile, error) {
	var phyfile model.PhysicalFile

	err := db.Where("file_hash = ?", fileHash).First(&phyfile).Error
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

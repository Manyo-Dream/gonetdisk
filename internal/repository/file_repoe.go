package repository

import (
	"github.com/manyodream/gonetdisk/internal/model"
	"gorm.io/gorm"
)

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{db: db}
}

func (r *FileRepo) CreatePhyFile(phyfile *model.PhyscialFile) error {
	return r.db.Create(phyfile).Error
}

func (r *FileRepo) CreateUserFile(userfile *model.UserFile) error {
	return r.db.Create(userfile).Error
}

func (r *FileRepo) HashDeduplication(fileHash string) (*model.PhyscialFile, error) {
	var phyfile model.PhyscialFile

	err := r.db.Where("file_hash = ?", fileHash).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, err
}

func (r *FileRepo) GetSpace(userid string) (uint64, error) {
	var user model.User
	err := r.db.Where("user_id = ?", userid).First(&user).Error
	return user.Total_Space, err
}

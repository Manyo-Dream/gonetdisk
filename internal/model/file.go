package model

import (
	"time"

	"gorm.io/gorm"
)

type PhysicalFile struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	FileHash    string         `gorm:"unique" json:"file_hash"`
	FileName    string         `gorm:"not null" json:"file_name"`
	FileExt     string         `gorm:"not null" json:"file_ext"`
	FileSize    int64          `json:"file_size"`
	FilePath    string         `json:"file_path"`
	StorageType string         `json:"storage_type"`
	RefCount    uint64         `json:"ref_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}

type UserFile struct {
	ID               uint64         `gorm:"primaryKey" json:"id"`
	UserID           uint64         `json:"user_id"`
	User             *User          `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	PhysicalID       uint64        `json:"physical_id"`
	PhysicalFile     *PhysicalFile  `gorm:"foreignKey:PhysicalID;references:ID" json:"physical_file,omitempty"`
	ParentID         uint64         `json:"parent_id"`
	FileName         string         `json:"file_name"`
	FileExt          string         `json:"file_ext"`
	FileSize         int64          `json:"file_size"`
	PathStack        string         `json:"path_stack"`
	IsDir            bool           `json:"is_dir"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at"`
}

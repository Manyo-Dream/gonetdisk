package model

import "time"

type PhyscialFile struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	FileHash    string    `gorm:"unique" json:"file_hash"`
	FileName    string    `gorm:"not null" json:"file_name"`
	FileExt     string    `gorm:"not null" json:"file_ext"`
	FileSize    int64     `json:"file_size"`
	FilePath    string    `json:"file_path"`
	StorageType string    `json:"storage_type"`
	RefCount    uint64    `json:"ref_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserFile struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	PhyscialID uint64    `json:"physical_id"`
	PhyscialFile   *PhyscialFile `gorm:"foreignKey:PhyscialID;references:ID" json:"physical,omitempty"`
	ParentID   uint64    `json:"parent_id"`
	FileName   string    `json:"file_name"`
	FileExt    string    `json:"file_ext"`
	PathStack  string    `json:"path_stack"`
	IsDir      bool      `json:"is_dir"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  time.Time `json:"deleted_at"`
}

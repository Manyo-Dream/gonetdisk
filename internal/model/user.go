package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string         `gorm:"column:username;size:64;not null;uniqueIndex" json:"username"`
	Email         string         `gorm:"column:email;size:255;not null;uniqueIndex" json:"email"`
	Password_Hash string         `gorm:"column:password_hash;size:255;not null" json:"-"`
	Avatar_Url    string         `gorm:"column:avatar_url;size:500;default:''" json:"avatar_url"`
	Used_Space    uint64         `gorm:"column:used_space;default:0" json:"used_space"`
	Total_Space   uint64         `gorm:"column:total_space;default:1073741824" json:"total_space"`
	Status        int            `gorm:"column:status;default:0;comment:'状态: 0正常, 1禁用'" json:"status"`
	Created_At    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Updated_At    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Deleted_At    gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

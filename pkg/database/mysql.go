package database

import (
	"fmt"

	"github.com/manyodream/gonetdisk/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.PhysicalFile{}, &model.UserFile{}); err != nil {
		return nil, fmt.Errorf("自动迁移操作失败: %w", err)
	}

	return db, nil
}

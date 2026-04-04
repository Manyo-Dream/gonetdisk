package configs

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port int
	Host string
	Mode string
}

type DatabaseConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	Name      string
	Charset   string
	ParseTime bool
	Loc       string
}

type JWTConfig struct {
	Secret       string
	ExpiresHours int
}

type StorageConfig struct {
	TempDir   string
	UploadDir string
}

type UploadConfig struct {
	MaxFileSizeMB int64
}

func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("yaml")

    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }

    if config.Storage.TempDir != "" {
        absPath, err := filepath.Abs(config.Storage.TempDir)
        if err != nil {
            return nil, fmt.Errorf("转换 storage.tempDir 失败: %w", err)
        }
        config.Storage.TempDir = absPath
    }

    if config.Storage.UploadDir != "" {
        absPath, err := filepath.Abs(config.Storage.UploadDir)
        if err != nil {
            return nil, fmt.Errorf("转换 storage.uploadDir 失败: %w", err)
        }
        config.Storage.UploadDir = absPath
    }

    return &config, nil
}

func (c *JWTConfig) GetTokenDuration() time.Duration {
	return time.Duration(c.ExpiresHours) * time.Hour
}

package configs

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int
	Host string
	Mode string
}

type DatabaseConfig struct {
	Type         string
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	Charset      string
	ParseTime    bool
	Loc          string
	MaxIdleConns int
	MaxOpenConns int
	LogMode      string
}

type JWTConfig struct {
	Secret    string
	ExpiresHours int
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *JWTConfig)GetTokenDuration() time.Duration {
	return time.Duration(c.ExpiresHours) * time.Hour
}

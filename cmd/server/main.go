package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/configs"
	"github.com/manyodream/gonetdisk/internal/router"
	"github.com/manyodream/gonetdisk/internal/util"
	"github.com/manyodream/gonetdisk/pkg/database"
)

func getConfigPath() string {
	// 优先使用环境变量（方便调试时覆盖）
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// 生产环境：基于可执行文件路径
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := filepath.Join(exeDir, "configs", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 调试环境：基于当前工作目录
	workDir, _ := os.Getwd()
	return filepath.Join(workDir, "configs", "config.yaml")
}

func main() {
	cfgDir := getConfigPath()

	cfg, err := configs.LoadConfig(cfgDir)

	if err != nil {
		log.Fatal("加载配置文件失败:", err)
	}

	gin.SetMode(cfg.Server.Mode)

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.Charset,
		cfg.Database.ParseTime,
		cfg.Database.Loc,
	)

	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	jwtManager := util.NewJWTManager(cfg.JWT.Secret, cfg.JWT.GetTokenDuration())

	// 设置路由
	r := router.SetupRouter(db, jwtManager, cfg)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	log.Printf("配置文件: %s", cfgDir)
	log.Printf("存储目录: temp=%s, upload=%s", cfg.Storage.TempDir, cfg.Storage.UploadDir)
	log.Printf("运行模式: %s", cfg.Server.Mode)
	log.Printf("监听地址: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

package main

import (
	"fmt"
	"log"

	"github.com/manyodream/gonetdisk/configs"
	"github.com/manyodream/gonetdisk/internal/router"
	"github.com/manyodream/gonetdisk/internal/util"
	"github.com/manyodream/gonetdisk/pkg/database"
)

func main() {
	// cfg, err := configs.LoadConfig("E:\\Go_Project\\GoNetDisk\\configs\\config.yaml")
	cfg, err := configs.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatal("加载配置文件失败:", err)
	}

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
	// fmt.Println("dsn:", dsn)
	// fmt.Println("JWT Secret:", cfg.JWT.Secret)
	// fmt.Println("JWT ExpiresHours:", cfg.JWT.ExpiresHours)
	// fmt.Println("JWT Duration:", cfg.JWT.GetTokenDuration())

	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	jwtManager := util.NewJWTManager(cfg.JWT.Secret, cfg.JWT.GetTokenDuration())

	// 设置路由
	r := router.SetupRouter(db, jwtManager)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

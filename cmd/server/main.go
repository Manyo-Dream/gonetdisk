package main

import (
	"log"

	"github.com/manyodream/gonetdisk/internal/router"
	"github.com/manyodream/gonetdisk/pkg/database"
)

func main() {
	// 数据库连接配置
	dsn := "root:gonetdisk@tcp(localhost:3306)/gonetdisk?charset=utf8mb4&parseTime=True&loc=Local"
	
	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	// 设置路由
	r := router.SetupRouter(db)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
package main

import (
	"backend/global"
	"backend/initialize"
	"backend/router"

	"go.uber.org/zap"
)

func main() {
	// 初始化所有基础组件
	initialize.InitAll()

	defer global.GVA_LOG.Sync()

	if global.GVA_DB != nil {
		// 程序结束前关闭数据库连接
		db, _ := global.GVA_DB.DB()
		defer db.Close()
	}

	// 启动路由
	r := router.InitRouter()
	global.GVA_LOG.Info("服务启动", zap.String("addr", ":8080"))

	if err := r.Run(":8080"); err != nil {
		global.GVA_LOG.Fatal("服务启动失败", zap.Error(err))
	}
}

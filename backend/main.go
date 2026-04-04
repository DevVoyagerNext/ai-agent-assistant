package main

import (
	"backend/global"
	"backend/initialize"
	"backend/router"
)

func main() {
	// 初始化所有基础组件
	initialize.InitAll()

	if global.GVA_DB != nil {
		// 程序结束前关闭数据库连接
		db, _ := global.GVA_DB.DB()
		defer db.Close()
	}

	// 启动路由
	r := router.InitRouter()
	r.Run(":8080")
}

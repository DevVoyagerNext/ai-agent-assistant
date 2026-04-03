package main

import (
	"backend/global"
	"backend/initialize"
)

func main() {
	// 初始化Viper
	global.GVA_VP = initialize.Viper()
	// 初始化GORM
	global.GVA_DB = initialize.Gorm()
	// 初始化Redis
	global.GVA_REDIS = initialize.Redis()

	if global.GVA_DB != nil {
		// 程序结束前关闭数据库连接
		db, _ := global.GVA_DB.DB()
		defer db.Close()
	}

	// 启动路由
	r := initialize.Routers()
	r.Run(":8080")
}

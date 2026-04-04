package initialize

import "backend/global"

func InitAll() {
	// 初始化Viper
	global.GVA_VP = Viper()
	// 初始化GORM
	global.GVA_DB = Gorm()
	// 初始化Redis
	global.GVA_REDIS = Redis()
}

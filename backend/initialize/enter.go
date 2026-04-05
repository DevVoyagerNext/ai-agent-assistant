package initialize

import "backend/global"

func InitAll() {
	global.GVA_VP = Viper()
	global.GVA_LOG = Zap()
	global.GVA_DB = Gorm()
	global.GVA_REDIS = Redis()
}

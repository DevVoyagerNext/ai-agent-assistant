package global

import (
	"backend/config"
	"backend/pkg/mq"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GVA_DB     *gorm.DB
	GVA_CONFIG config.Server
	GVA_VP     *viper.Viper
	GVA_REDIS  *redis.Client
	GVA_LOG    *zap.Logger
	GVA_MQ     *mq.RedisMQ
)

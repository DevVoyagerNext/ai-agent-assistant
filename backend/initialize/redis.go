package initialize

import (
	"backend/global"
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Redis() *redis.Client {
	redisConfig := global.GVA_CONFIG.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		global.GVA_LOG.Error("Redis 连接测试失败", zap.Error(err))
		return nil
	} else {
		global.GVA_LOG.Info("Redis 连接测试成功", zap.String("pong", pong))
		return client
	}
}

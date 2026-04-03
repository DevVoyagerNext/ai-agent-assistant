package initialize

import (
	"backend/global"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
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
		fmt.Println("Redis 连接测试失败, err:", err)
		return nil
	} else {
		fmt.Println("Redis 连接测试成功, 响应:", pong)
		return client
	}
}

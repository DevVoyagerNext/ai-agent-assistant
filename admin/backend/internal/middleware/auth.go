package middleware

import (
	"time"

	"admin/backend/internal/config"
	"admin/backend/internal/httpx"
	adminsession "admin/backend/internal/session"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AdminAuth(redisClient *redis.Client, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := adminsession.ReadID(c)
		if sessionID == "" {
			httpx.Unauthorized(c, "请先登录管理员账号")
			c.Abort()
			return
		}

		ttl := time.Duration(cfg.Session.ExpiresTime) * time.Second
		if ttl <= 0 {
			ttl = 7 * 24 * time.Hour
		}
		data, err := adminsession.Validate(c.Request.Context(), redisClient, sessionID, ttl)
		if err != nil {
			httpx.Unauthorized(c, "登录状态已失效或账号已在其他设备登录")
			c.Abort()
			return
		}
		if data.UserID == 0 || data.Role != "admin" || data.Status != 1 {
			httpx.Unauthorized(c, "当前账号没有管理员权限")
			c.Abort()
			return
		}

		c.Set("adminId", data.UserID)
		c.Set("adminSession", data)
		c.Next()
	}
}

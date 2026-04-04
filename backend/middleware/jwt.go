package middleware

import (
	"backend/global"
	"backend/model"
	"backend/pkg/errmsg"
	"backend/pkg/utils"
	"backend/pkg/utils/response"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// JWTAuth 中间件，用于 JWT 认证
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 Token
		tokenString := c.GetHeader("x-token")
		if tokenString == "" {
			response.FailWithCode(errmsg.UserTokenNotExist, c)
			c.Abort()
			return
		}

		// 解析 Token
		j := utils.NewJWT(
			global.GVA_CONFIG.JWT.SigningKey,
			global.GVA_CONFIG.JWT.Issuer,
			global.GVA_CONFIG.JWT.ExpiresTime,
			global.GVA_CONFIG.JWT.RefreshExpiresTime,
		)
		claims, err := j.ParseToken(tokenString)
		if err != nil {
			// Token 过期
			if strings.Contains(err.Error(), "token is expired") {
				response.FailWithCode(errmsg.UserTokenExpired, c)
				c.Abort()
				return
			}
			// 其他 Token 错误
			response.FailWithCode(errmsg.UserTokenInvalid, c)
			c.Abort()
			return
		}

		// 判断 Token 是否在 Redis 黑名单中 (可选，如果实现了黑名单机制)
		// if _, err := global.GVA_REDIS.Get(context.Background(), tokenString).Result(); err == nil {
		// 	response.FailWithCode(errmsg.UserTokenInvalid, c)
		// 	c.Abort()
		// 	return
		// }

		// 检查用户角色 (普通用户)
		if claims.Role != "user" {
			response.FailWithCode(errmsg.UserPermissionDenied, c)
			c.Abort()
			return
		}

		// 检查用户状态 (正常)
		var user model.User
		err = global.GVA_DB.Where("id = ?", claims.UserID).First(&user).Error
		if err != nil {
			response.FailWithCode(errmsg.UserNotExist, c)
			c.Abort()
			return
		}
		if user.Status != 1 { // 假设 1 为正常状态
			response.FailWithCode(errmsg.UserAccountDisabled, c)
			c.Abort()
			return
		}

		// Token 续签 (如果需要)
		// 如果当前时间距离 Token 过期时间小于 BufferTime，则生成新 Token
		if claims.ExpiresAt.Unix()-time.Now().Unix() < global.GVA_CONFIG.JWT.BufferTime {
			// 创建新的 Claims，使用长 Token 过期时间
			newClaims := j.CreateClaims(claims.UserID, claims.Role, true)
			newToken, err := j.CreateToken(newClaims)
			if err == nil {
				c.Header("new-token", newToken)
				c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
			}
		}

		// 将用户信息存储到 Context，方便后续处理
		c.Set("claims", claims)
		c.Set("userId", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

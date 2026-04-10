package service

import (
	"backend/global"
	"backend/pkg/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthService struct{}

// GetTokenPair 生成长短 Token 并在 Redis 中记录白名单
// userID: 用户 ID
// role: 用户角色
// 返回: 短 Token, 长 Token, 短 Token 过期时间戳, 错误
func (s *AuthService) GetTokenPair(ctx context.Context, userID uint, role string) (string, string, int64, error) {
	jwtUtil := utils.NewJWT(
		global.GVA_CONFIG.JWT.SigningKey,
		global.GVA_CONFIG.JWT.Issuer,
		global.GVA_CONFIG.JWT.ExpiresTime,
		global.GVA_CONFIG.JWT.RefreshExpiresTime,
	)

	// 1. 生成短 Token (用于业务请求)
	claims := jwtUtil.CreateClaims(userID, role, false)
	token, err := jwtUtil.CreateToken(claims)
	if err != nil {
		return "", "", 0, err
	}

	// 2. 生成长 Token (用于续签)
	refreshClaims := jwtUtil.CreateClaims(userID, role, true)
	refreshToken, err := jwtUtil.CreateToken(refreshClaims)
	if err != nil {
		return "", "", 0, err
	}

	// 3. 将长 Token 加入 Redis 白名单 (key: whitelist:userID:token)
	whitelistKey := fmt.Sprintf("whitelist:%d:%s", userID, refreshToken)
	// 过期时间与长 Token 过期时间一致
	err = global.GVA_REDIS.Set(ctx, whitelistKey, "1", time.Duration(global.GVA_CONFIG.JWT.RefreshExpiresTime)*time.Second).Err()
	if err != nil {
		return "", "", 0, err
	}

	return token, refreshToken, claims.ExpiresAt.Unix(), nil
}

// GetUserID 从请求中尝试获取用户 ID，如果未登录则返回 error
// 兼顾了走过 JWTAuth 中间件的请求和未走过中间件的公开请求
func (s *AuthService) GetUserID(c *gin.Context) (uint, error) {
	// 如果经过了 JWTAuth 中间件，直接从 Context 获取
	userId, exists := c.Get("userId")
	if exists {
		return userId.(uint), nil
	}

	// 对于无需登录的公开接口，尝试解析 Header 中的 token
	tokenString := c.GetHeader("x-token")
	if tokenString == "" {
		return 0, errors.New("token is empty")
	}

	j := utils.NewJWT(
		global.GVA_CONFIG.JWT.SigningKey,
		global.GVA_CONFIG.JWT.Issuer,
		global.GVA_CONFIG.JWT.ExpiresTime,
		global.GVA_CONFIG.JWT.RefreshExpiresTime,
	)
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

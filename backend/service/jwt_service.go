package service

import (
	"backend/global"
	"backend/pkg/utils"
	"context"
	"fmt"
	"time"
)

type JwtService struct{}

// GetTokenPair 生成长短 Token 并在 Redis 中记录白名单
// userID: 用户 ID
// role: 用户角色
// 返回: 短 Token, 长 Token, 短 Token 过期时间戳, 错误
func (j *JwtService) GetTokenPair(ctx context.Context, userID uint, role string) (string, string, int64, error) {
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

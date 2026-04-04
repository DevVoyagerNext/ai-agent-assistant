package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义载荷
type CustomClaims struct {
	UserID uint   `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT JWT结构体
type JWT struct {
	SigningKey          []byte
	Issuer              string
	ExpiresTime         int64
	RefreshExpiresTime  int64
}

func NewJWT(signingKey, issuer string, expiresTime, refreshExpiresTime int64) *JWT {
	return &JWT{
		SigningKey:         []byte(signingKey),
		Issuer:             issuer,
		ExpiresTime:        expiresTime,
		RefreshExpiresTime: refreshExpiresTime,
	}
}

// CreateClaims 创建载荷 (isRefresh 为 true 则使用长 Token 过期时间)
func (j *JWT) CreateClaims(userID uint, role string, isRefresh bool) CustomClaims {
	var expireTime int64
	if isRefresh {
		expireTime = j.RefreshExpiresTime
	} else {
		expireTime = j.ExpiresTime
	}

	return CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Second)),
			Issuer:    j.Issuer,
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
}

// CreateToken 创建一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken 解析 token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

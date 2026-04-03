package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptHash 离散哈希加盐加密
func BcryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// BcryptCheck 校验密码
func BcryptCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

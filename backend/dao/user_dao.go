package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"time"
)

// GetUserByID 根据ID获取用户
func GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := global.GVA_DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := global.GVA_DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

// UpdateUserLastLogin 更新最后登录时间
func UpdateUserLastLogin(ctx context.Context, id uint, loginTime time.Time) error {
	return global.GVA_DB.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("last_login_at", loginTime).Error
}

// CheckUserExist 检查用户名或邮箱是否存在
func CheckUserExist(ctx context.Context, username, email string) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error
	return count, err
}

// CreateUser 创建新用户
func CreateUser(ctx context.Context, user *model.User) error {
	return global.GVA_DB.WithContext(ctx).Create(user).Error
}

package model

import (
	"backend/global"
	"context"
	"time"
)

// User 用户表：存储系统用户基础信息
type User struct {
	ID            uint       `gorm:"primaryKey;autoIncrement;comment:用户主键ID" json:"id"`
	Username      string     `gorm:"unique;not null;type:varchar(50);comment:用户名，唯一标识" json:"username"`
	Email         string     `gorm:"unique;type:varchar(100);comment:用户邮箱，用于找回密码/登录" json:"email"`
	PasswordHash  string     `gorm:"not null;type:varchar(255);comment:加密后的用户密码" json:"-"`
	AvatarUrl     string     `gorm:"default:'';type:varchar(255);comment:头像URL地址" json:"avatarUrl"`
	Signature     string     `gorm:"default:'';type:varchar(100);comment:个性签名/标签" json:"signature"`
	AvatarImageId int        `gorm:"default:0;comment:用户头像ID，关联images表主键" json:"avatarImageId"`
	Role          string     `gorm:"default:'user';type:enum('user','admin');comment:角色：user=普通用户, admin=管理员" json:"role"`
	Status        int8       `gorm:"default:1;index:idx_status;comment:状态：1=正常, 0=禁用" json:"status"`
	LastLoginAt   *time.Time `gorm:"comment:上次登录时间" json:"lastLoginAt"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

// TableName User 表名
func (User) TableName() string {
	return "users"
}

// GetUserByID 根据ID获取用户
func GetUserByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := global.GVA_DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := global.GVA_DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

// UpdateUserLastLogin 更新最后登录时间
func UpdateUserLastLogin(ctx context.Context, id uint, loginTime time.Time) error {
	return global.GVA_DB.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("last_login_at", loginTime).Error
}

// CheckUserExist 检查用户名或邮箱是否存在
func CheckUserExist(ctx context.Context, username, email string) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error
	return count, err
}

// CreateUser 创建新用户
func CreateUser(ctx context.Context, user *User) error {
	return global.GVA_DB.WithContext(ctx).Create(user).Error
}

package model

import "time"

type User struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string     `gorm:"unique;not null;type:varchar(50)" json:"username"`
	Email         string     `gorm:"unique;type:varchar(100)" json:"email"`
	PasswordHash  string     `gorm:"not null;type:varchar(255)" json:"-"`
	AvatarUrl     string     `gorm:"default:'';type:varchar(255)" json:"avatarUrl"`
	Signature     string     `gorm:"default:'';type:varchar(100)" json:"signature"`
	AvatarImageID int        `gorm:"column:avatar_image_id;default:0" json:"avatarImageId"`
	Role          string     `gorm:"default:'user';type:enum('user','admin')" json:"role"`
	Status        int8       `gorm:"default:1;index:idx_status" json:"status"`
	LastLoginAt   *time.Time `json:"lastLoginAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}

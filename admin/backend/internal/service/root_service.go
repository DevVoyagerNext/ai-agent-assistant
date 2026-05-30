package service

import (
	"admin/backend/internal/model"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	defaultRootUsername  = "root"
	defaultRootPassword  = "123456"
	defaultRootEmail     = "root@admin.local"
	defaultRootSignature = "超级管理员"
)

func (s *AdminService) EnsureRootAdmin() error {
	user, err := s.adminDao.FindUserByUsername(defaultRootUsername)
	if err == nil {
		updates := map[string]any{}
		if user.Role != "admin" {
			updates["role"] = "admin"
		}
		if user.Status != 1 {
			updates["status"] = 1
		}
		if len(updates) == 0 {
			return nil
		}
		return s.adminDao.UpdateUser(user.ID, updates)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultRootPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	email := defaultRootEmail
	emailCount, err := s.adminDao.CountUsersByEmail(email)
	if err != nil {
		return err
	}
	if emailCount > 0 {
		email = "root-" + strconv.FormatInt(time.Now().Unix(), 10) + "@admin.local"
	}

	return s.adminDao.CreateUser(&model.User{
		Username:     defaultRootUsername,
		Email:        email,
		PasswordHash: string(hash),
		Signature:    defaultRootSignature,
		Role:         "admin",
		Status:       1,
	})
}

package service

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"
	adminsession "admin/backend/internal/session"
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *AdminService) Register(ctx context.Context, req dto.RegisterRequest, device adminsession.DeviceInfo) (dto.AuthResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Signature = strings.TrimSpace(req.Signature)

	if msg := validateRegister(req, s.cfg.Admin.InviteCode); msg != "" {
		return dto.AuthResponse{}, errors.New(msg)
	}

	count, err := s.adminDao.CountUsersByUsernameOrEmail(req.Username, req.Email)
	if err != nil {
		return dto.AuthResponse{}, errors.New("检查账号失败")
	}
	if count > 0 {
		return dto.AuthResponse{}, errors.New("用户名或邮箱已被使用")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, errors.New("密码加密失败")
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Signature:    req.Signature,
		Role:         "admin",
		Status:       1,
	}
	if err := s.adminDao.CreateUser(&user); err != nil {
		return dto.AuthResponse{}, errors.New("创建管理员账号失败")
	}

	return s.authPayload(ctx, user, device)
}

func (s *AdminService) Login(ctx context.Context, req dto.LoginRequest, device adminsession.DeviceInfo) (dto.AuthResponse, error) {
	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = strings.TrimSpace(req.Email)
	}
	if account == "" || req.Password == "" {
		return dto.AuthResponse{}, errors.New("请填写账号和密码")
	}

	var user model.User
	var err error
	if _, parseErr := mail.ParseAddress(account); parseErr == nil {
		user, err = s.adminDao.FindUserByEmail(account)
	} else {
		user, err = s.adminDao.FindUserByUsername(account)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AuthResponse{}, errors.New("管理员账号或密码错误")
		}
		return dto.AuthResponse{}, errors.New("查询管理员账号失败")
	}
	if user.Role != "admin" {
		return dto.AuthResponse{}, errors.New("该账号不是管理员账号")
	}
	if user.Status != 1 {
		return dto.AuthResponse{}, errors.New("该管理员账号已被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return dto.AuthResponse{}, errors.New("管理员账号或密码错误")
	}

	now := time.Now()
	_ = s.adminDao.UpdateLastLogin(user.ID, now)
	user.LastLoginAt = &now

	return s.authPayload(ctx, user, device)
}

func validateRegister(req dto.RegisterRequest, inviteCode string) string {
	if utf8.RuneCountInString(req.Username) < 1 || utf8.RuneCountInString(req.Username) > 20 {
		return "用户名需要控制在 1-20 个字符"
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return "邮箱格式不正确"
	}
	if len(req.Password) < 8 || len(req.Password) > 64 {
		return "密码长度需要在 8-64 位之间"
	}
	if inviteCode != "" && req.InviteCode != inviteCode {
		return "管理员邀请码不正确"
	}
	if utf8.RuneCountInString(req.Signature) > 50 {
		return "个人签名不能超过 50 个字符"
	}
	return ""
}

func (s *AdminService) authPayload(ctx context.Context, user model.User, device adminsession.DeviceInfo) (dto.AuthResponse, error) {
	ttl := time.Duration(s.cfg.Session.ExpiresTime) * time.Second
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	now := time.Now()
	expiresAt := now.Add(ttl)
	sessionID, err := adminsession.GenerateID()
	if err != nil {
		return dto.AuthResponse{}, errors.New("生成登录会话失败")
	}
	if device.LoginAt == 0 {
		device.LoginAt = now.Unix()
	}
	data := adminsession.Data{
		SessionID: sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		Device:    device,
		CreatedAt: now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}
	if err := adminsession.Save(ctx, s.redis, data, ttl); err != nil {
		return dto.AuthResponse{}, errors.New("保存登录会话失败")
	}
	return dto.AuthResponse{
		Session:   sessionID,
		SessionID: sessionID,
		ExpiresAt: expiresAt.Unix(),
		Admin:     toAdminInfo(user),
		Device: dto.SessionDeviceInfo{
			DeviceID:  device.DeviceID,
			UserAgent: device.UserAgent,
			IP:        device.IP,
			LoginAt:   device.LoginAt,
		},
	}, nil
}

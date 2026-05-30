package service

import (
	"admin/backend/internal/config"
	"admin/backend/internal/dao"

	"github.com/redis/go-redis/v9"
)

type AdminService struct {
	adminDao *dao.AdminDao
	redis    *redis.Client
	cfg      config.Config
}

func NewAdminService(adminDao *dao.AdminDao, redisClient *redis.Client, cfg config.Config) *AdminService {
	return &AdminService{adminDao: adminDao, redis: redisClient, cfg: cfg}
}

package service

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"
	"errors"
	"time"
)

func (s *AdminService) Me(adminID uint) (dto.MeResponse, error) {
	user, err := s.adminDao.FindAdmin(adminID)
	if err != nil {
		return dto.MeResponse{}, errors.New("管理员账号不存在或已失效")
	}

	stats, err := s.DashboardStats(adminID)
	if err != nil {
		return dto.MeResponse{}, errors.New("加载统计数据失败")
	}

	return dto.MeResponse{
		Admin: toAdminInfo(user),
		Stats: stats,
	}, nil
}

func (s *AdminService) DashboardStats(adminID uint) (dto.DashboardStats, error) {
	var stats dto.DashboardStats
	var err error

	if stats.PendingSubject, err = s.adminDao.CountPendingSubjects(); err != nil {
		return stats, err
	}
	if stats.PendingNode, err = s.adminDao.CountPendingNodes(); err != nil {
		return stats, err
	}
	stats.Pending = stats.PendingSubject + stats.PendingNode

	if stats.Approved, err = s.adminDao.CountAuditByAction("approve"); err != nil {
		return stats, err
	}
	if stats.Rejected, err = s.adminDao.CountAuditByAction("reject"); err != nil {
		return stats, err
	}

	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if stats.TodayReview, err = s.adminDao.CountTodayReview(adminID, startOfDay); err != nil {
		return stats, err
	}

	return stats, nil
}

func toAdminInfo(user model.User) dto.AdminInfo {
	return dto.AdminInfo{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Signature:   user.Signature,
		AvatarUrl:   user.AvatarUrl,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

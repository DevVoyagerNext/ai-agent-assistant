package service

import (
	"admin/backend/internal/dto"
	"errors"
)

func (s *AdminService) AuditRecords() (dto.ListResponse[dto.AuditRecord], error) {
	list, err := s.adminDao.ListAuditRecords(100)
	if err != nil {
		return dto.ListResponse[dto.AuditRecord]{}, errors.New("加载审批记录失败")
	}
	if list == nil {
		list = []dto.AuditRecord{}
	}
	return dto.ListResponse[dto.AuditRecord]{List: list, Total: len(list)}, nil
}

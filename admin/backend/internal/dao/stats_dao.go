package dao

import "admin/backend/internal/model"

func (d *AdminDao) CountPendingSubjects() (int64, error) {
	var count int64
	err := d.db.Model(&model.Subject{}).Where("audit_status = ?", 1).Count(&count).Error
	return count, err
}

func (d *AdminDao) CountPendingNodes() (int64, error) {
	var count int64
	err := d.db.Model(&model.KnowledgeNode{}).Where("audit_status = ?", 1).Count(&count).Error
	return count, err
}

func (d *AdminDao) CountAuditByAction(action string) (int64, error) {
	var count int64
	err := d.db.Model(&model.AuditLog{}).Where("action = ?", action).Count(&count).Error
	return count, err
}

func (d *AdminDao) CountTodayReview(adminID uint, startOfDay any) (int64, error) {
	var count int64
	err := d.db.Model(&model.AuditLog{}).
		Where("admin_id = ? AND created_at >= ?", adminID, startOfDay).
		Count(&count).Error
	return count, err
}

package dao

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"

	"gorm.io/gorm"
)

func (d *AdminDao) ListAuditRecords(limit int) ([]dto.AuditRecord, error) {
	var list []dto.AuditRecord
	err := d.db.Table("audit_logs").
		Select(`audit_logs.id, audit_logs.target_type, audit_logs.target_id, audit_logs.admin_id,
			audit_logs.action, audit_logs.remark, audit_logs.created_at,
			CASE
				WHEN audit_logs.target_type = 'subject' THEN COALESCE(subjects.name, subjects.name_draft, '')
				WHEN audit_logs.target_type = 'node' THEN COALESCE(knowledge_nodes.name, knowledge_nodes.name_draft, '')
				ELSE ''
			END AS target_name,
			COALESCE(subjects.name, node_subjects.name, '') AS subject_name,
			COALESCE(subjects.name_draft, node_subjects.name_draft, '') AS subject_draft_name,
			COALESCE(users.username, '') AS admin_name`).
		Joins("LEFT JOIN subjects ON subjects.id = audit_logs.target_id AND audit_logs.target_type = 'subject'").
		Joins("LEFT JOIN knowledge_nodes ON knowledge_nodes.id = audit_logs.target_id AND audit_logs.target_type = 'node'").
		Joins("LEFT JOIN subjects AS node_subjects ON node_subjects.id = knowledge_nodes.subject_id").
		Joins("LEFT JOIN users ON users.id = audit_logs.admin_id").
		Where("audit_logs.target_type IN ?", []string{"subject", "node"}).
		Order("audit_logs.created_at DESC").
		Limit(limit).
		Scan(&list).Error
	return list, err
}

func (d *AdminDao) CreateAuditLog(tx *gorm.DB, log *model.AuditLog) error {
	return tx.Create(log).Error
}

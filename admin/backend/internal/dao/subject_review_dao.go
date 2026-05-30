package dao

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *AdminDao) ListPendingSubjects() ([]dto.SubjectReviewItem, error) {
	var list []dto.SubjectReviewItem
	err := d.db.Table("subjects").
		Select(`subjects.id, subjects.creator_id, subjects.slug, subjects.name, subjects.name_draft,
			subjects.icon, subjects.icon_draft, subjects.description, subjects.description_draft,
			subjects.cover_image_id, subjects.cover_image_id_draft, subjects.status, subjects.audit_status,
			subjects.has_draft, subjects.created_at, COALESCE(users.username, '') AS creator_name,
			COALESCE(users.email, '') AS creator_email`).
		Joins("LEFT JOIN users ON users.id = subjects.creator_id").
		Where("subjects.audit_status = ?", 1).
		Order("subjects.created_at DESC").
		Scan(&list).Error
	return list, err
}

func (d *AdminDao) FindSubjectForUpdate(tx *gorm.DB, subjectID int) (model.Subject, error) {
	var subject model.Subject
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", subjectID).First(&subject).Error
	return subject, err
}

func (d *AdminDao) FindSubjectByID(tx *gorm.DB, subjectID int) (model.Subject, error) {
	var subject model.Subject
	err := tx.Where("id = ?", subjectID).First(&subject).Error
	return subject, err
}

func (d *AdminDao) UpdateSubject(tx *gorm.DB, subjectID uint, updates map[string]any) error {
	return tx.Model(&model.Subject{}).Where("id = ?", subjectID).Updates(updates).Error
}

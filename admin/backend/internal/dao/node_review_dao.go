package dao

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *AdminDao) ListPendingNodes() ([]dto.NodeReviewItem, error) {
	var list []dto.NodeReviewItem
	err := d.db.Table("knowledge_nodes").
		Select(`knowledge_nodes.id, knowledge_nodes.subject_id, knowledge_nodes.parent_id, knowledge_nodes.path,
			knowledge_nodes.name, knowledge_nodes.name_draft, knowledge_nodes.status, knowledge_nodes.audit_status,
			knowledge_nodes.has_draft, knowledge_nodes.draft_descendant_count, knowledge_nodes.level,
			knowledge_nodes.is_leaf, COALESCE(subjects.name, '') AS subject_name, subjects.creator_id,
			COALESCE(users.username, '') AS creator_name, COALESCE(users.email, '') AS creator_email,
			COALESCE(knowledge_contents.content, '') AS content,
			COALESCE(knowledge_contents.content_draft, '') AS content_draft,
			knowledge_contents.updated_at AS updated_at`).
		Joins("JOIN subjects ON subjects.id = knowledge_nodes.subject_id").
		Joins("LEFT JOIN users ON users.id = subjects.creator_id").
		Joins("LEFT JOIN knowledge_contents ON knowledge_contents.node_id = knowledge_nodes.id").
		Where("knowledge_nodes.audit_status = ?", 1).
		Order("knowledge_nodes.id DESC").
		Scan(&list).Error
	return list, err
}

func (d *AdminDao) FindNodeForUpdate(tx *gorm.DB, nodeID int) (model.KnowledgeNode, error) {
	var node model.KnowledgeNode
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", nodeID).First(&node).Error
	return node, err
}

func (d *AdminDao) FindContentByNodeID(tx *gorm.DB, nodeID uint) (model.KnowledgeContent, error) {
	var content model.KnowledgeContent
	err := tx.Where("node_id = ?", nodeID).First(&content).Error
	return content, err
}

func (d *AdminDao) UpdateTopNodeBySubjectID(tx *gorm.DB, subjectID uint, updates map[string]any) error {
	return tx.Model(&model.KnowledgeNode{}).
		Where("subject_id = ? AND parent_id = ?", subjectID, 0).
		Updates(updates).Error
}

func (d *AdminDao) PublishNodesBySubjectID(tx *gorm.DB, subjectID uint, logID int64) error {
	return tx.Model(&model.KnowledgeNode{}).
		Where("subject_id = ?", subjectID).
		Updates(map[string]any{
			"name":                   gorm.Expr("CASE WHEN name_draft IS NULL OR name_draft = '' THEN name ELSE name_draft END"),
			"status":                 "published",
			"audit_status":           2,
			"has_draft":              0,
			"draft_descendant_count": 0,
			"last_log_id":            logID,
		}).Error
}

func (d *AdminDao) PublishContentsBySubjectID(tx *gorm.DB, subjectID uint, logID int64) error {
	nodeIDs := tx.Model(&model.KnowledgeNode{}).Select("id").Where("subject_id = ?", subjectID)
	return tx.Model(&model.KnowledgeContent{}).
		Where("node_id IN (?)", nodeIDs).
		Updates(map[string]any{
			"content":      gorm.Expr("CASE WHEN content_draft IS NULL THEN content ELSE content_draft END"),
			"audit_status": 2,
			"has_draft":    0,
			"last_log_id":  logID,
		}).Error
}

func (d *AdminDao) UpdateNode(tx *gorm.DB, nodeID uint, updates map[string]any) error {
	return tx.Model(&model.KnowledgeNode{}).Where("id = ?", nodeID).Updates(updates).Error
}

func (d *AdminDao) UpdateContent(tx *gorm.DB, contentID uint, updates map[string]any) error {
	return tx.Model(&model.KnowledgeContent{}).Where("id = ?", contentID).Updates(updates).Error
}

func (d *AdminDao) DecrementAncestorDraftCount(tx *gorm.DB, subjectID int, ancestorIDs []int) error {
	return tx.Model(&model.KnowledgeNode{}).
		Where("subject_id = ? AND id IN ?", subjectID, ancestorIDs).
		Update("draft_descendant_count", gorm.Expr("CASE WHEN draft_descendant_count - 1 < 0 THEN 0 ELSE draft_descendant_count - 1 END")).Error
}

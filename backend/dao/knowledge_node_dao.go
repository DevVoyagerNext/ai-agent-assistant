package dao

import (
	"backend/global"
	"backend/model"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type KnowledgeNodeDao struct{}

// GetNodesByParent 统一根据 parentId 获取知识点，如果是顶级节点 parentId 传 0
func (dao *KnowledgeNodeDao) GetNodesByParent(subjectID, parentID int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	err := global.GVA_DB.Where("subject_id = ? AND parent_id = ? AND status = ?", subjectID, parentID, "published").
		Order("sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetChildNodes 仅根据 parentId 获取直属下级知识点（当不需要限制 subjectId 时，可选）
func (dao *KnowledgeNodeDao) GetChildNodes(parentID int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	err := global.GVA_DB.Where("parent_id = ? AND status = ?", parentID, "published").
		Order("sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetChildNodesWithoutStatus 获取直属下级知识点（不限制发布状态，用于创作者视角）
func (dao *KnowledgeNodeDao) GetChildNodesWithoutStatus(parentID int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	err := global.GVA_DB.Where("parent_id = ?", parentID).
		Order("sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetNodesByParentIDs 批量根据 parentId 列表获取子节点列表
func (dao *KnowledgeNodeDao) GetNodesByParentIDs(parentIDs []int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	if len(parentIDs) == 0 {
		return nodes, nil
	}
	err := global.GVA_DB.Where("parent_id IN ? AND status = ?", parentIDs, "published").
		Order("parent_id asc, sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetNodesByParentIDsWithoutStatus 批量根据 parentId 列表获取子节点列表（不限制状态）
func (dao *KnowledgeNodeDao) GetNodesByParentIDsWithoutStatus(parentIDs []int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	if len(parentIDs) == 0 {
		return nodes, nil
	}
	err := global.GVA_DB.Where("parent_id IN ?", parentIDs).
		Order("parent_id asc, sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetNodeByID 获取知识点基础信息
func (dao *KnowledgeNodeDao) GetNodeByID(nodeID int) (model.KnowledgeNode, error) {
	var node model.KnowledgeNode
	err := global.GVA_DB.Where("id = ? AND status = ?", nodeID, "published").First(&node).Error
	return node, err
}

// GetNodeByIDWithoutStatus 获取知识点基础信息（不限制状态，用于创建/编辑时的校验）
func (dao *KnowledgeNodeDao) GetNodeByIDWithoutStatus(nodeID int) (model.KnowledgeNode, error) {
	var node model.KnowledgeNode
	err := global.GVA_DB.Where("id = ?", nodeID).First(&node).Error
	return node, err
}

// CreateKnowledgeNodeWithTx 在事务中创建知识节点
func (dao *KnowledgeNodeDao) CreateKnowledgeNodeWithTx(tx *gorm.DB, node *model.KnowledgeNode) error {
	return tx.Create(node).Error
}

// GetMaxSortOrderByParent 获取同级节点下最大的 SortOrder
func (dao *KnowledgeNodeDao) GetMaxSortOrderByParent(subjectID int, parentID int) int {
	var maxSortOrder int
	global.GVA_DB.Model(&model.KnowledgeNode{}).
		Where("subject_id = ? AND parent_id = ?", subjectID, parentID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxSortOrder)
	return maxSortOrder
}

// UpdateSubjectTopNodeDraftWithTx 更新某个教材下顶级节点的草稿信息（随教材信息同步）
func (dao *KnowledgeNodeDao) UpdateSubjectTopNodeDraftWithTx(tx *gorm.DB, subjectID int, nameDraft string) error {
	return tx.Model(&model.KnowledgeNode{}).
		Where("subject_id = ? AND parent_id = 0", subjectID).
		Updates(map[string]interface{}{
			"name_draft": nameDraft,
			"has_draft":  1,
		}).Error
}

// UpdateKnowledgeNodeDraft 更新知识点节点的草稿名称
func (dao *KnowledgeNodeDao) UpdateKnowledgeNodeDraft(nodeID int, nameDraft string) error {
	return global.GVA_DB.Model(&model.KnowledgeNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"name_draft": nameDraft,
			"has_draft":  1,
		}).Error
}

// GetNodeContentByID 获取知识点内容（正文等）
func (dao *KnowledgeNodeDao) GetNodeContentByID(nodeID int) (model.KnowledgeContent, error) {
	var content model.KnowledgeContent
	err := global.GVA_DB.Where("node_id = ?", nodeID).First(&content).Error
	return content, err
}

// UpsertKnowledgeContent 更新或创建知识点正文内容草稿
func (dao *KnowledgeNodeDao) UpsertKnowledgeContent(nodeID int, contentDraft string) error {
	var content model.KnowledgeContent
	err := global.GVA_DB.Where("node_id = ?", nodeID).First(&content).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果该节点还没有内容记录，则创建一条
			content = model.KnowledgeContent{
				NodeID:       nodeID,
				Content:      "", // 发布前内容为空
				ContentDraft: contentDraft,
				AuditStatus:  0,
				HasDraft:     1,
			}
			return global.GVA_DB.Create(&content).Error
		}
		return err
	}

	// 如果存在记录，则更新其 draft 字段
	return global.GVA_DB.Model(&content).Updates(map[string]interface{}{
		"content_draft": contentDraft,
		"has_draft":     1,
	}).Error
}

// GetNodeMetricsByNodeIDs 批量获取这些节点的统计指标（难度等）
func (dao *KnowledgeNodeDao) GetNodeMetricsByNodeIDs(nodeIDs []uint) ([]model.NodeMetric, error) {
	var metrics []model.NodeMetric
	if len(nodeIDs) == 0 {
		return metrics, nil
	}
	err := global.GVA_DB.Where("node_id IN ?", nodeIDs).Find(&metrics).Error
	return metrics, err
}

// GetUserStudyStatusByNodeIDs 批量获取用户在 these 节点上的学习进度
func (dao *KnowledgeNodeDao) GetUserStudyStatusByNodeIDs(userID uint, nodeIDs []uint) ([]model.UserStudyStatus, error) {
	var statuses []model.UserStudyStatus
	if len(nodeIDs) == 0 || userID == 0 {
		return statuses, nil
	}
	err := global.GVA_DB.Where("user_id = ? AND node_id IN ?", userID, nodeIDs).Find(&statuses).Error
	return statuses, err
}

// GetUserStudyNote 获取用户对某个节点的随堂笔记
func (dao *KnowledgeNodeDao) GetUserStudyNote(userID uint, nodeID int) (model.UserStudyNote, error) {
	var note model.UserStudyNote
	err := global.GVA_DB.
		Joins("JOIN knowledge_nodes ON knowledge_nodes.id = user_study_notes.node_id").
		Where("user_study_notes.user_id = ? AND user_study_notes.node_id = ? AND knowledge_nodes.status = ?", userID, nodeID, "published").
		First(&note).Error
	return note, err
}

// UpsertUserStudyNote 创建或更新用户随堂笔记
func (dao *KnowledgeNodeDao) UpsertUserStudyNote(userID uint, nodeID int, noteContent string, isImportant int8) error {
	var note model.UserStudyNote
	err := global.GVA_DB.Where("user_id = ? AND node_id = ?", userID, nodeID).First(&note).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没查到记录，则创建
			note = model.UserStudyNote{
				UserID:      int(userID),
				NodeID:      nodeID,
				NoteContent: noteContent,
				IsImportant: isImportant,
			}
			return global.GVA_DB.Create(&note).Error
		}
		return err
	}

	// 如果查询到了记录，则更新内容和标记
	return global.GVA_DB.Model(&note).Updates(map[string]interface{}{
		"note_content": noteContent,
		"is_important": isImportant,
	}).Error
}

// UpsertUserStudyStatus 更新或创建用户在某个知识点上的学习状态
func (dao *KnowledgeNodeDao) UpsertUserStudyStatus(userID uint, nodeID int, status string) error {
	var studyStatus model.UserStudyStatus
	err := global.GVA_DB.Where("user_id = ? AND node_id = ?", userID, nodeID).First(&studyStatus).Error

	now := time.Now()
	if err != nil {
		// 没查到记录，则创建
		studyStatus = model.UserStudyStatus{
			UserID:        int(userID),
			NodeID:        nodeID,
			Status:        status,
			LastStudyTime: &now,
		}
		return global.GVA_DB.Create(&studyStatus).Error
	}

	// 如果查询到了记录，则更新状态和最后学习时间
	return global.GVA_DB.Model(&studyStatus).Updates(map[string]interface{}{
		"status":          status,
		"last_study_time": &now,
	}).Error
}

// CountTotalLeafNodes 统计某个教材下的总叶子节点数
func (dao *KnowledgeNodeDao) CountTotalLeafNodes(ctx context.Context, subjectID int) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.KnowledgeNode{}).
		Where("subject_id = ? AND is_leaf = ? AND status = ?", subjectID, 1, "published").
		Count(&count).Error
	return count, err
}

// CountLearnedLeafNodes 统计用户在某个教材下已学完的叶子节点数
func (dao *KnowledgeNodeDao) CountLearnedLeafNodes(ctx context.Context, userID uint, subjectID int) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.UserStudyStatus{}).
		Joins("JOIN knowledge_nodes ON knowledge_nodes.id = user_study_status.node_id").
		Where("user_study_status.user_id = ? AND user_study_status.status = ? AND knowledge_nodes.subject_id = ? AND knowledge_nodes.is_leaf = ? AND knowledge_nodes.status = ?",
			userID, "completed", subjectID, 1, "published").
		Count(&count).Error
	return count, err
}

// GetUserNodeDifficulty 获取用户对知识点的难度评价
func (dao *KnowledgeNodeDao) GetUserNodeDifficulty(tx *gorm.DB, userID uint, nodeID int) (model.UserNodeDifficulty, error) {
	var diff model.UserNodeDifficulty
	err := tx.Where("user_id = ? AND node_id = ?", userID, nodeID).First(&diff).Error
	return diff, err
}

// UpsertUserNodeDifficulty 更新或创建用户对知识点的难度评价
func (dao *KnowledgeNodeDao) UpsertUserNodeDifficulty(tx *gorm.DB, userID uint, nodeID int, difficulty string) error {
	var diff model.UserNodeDifficulty
	err := tx.Where("user_id = ? AND node_id = ?", userID, nodeID).First(&diff).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建
			diff = model.UserNodeDifficulty{
				UserID:     int(userID),
				NodeID:     nodeID,
				Difficulty: difficulty,
			}
			return tx.Create(&diff).Error
		}
		return err
	}

	// 更新
	return tx.Model(&diff).Update("difficulty", difficulty).Error
}

// UpdateNodeMetric 原子更新节点的难度聚合指标
func (dao *KnowledgeNodeDao) UpdateNodeMetric(tx *gorm.DB, nodeID int, metricType string, delta int) error {
	var metric model.NodeMetric
	err := tx.Where("node_id = ? AND metric_type = ?", nodeID, metricType).First(&metric).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建
			if delta < 0 {
				delta = 0
			}
			metric = model.NodeMetric{
				NodeID:      nodeID,
				MetricType:  metricType,
				MetricValue: delta,
			}
			return tx.Create(&metric).Error
		}
		return err
	}

	// 更新
	return tx.Model(&metric).Update("metric_value", gorm.Expr("metric_value + ?", delta)).Error
}

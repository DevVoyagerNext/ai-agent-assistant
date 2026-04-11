package dao

import (
	"backend/global"
	"backend/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type KnowledgeNodeDao struct{}

// GetNodesByParent 统一根据 parentId 获取知识点，如果是顶级节点 parentId 传 0
func (dao *KnowledgeNodeDao) GetNodesByParent(subjectID, parentID int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	err := global.GVA_DB.Where("subject_id = ? AND parent_id = ?", subjectID, parentID).
		Order("sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetChildNodes 仅根据 parentId 获取直属下级知识点（当不需要限制 subjectId 时，可选）
func (dao *KnowledgeNodeDao) GetChildNodes(parentID int) ([]model.KnowledgeNode, error) {
	var nodes []model.KnowledgeNode
	err := global.GVA_DB.Where("parent_id = ?", parentID).
		Order("sort_order asc").
		Find(&nodes).Error
	return nodes, err
}

// GetNodeByID 获取知识点基础信息
func (dao *KnowledgeNodeDao) GetNodeByID(nodeID int) (model.KnowledgeNode, error) {
	var node model.KnowledgeNode
	err := global.GVA_DB.Where("id = ?", nodeID).First(&node).Error
	return node, err
}

// GetNodeContentByID 获取知识点内容（正文等）
func (dao *KnowledgeNodeDao) GetNodeContentByID(nodeID int) (model.KnowledgeContent, error) {
	var content model.KnowledgeContent
	err := global.GVA_DB.Where("node_id = ?", nodeID).First(&content).Error
	return content, err
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
	err := global.GVA_DB.Where("user_id = ? AND node_id = ?", userID, nodeID).First(&note).Error
	return note, err
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
		Where("subject_id = ? AND is_leaf = ?", subjectID, 1).
		Count(&count).Error
	return count, err
}

// CountLearnedLeafNodes 统计用户在某个教材下已学完的叶子节点数
func (dao *KnowledgeNodeDao) CountLearnedLeafNodes(ctx context.Context, userID uint, subjectID int) (int64, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).Model(&model.UserStudyStatus{}).
		Joins("JOIN knowledge_nodes ON knowledge_nodes.id = user_study_status.node_id").
		Where("user_study_status.user_id = ? AND user_study_status.status = ? AND knowledge_nodes.subject_id = ? AND knowledge_nodes.is_leaf = ?",
			userID, "completed", subjectID, 1).
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

package service

import (
	"backend/dao"
	"backend/dto"
	"backend/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type KnowledgeNodeService struct {
	nodeDao dao.KnowledgeNodeDao
}

// enrichNodes 内部方法：批量组装知识点的难度评价与用户进度
func (s *KnowledgeNodeService) enrichNodes(nodes []model.KnowledgeNode, userID uint) ([]dto.KnowledgeNodeItemRes, error) {
	if len(nodes) == 0 {
		return []dto.KnowledgeNodeItemRes{}, nil
	}

	var nodeIDs []uint
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}

	// 1. 获取难度评价 (node_metrics)
	metrics, err := s.nodeDao.GetNodeMetricsByNodeIDs(nodeIDs)
	if err != nil {
		return nil, err
	}

	// 组装 metrics 数据: map[nodeID]map[metricType]value
	metricMap := make(map[uint]map[string]int)
	for _, m := range metrics {
		nid := uint(m.NodeID)
		if metricMap[nid] == nil {
			metricMap[nid] = make(map[string]int)
		}
		metricMap[nid][m.MetricType] = m.MetricValue
	}

	// 2. 获取用户学习进度 (user_study_status)
	var statusMap = make(map[uint]string)
	if userID > 0 {
		statuses, err := s.nodeDao.GetUserStudyStatusByNodeIDs(userID, nodeIDs)
		if err != nil {
			return nil, err
		}
		for _, st := range statuses {
			statusMap[uint(st.NodeID)] = st.Status
		}
	}

	// 3. 组装最终数据
	var result []dto.KnowledgeNodeItemRes
	for _, node := range nodes {
		// 默认进度为 unstarted
		progress := "unstarted"
		if st, ok := statusMap[node.ID]; ok {
			progress = st
		}

		// 难度指标
		easy := 0
		medium := 0
		hard := 0
		if mm, ok := metricMap[node.ID]; ok {
			easy = mm["easy"]
			medium = mm["medium"]
			hard = mm["hard"]
		}

		item := dto.KnowledgeNodeItemRes{
			ID:                 node.ID,
			SubjectID:          node.SubjectID,
			ParentID:           node.ParentID,
			Path:               node.Path,
			Name:               node.Name,
			Level:              node.Level,
			IsLeaf:             node.IsLeaf,
			SortOrder:          node.SortOrder,
			ImageID:            node.ImageID,
			EasyCount:          easy,
			MediumCount:        medium,
			HardCount:          hard,
			UserProgressStatus: progress,
		}
		result = append(result, item)
	}

	return result, nil
}

// GetTopLevelNodes 获取某个教材下的顶级知识点
func (s *KnowledgeNodeService) GetTopLevelNodes(ctx context.Context, subjectID int, userID uint) ([]dto.KnowledgeNodeItemRes, error) {
	nodes, err := s.nodeDao.GetNodesByParent(subjectID, 0)
	if err != nil {
		return nil, err
	}
	return s.enrichNodes(nodes, userID)
}

// GetChildNodes 获取某个知识点下的直接子节点
func (s *KnowledgeNodeService) GetChildNodes(ctx context.Context, parentID int, userID uint) ([]dto.KnowledgeNodeItemRes, error) {
	nodes, err := s.nodeDao.GetChildNodes(parentID)
	if err != nil {
		return nil, err
	}
	return s.enrichNodes(nodes, userID)
}

// GetNodeDetail 获取知识点详情（包含正文、难度评价、用户进度）
func (s *KnowledgeNodeService) GetNodeDetail(ctx context.Context, nodeID int, userID uint) (*dto.KnowledgeNodeDetailRes, error) {
	// 1. 获取基本信息
	node, err := s.nodeDao.GetNodeByID(nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("知识点不存在")
		}
		return nil, err
	}

	// 2. 组装基础信息（复用 enrichNodes 保证逻辑统一）
	enrichedItems, err := s.enrichNodes([]model.KnowledgeNode{node}, userID)
	if err != nil || len(enrichedItems) == 0 {
		return nil, err
	}
	baseInfo := enrichedItems[0]

	// 3. 获取正文
	var contentStr string
	content, err := s.nodeDao.GetNodeContentByID(nodeID)
	if err == nil {
		contentStr = content.Content
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 4. 返回完整结构
	return &dto.KnowledgeNodeDetailRes{
		KnowledgeNodeItemRes: baseInfo,
		Content:              contentStr,
	}, nil
}

// GetUserStudyNote 获取用户对某个知识点的随堂笔记
func (s *KnowledgeNodeService) GetUserStudyNote(ctx context.Context, nodeID int, userID uint) (*dto.UserStudyNoteRes, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}
	note, err := s.nodeDao.GetUserStudyNote(userID, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 若没写过笔记，返回空对象而不是错误
			return &dto.UserStudyNoteRes{
				NodeID:      nodeID,
				NoteContent: "",
				IsImportant: 0,
			}, nil
		}
		return nil, err
	}

	return &dto.UserStudyNoteRes{
		ID:          note.ID,
		NodeID:      note.NodeID,
		NoteContent: note.NoteContent,
		IsImportant: note.IsImportant,
		UpdatedAt:   note.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

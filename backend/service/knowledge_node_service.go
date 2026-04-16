package service

import (
	"backend/dao"
	"backend/dto"
	"backend/global"
	"backend/model"
	"backend/pkg/mq/tasks"
	"backend/pkg/utils"
	"context"
	"errors"
	"strconv"
	"strings"

	"go.uber.org/zap"
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

// GetPathNodes 根据 nodeId 查询路径，并返回路径各层级的同级节点列表
func (s *KnowledgeNodeService) GetPathNodes(ctx context.Context, nodeID int, userID uint) ([]dto.KnowledgeNodeSimpleRes, error) {
	// 1. 获取当前节点获取路径
	node, err := s.nodeDao.GetNodeByID(nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("知识点不存在")
		}
		return nil, err
	}

	// 2. 解析路径，例如 "0/127/128/129/130/"
	// path 包含自身ID，例如：0/127/128/129/130/
	trimmedPath := strings.Trim(node.Path, "/")
	if trimmedPath == "" {
		return []dto.KnowledgeNodeSimpleRes{}, nil
	}
	pathParts := strings.Split(trimmedPath, "/")

	start := 0
	if len(pathParts) > 0 && pathParts[0] == "0" {
		start = 1
	}
	if len(pathParts) <= start {
		return []dto.KnowledgeNodeSimpleRes{}, nil
	}

	pathNodeIDs := pathParts[start:]
	var parentIDs []int
	seen := make(map[int]struct{})
	parentIDs = append(parentIDs, 0)
	seen[0] = struct{}{}
	for i := 0; i < len(pathNodeIDs)-1; i++ {
		idStr := pathNodeIDs[i]
		id, _ := strconv.Atoi(idStr)
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		parentIDs = append(parentIDs, id)
	}

	// 3. 批量查询这些 parentId 下的所有子节点
	nodes, err := s.nodeDao.GetNodesByParentIDs(parentIDs)
	if err != nil {
		return nil, err
	}

	// 按路径层级顺序重新排序，避免数据库按 parent_id 数值排序打乱层级展示。
	nodesByParentID := make(map[int][]model.KnowledgeNode)
	for _, item := range nodes {
		nodesByParentID[item.ParentID] = append(nodesByParentID[item.ParentID], item)
	}
	var orderedNodes []model.KnowledgeNode
	for _, parentID := range parentIDs {
		orderedNodes = append(orderedNodes, nodesByParentID[parentID]...)
	}

	// 4. 获取用户在这些节点上的学习进度
	var nodeIDs []uint
	for _, n := range orderedNodes {
		nodeIDs = append(nodeIDs, n.ID)
	}

	var statusMap = make(map[uint]string)
	if userID > 0 && len(nodeIDs) > 0 {
		statuses, err := s.nodeDao.GetUserStudyStatusByNodeIDs(userID, nodeIDs)
		if err != nil {
			return nil, err
		}
		for _, st := range statuses {
			statusMap[uint(st.NodeID)] = st.Status
		}
	}

	// 5. 组装返回数据
	var result []dto.KnowledgeNodeSimpleRes
	for _, n := range orderedNodes {
		// 默认进度为 unstarted
		progress := "unstarted"
		if st, ok := statusMap[n.ID]; ok {
			progress = st
		}

		result = append(result, dto.KnowledgeNodeSimpleRes{
			ID:                 n.ID,
			ParentID:           n.ParentID,
			Name:               n.Name,
			Path:               n.Path,
			SortOrder:          n.SortOrder,
			ImageUrl:           n.ImageUrl,
			IsLeaf:             n.IsLeaf,
			UserProgressStatus: progress,
		})
	}

	return result, nil
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
	res := &dto.KnowledgeNodeDetailRes{
		KnowledgeNodeItemRes: baseInfo,
		Content:              contentStr,
	}

	// 5. 异步保存教材学习进度 (如果用户已登录)
	if userID > 0 {
		if err := global.GVA_MQ.Publish(ctx, "subject_progress", tasks.SubjectProgressPayload{
			UserID:    userID,
			SubjectID: node.SubjectID,
			NodeID:    nodeID,
		}); err != nil {
			global.GVA_LOG.Error("发布教材进度更新消息失败", zap.Error(err))
		}
	}

	return res, nil
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

// UpsertUserStudyNote 创建或更新用户随堂笔记
func (s *KnowledgeNodeService) UpsertUserStudyNote(ctx context.Context, userID uint, nodeID int, req dto.UpsertUserStudyNoteReq) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}

	// 1. 校验内容长度（允许空字符串）
	noteContent := strings.TrimSpace(req.NoteContent)
	if len([]rune(noteContent)) > 1000 {
		return errors.New("笔记内容不能超过 1000 个字符")
	}

	// 2. 防止 XSS 攻击
	safeContent := utils.XSSFilter(noteContent)

	// 4. 检查知识点是否存在
	_, err := s.nodeDao.GetNodeByID(nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("知识点不存在")
		}
		return err
	}

	err = s.nodeDao.UpsertUserStudyNote(userID, nodeID, safeContent, req.IsImportant)
	if err != nil {
		return err
	}

	// 5. 异步发布用户活跃度更新消息 (笔记创建/修改，5分)
	if err := global.GVA_MQ.Publish(ctx, "user_activity", tasks.UserActivityPayload{
		UserID:     userID,
		ActionType: "create_note",
		TargetType: "knowledge_nodes",
		TargetID:   nodeID,
		Score:      5,
	}); err != nil {
		global.GVA_LOG.Error("发布用户活跃度消息失败", zap.Error(err))
	}

	return nil
}

// UpdateNodeStatus 更新用户对知识点的学习状态
func (s *KnowledgeNodeService) UpdateNodeStatus(ctx context.Context, userID uint, nodeID int, status string) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}
	err := s.nodeDao.UpsertUserStudyStatus(userID, nodeID, status)
	if err != nil {
		return err
	}

	// 如果状态是已完成，异步发送活跃度更新消息
	if status == "completed" {
		if err := global.GVA_MQ.Publish(ctx, "user_activity", tasks.UserActivityPayload{
			UserID:     userID,
			ActionType: "study_note",
			TargetType: "knowledge_nodes",
			TargetID:   nodeID,
			Score:      2,
		}); err != nil {
			global.GVA_LOG.Error("发布用户活跃度更新消息失败", zap.Error(err))
		}
	}

	return nil
}

// MarkNodeDifficulty 标记或更新知识点的难度评价
func (s *KnowledgeNodeService) MarkNodeDifficulty(ctx context.Context, userID uint, nodeID int, difficulty string) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}

	// 开启事务
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查询该用户是否已对该节点标记过难度
		oldDiff, err := s.nodeDao.GetUserNodeDifficulty(tx, userID, nodeID)
		isUpdate := true
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				isUpdate = false
			} else {
				return err
			}
		}

		// 2. 如果已存在且难度评价相同，直接返回（幂等处理）
		if isUpdate && oldDiff.Difficulty == difficulty {
			return nil
		}

		// 3. 更新/创建个人评价记录
		err = s.nodeDao.UpsertUserNodeDifficulty(tx, userID, nodeID, difficulty)
		if err != nil {
			return err
		}

		// 4. 更新聚合统计表 (node_metrics)
		if isUpdate {
			// 如果是更新，旧的评价数 -1，新的评价数 +1
			err = s.nodeDao.UpdateNodeMetric(tx, nodeID, oldDiff.Difficulty, -1)
			if err != nil {
				return err
			}
		}
		// 新的评价数 +1
		err = s.nodeDao.UpdateNodeMetric(tx, nodeID, difficulty, 1)
		if err != nil {
			return err
		}

		return nil
	})
}

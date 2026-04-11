package tasks

import (
	"backend/dao"
	"backend/global"
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

// SubjectProgressPayload 教材学习进度更新负载
type SubjectProgressPayload struct {
	UserID    uint `json:"userId"`
	SubjectID int  `json:"subjectId"`
	NodeID    int  `json:"nodeId"`
}

// HandleSubjectProgress 处理教材进度更新任务
func HandleSubjectProgress(ctx context.Context, payloadStr string) error {
	var payload struct {
		Payload SubjectProgressPayload `json:"payload"`
	}

	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		global.GVA_LOG.Error("解析进度更新消息失败", zap.Error(err))
		return err
	}

	p := payload.Payload
	if p.UserID == 0 || p.SubjectID == 0 || p.NodeID == 0 {
		return nil // 忽略非法负载
	}

	// 使用 DAO 更新进度
	var subjectDao dao.SubjectDao
	return subjectDao.UpsertUserSubjectProgress(ctx, p.UserID, p.SubjectID, p.NodeID)
}

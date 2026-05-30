package service

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func (s *AdminService) PendingNodes() (dto.ListResponse[dto.NodeReviewItem], error) {
	list, err := s.adminDao.ListPendingNodes()
	if err != nil {
		return dto.ListResponse[dto.NodeReviewItem]{}, errors.New("加载待审批节点失败")
	}
	if list == nil {
		list = []dto.NodeReviewItem{}
	}
	return dto.ListResponse[dto.NodeReviewItem]{List: list, Total: len(list)}, nil
}

func (s *AdminService) ApproveNode(adminID uint, nodeID int, remark string) error {
	return s.reviewNode(adminID, nodeID, "approve", strings.TrimSpace(remark))
}

func (s *AdminService) RejectNode(adminID uint, nodeID int, remark string) error {
	remark = strings.TrimSpace(remark)
	if remark == "" {
		return errors.New("驳回时需要填写审批意见")
	}
	return s.reviewNode(adminID, nodeID, "reject", remark)
}

func (s *AdminService) reviewNode(adminID uint, nodeID int, action string, remark string) error {
	if action != "approve" && action != "reject" {
		return errors.New("审批动作不正确")
	}

	return s.adminDao.Transaction(func(tx *gorm.DB) error {
		node, err := s.adminDao.FindNodeForUpdate(tx, nodeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("节点不存在")
			}
			return err
		}
		if node.AuditStatus != 1 {
			return errors.New("该节点当前不在待审批状态")
		}

		subject, err := s.adminDao.FindSubjectByID(tx, node.SubjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("节点所属教材不存在")
			}
			return err
		}
		if subject.Status != "published" {
			return errors.New("教材通过审核后才能审批节点")
		}

		content, contentErr := s.adminDao.FindContentByNodeID(tx, node.ID)
		if contentErr != nil && !errors.Is(contentErr, gorm.ErrRecordNotFound) {
			return contentErr
		}

		snapshot, err := json.Marshal(map[string]any{
			"node":    node,
			"content": content,
		})
		if err != nil {
			return err
		}

		log := model.AuditLog{
			TargetType:    "node",
			TargetID:      int(node.ID),
			AdminID:       int(adminID),
			Action:        action,
			Remark:        remark,
			DraftSnapshot: string(snapshot),
		}
		if err := s.adminDao.CreateAuditLog(tx, &log); err != nil {
			return err
		}

		nodeUpdates := map[string]any{
			"audit_status": statusForAction(action),
			"last_log_id":  log.ID,
		}
		if action == "approve" {
			nodeUpdates["status"] = "published"
			nodeUpdates["has_draft"] = 0
			if node.NameDraft != "" {
				nodeUpdates["name"] = node.NameDraft
			}
		}
		if err := s.adminDao.UpdateNode(tx, node.ID, nodeUpdates); err != nil {
			return err
		}

		if contentErr == nil {
			contentUpdates := map[string]any{
				"audit_status": statusForAction(action),
				"last_log_id":  log.ID,
			}
			if action == "approve" {
				contentUpdates["content"] = content.ContentDraft
				contentUpdates["has_draft"] = 0
			}
			if err := s.adminDao.UpdateContent(tx, content.ID, contentUpdates); err != nil {
				return err
			}
		}

		if action == "approve" && node.HasDraft == 1 {
			ancestorIDs := ancestorNodeIDsFromPath(node.Path)
			if len(ancestorIDs) > 0 {
				return s.adminDao.DecrementAncestorDraftCount(tx, node.SubjectID, ancestorIDs)
			}
		}

		return nil
	})
}

func statusForAction(action string) int8 {
	if action == "approve" {
		return 2
	}
	return 3
}

func ancestorNodeIDsFromPath(path string) []int {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.Atoi(part)
		if err != nil || id <= 0 {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

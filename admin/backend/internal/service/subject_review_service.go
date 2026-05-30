package service

import (
	"admin/backend/internal/dto"
	"admin/backend/internal/model"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

func (s *AdminService) PendingSubjects() (dto.ListResponse[dto.SubjectReviewItem], error) {
	list, err := s.adminDao.ListPendingSubjects()
	if err != nil {
		return dto.ListResponse[dto.SubjectReviewItem]{}, errors.New("加载待审批教材失败")
	}
	if list == nil {
		list = []dto.SubjectReviewItem{}
	}
	return dto.ListResponse[dto.SubjectReviewItem]{List: list, Total: len(list)}, nil
}

func (s *AdminService) ApproveSubject(adminID uint, subjectID int, remark string) error {
	return s.reviewSubject(adminID, subjectID, "approve", strings.TrimSpace(remark))
}

func (s *AdminService) RejectSubject(adminID uint, subjectID int, remark string) error {
	remark = strings.TrimSpace(remark)
	if remark == "" {
		return errors.New("驳回时需要填写审批意见")
	}
	return s.reviewSubject(adminID, subjectID, "reject", remark)
}

func (s *AdminService) reviewSubject(adminID uint, subjectID int, action string, remark string) error {
	if action != "approve" && action != "reject" {
		return errors.New("审批动作不正确")
	}

	return s.adminDao.Transaction(func(tx *gorm.DB) error {
		subject, err := s.adminDao.FindSubjectForUpdate(tx, subjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("教材不存在")
			}
			return err
		}
		if subject.AuditStatus != 1 {
			return errors.New("该教材当前不在待审批状态")
		}

		snapshot, err := json.Marshal(subject)
		if err != nil {
			return err
		}

		log := model.AuditLog{
			TargetType:    "subject",
			TargetID:      int(subject.ID),
			AdminID:       int(adminID),
			Action:        action,
			Remark:        remark,
			DraftSnapshot: string(snapshot),
		}
		if err := s.adminDao.CreateAuditLog(tx, &log); err != nil {
			return err
		}

		updates := map[string]any{"last_log_id": log.ID}
		if action == "approve" {
			updates["audit_status"] = 2
			updates["status"] = "published"
			updates["has_draft"] = 0
			if subject.NameDraft != "" {
				updates["name"] = subject.NameDraft
			}
			if subject.IconDraft != "" {
				updates["icon"] = subject.IconDraft
			}
			if subject.DescriptionDraft != "" {
				updates["description"] = subject.DescriptionDraft
			}
			if subject.CoverImageIDDraft != 0 {
				updates["cover_image_id"] = subject.CoverImageIDDraft
			}
		} else {
			updates["audit_status"] = 3
			updates["has_draft"] = 1
		}

		if err := s.adminDao.UpdateSubject(tx, subject.ID, updates); err != nil {
			return err
		}

		if action != "approve" {
			return nil
		}

		if err := s.adminDao.PublishNodesBySubjectID(tx, subject.ID, log.ID); err != nil {
			return err
		}
		if err := s.adminDao.PublishContentsBySubjectID(tx, subject.ID, log.ID); err != nil {
			return err
		}

		if subject.NameDraft != "" {
			return s.adminDao.UpdateTopNodeBySubjectID(tx, subject.ID, map[string]any{
				"name":       subject.NameDraft,
				"name_draft": subject.NameDraft,
			})
		}
		return nil
	})
}

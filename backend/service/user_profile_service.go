package service

import (
	"backend/dao"
	"backend/dto"
	"backend/pkg/errmsg"
	"context"
	"time"
)

// GetUserActivitiesCalendar 获取用户一年的活跃度日历
func (u *UserService) GetUserActivitiesCalendar(ctx context.Context, userID uint) (int, dto.UserActivityCalendarRes) {
	now := time.Now()
	startTime := now.AddDate(-1, 0, 0) // 一年前

	activities, err := dao.GetUserActivities(ctx, userID, startTime, now)
	if err != nil {
		return errmsg.CodeError, dto.UserActivityCalendarRes{}
	}

	var res dto.UserActivityCalendarRes
	for _, act := range activities {
		res.Activities = append(res.Activities, dto.UserActivityItem{
			Date:  act.ActivityDate.Format("2006-01-02"),
			Count: act.ActivityCount,
		})
	}
	return errmsg.CodeSuccess, res
}

// GetPublicPrivateNotes 获取公开的私人笔记列表
func (u *UserService) GetPublicPrivateNotes(ctx context.Context, userID uint, req dto.PublicPrivateNoteListReq) (int, dto.PublicPrivateNoteListRes) {
	offset := (req.Page - 1) * req.PageSize
	total, notes, err := dao.GetPublicPrivateNotes(ctx, userID, offset, req.PageSize)
	if err != nil {
		return errmsg.CodeError, dto.PublicPrivateNoteListRes{}
	}

	var res dto.PublicPrivateNoteListRes
	res.Total = total
	for _, note := range notes {
		res.List = append(res.List, dto.PublicPrivateNoteItem{
			ID:        note.ID,
			Title:     note.Title,
			UpdatedAt: note.UpdatedAt,
		})
	}
	return errmsg.CodeSuccess, res
}

// GetLearnedSubjects 获取已学/在学教材列表
func (u *UserService) GetLearnedSubjects(ctx context.Context, userID uint) (int, dto.LearnedSubjectListRes) {
	subjects, err := dao.GetLearnedSubjects(ctx, userID)
	if err != nil {
		return errmsg.CodeError, dto.LearnedSubjectListRes{}
	}

	var res dto.LearnedSubjectListRes
	for _, subject := range subjects {
		learned, _ := dao.GetLearnedNodeCountBySubject(ctx, userID, subject.ID)
		total, _ := dao.GetTotalNodeCountBySubject(ctx, subject.ID)

		res.List = append(res.List, dto.LearnedSubjectItem{
			SubjectID:   subject.ID,
			SubjectName: subject.Name,
			CoverImage:  subject.Icon, // 或关联 images 表取图片路径
			Learned:     learned,
			Total:       total,
		})
	}
	return errmsg.CodeSuccess, res
}

// GetSharedNotes 获取已分享的笔记列表
func (u *UserService) GetSharedNotes(ctx context.Context, userID uint, req dto.SharedNoteListReq) (int, dto.SharedNoteListRes) {
	offset := (req.Page - 1) * req.PageSize
	total, notes, err := dao.GetSharedNotes(ctx, userID, offset, req.PageSize)
	if err != nil {
		return errmsg.CodeError, dto.SharedNoteListRes{}
	}

	var res dto.SharedNoteListRes
	res.Total = total
	for _, note := range notes {
		res.List = append(res.List, dto.SharedNoteItem{
			ID:         note.ID,
			NodeID:     note.NodeID,
			NodeName:   note.NodeName,
			ShareToken: note.ShareToken,
			ViewCount:  note.ViewCount,
			CreatedAt:  note.CreatedAt,
			ExpiresAt:  note.ExpiresAt,
		})
	}
	return errmsg.CodeSuccess, res
}

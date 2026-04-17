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
			Score: act.ActivityScore,
		})
	}
	return errmsg.CodeSuccess, res
}

// UpdateSharedNoteStatus 更新分享状态
func (u *UserService) UpdateSharedNoteStatus(ctx context.Context, userID uint, shareID int, isActive int8) int {
	var privateNoteDao dao.UserPrivateNoteDao
	if err := privateNoteDao.UpdateNoteShareStatus(ctx, userID, shareID, isActive); err != nil {
		return errmsg.CodeError
	}
	return errmsg.CodeSuccess
}

// DeleteSharedNote 删除分享记录
func (u *UserService) DeleteSharedNote(ctx context.Context, userID uint, shareID int) int {
	var privateNoteDao dao.UserPrivateNoteDao
	if err := privateNoteDao.DeleteNoteShare(ctx, userID, shareID); err != nil {
		return errmsg.CodeError
	}
	return errmsg.CodeSuccess
}

// UpdateSharedNoteExpire 更新分享过期时间
func (u *UserService) UpdateSharedNoteExpire(ctx context.Context, userID uint, shareID int, req dto.UpdateSharedNoteExpireReq) int {
	var privateNoteDao dao.UserPrivateNoteDao
	var expireAt time.Time

	if req.ExpireMinutes > 0 {
		// 优先使用传递的分钟数，以当前时间为基准向后延长
		expireAt = time.Now().Add(time.Duration(req.ExpireMinutes) * time.Minute)
	} else if req.ExpireAt != "" {
		// 其次解析传递的具体时间
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExpireAt, time.Local)
		if err != nil {
			return errmsg.CodeError // 或者你可以定义一个更具体的错误码，如时间格式错误
		}
		if parsedTime.Before(time.Now()) {
			return errmsg.CodeError // 不能设置为过去的时间
		}
		expireAt = parsedTime
	} else {
		// 都没传，直接返回错误
		return errmsg.CodeError
	}

	if err := privateNoteDao.UpdateNoteShareExpire(ctx, userID, shareID, expireAt); err != nil {
		return errmsg.CodeError
	}

	return errmsg.CodeSuccess
}

// GetPublicPrivateNotes 获取公开的私人笔记列表
func (u *UserService) GetPublicPrivateNotes(ctx context.Context, userID uint, req dto.PublicPrivateNoteListReq) (int, dto.PublicPrivateNoteListRes) {
	offset := (req.Page - 1) * req.PageSize
	total, notes, err := dao.GetPublicPrivateNotes(ctx, userID, offset, req.PageSize)
	if err != nil {
		return errmsg.CodeError, dto.PublicPrivateNoteListRes{}
	}

	// 批量查询分享状态
	var noteIDs []uint
	for _, note := range notes {
		noteIDs = append(noteIDs, note.ID)
	}
	var privateNoteDao dao.UserPrivateNoteDao
	sharedStatusMap, _ := privateNoteDao.CheckNotesSharedStatus(ctx, userID, noteIDs)

	var res dto.PublicPrivateNoteListRes
	res.Total = total
	for _, note := range notes {
		res.List = append(res.List, dto.PublicPrivateNoteItem{
			ID:        note.ID,
			Title:     note.Title,
			UpdatedAt: note.UpdatedAt,
			IsShared:  sharedStatusMap[note.ID],
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
			ID:            note.ID,
			PrivateNoteID: note.PrivateNoteID,
			NoteTitle:     note.NoteTitle,
			NoteType:      note.NoteType,
			ShareToken:    note.ShareToken,
			ShareCode:     note.ShareCode,
			ViewCount:     note.ViewCount,
			IsActive:      note.IsActive,
			CreatedAt:     note.CreatedAt,
			ExpiresAt:     note.ExpiresAt,
		})
	}
	return errmsg.CodeSuccess, res
}

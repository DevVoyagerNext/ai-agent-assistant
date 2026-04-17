package dto

import "time"

// UserActivityItem 用户活跃度日历项
type UserActivityItem struct {
	Date  string `json:"date"`  // 格式: YYYY-MM-DD
	Count int    `json:"count"` // 当日活跃次数/分数
	Score int    `json:"score"`
}

// UserActivityCalendarRes 用户活跃度日历响应
type UserActivityCalendarRes struct {
	Activities []UserActivityItem `json:"activities"`
}

// PaginationReq 分用请求基础
type PaginationReq struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// PublicPrivateNoteListReq 公开私人笔记列表请求
type PublicPrivateNoteListReq struct {
	PaginationReq
}

// PublicPrivateNoteItem 公开私人笔记项
type PublicPrivateNoteItem struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsShared  bool      `json:"isShared"` // 是否正在被分享
}

// PublicPrivateNoteListRes 公开私人笔记列表响应
type PublicPrivateNoteListRes struct {
	Total int64                   `json:"total"`
	List  []PublicPrivateNoteItem `json:"list"`
}

// LearnedSubjectItem 已学教材项
type LearnedSubjectItem struct {
	SubjectID   uint   `json:"subjectId"`
	SubjectName string `json:"subjectName"`
	CoverImage  string `json:"coverImage"` // 可以结合 images 表或直接使用 subject.icon
	Learned     int64  `json:"learned"`    // 已学知识点数量
	Total       int64  `json:"total"`      // 总知识点数量
}

// LearnedSubjectListRes 已学教材列表响应
type LearnedSubjectListRes struct {
	List []LearnedSubjectItem `json:"list"`
}

// SharedNoteListReq 已分享笔记列表请求
type SharedNoteListReq struct {
	PaginationReq
}

// SharedNoteItem 已分享笔记项
type SharedNoteItem struct {
	ID            uint      `json:"id"`
	PrivateNoteID int       `json:"privateNoteId"`
	NoteTitle     string    `json:"noteTitle"`
	NoteType      string    `json:"noteType"`
	ShareToken    string    `json:"shareToken"`
	ShareCode     string    `json:"shareCode"`
	ViewCount     int       `json:"viewCount"`
	IsActive      int8      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

// SharedNoteListRes 已分享笔记列表响应
type SharedNoteListRes struct {
	Total int64            `json:"total"`
	List  []SharedNoteItem `json:"list"`
}

// UpdateSharedNoteStatusReq 更新分享状态请求
type UpdateSharedNoteStatusReq struct {
	IsActive int8 `json:"isActive" binding:"oneof=0 1"`
}

// UpdateSharedNoteExpireReq 更新分享过期时间请求
type UpdateSharedNoteExpireReq struct {
	ExpireMinutes int    `json:"expireMinutes"` // 延长多少分钟，如果传了优先使用这个
	ExpireAt      string `json:"expireAt"`      // 过期的具体时间，如 2006-01-02 15:04:05
}

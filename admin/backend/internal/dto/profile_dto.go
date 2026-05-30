package dto

import "time"

type AdminInfo struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Signature   string     `json:"signature"`
	AvatarUrl   string     `json:"avatarUrl"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type DashboardStats struct {
	Pending        int64 `json:"pending"`
	PendingSubject int64 `json:"pendingSubject"`
	PendingNode    int64 `json:"pendingNode"`
	Approved       int64 `json:"approved"`
	Rejected       int64 `json:"rejected"`
	TodayReview    int64 `json:"todayReview"`
}

type MeResponse struct {
	Admin AdminInfo      `json:"admin"`
	Stats DashboardStats `json:"stats"`
}

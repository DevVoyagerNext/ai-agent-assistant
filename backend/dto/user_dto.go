package dto

// SendEmailReq 发送邮件验证码请求
type SendEmailReq struct {
	Email string `json:"email" binding:"required"`
}

// RegisterReq 用户注册请求
type RegisterReq struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Signature string `json:"signature"`
}

// RegisterRes 注册成功响应数据
type RegisterRes struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken"`
	ExpiresAt    int64       `json:"expiresAt"`
	User         UserInfoRes `json:"user"`
}

// UserInfoRes 用户信息响应数据
type UserInfoRes struct {
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarUrl"`
	Signature string `json:"signature"`

	// 核心统计数据
	FollowersCount       int64 `json:"followersCount"`       // 被关注数量 (粉丝数)
	FollowingCount       int64 `json:"followingCount"`       // 关注数量
	LearnedSubjectsCount int64 `json:"learnedSubjectsCount"` // 已学/在学教材总数
	SharedNotesCount     int64 `json:"sharedNotesCount"`     // 分享笔记总数
}

// LoginReq 用户登录请求
type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRes 用户登录响应数据
type LoginRes struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken"`
	ExpiresAt    int64       `json:"expiresAt"`
	User         UserInfoRes `json:"user"`
}

// RefreshTokenReq 刷新 Token 请求
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshTokenRes 刷新 Token 响应
type RefreshTokenRes struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

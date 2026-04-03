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
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarUrl"`
	Signature string `json:"signature"`
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

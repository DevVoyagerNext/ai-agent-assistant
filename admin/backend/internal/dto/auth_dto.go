package dto

type RegisterRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	InviteCode string `json:"inviteCode"`
	Signature  string `json:"signature"`
}

type LoginRequest struct {
	Account  string `json:"account"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Session   string            `json:"session"`
	SessionID string            `json:"sessionId"`
	ExpiresAt int64             `json:"expiresAt"`
	Admin     AdminInfo         `json:"admin"`
	Device    SessionDeviceInfo `json:"device"`
}

type SessionDeviceInfo struct {
	DeviceID  string `json:"deviceId"`
	UserAgent string `json:"userAgent"`
	IP        string `json:"ip"`
	LoginAt   int64  `json:"loginAt"`
}

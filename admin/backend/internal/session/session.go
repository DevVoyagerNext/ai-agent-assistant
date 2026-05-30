package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	HeaderSessionID = "x-session-id"
	HeaderDeviceID  = "x-device-id"
)

type DeviceInfo struct {
	DeviceID  string `json:"deviceId"`
	UserAgent string `json:"userAgent"`
	IP        string `json:"ip"`
	LoginAt   int64  `json:"loginAt"`
}

type Data struct {
	SessionID string     `json:"sessionId"`
	UserID    uint       `json:"userId"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Status    int8       `json:"status"`
	Device    DeviceInfo `json:"device"`
	CreatedAt int64      `json:"createdAt"`
	ExpiresAt int64      `json:"expiresAt"`
}

func GenerateID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func DeviceFromRequest(c *gin.Context) DeviceInfo {
	deviceID := strings.TrimSpace(c.GetHeader(HeaderDeviceID))
	if deviceID == "" {
		deviceID = strings.TrimSpace(c.GetHeader("x-admin-device-id"))
	}
	return DeviceInfo{
		DeviceID:  deviceID,
		UserAgent: c.GetHeader("User-Agent"),
		IP:        c.ClientIP(),
		LoginAt:   time.Now().Unix(),
	}
}

func ReadID(c *gin.Context) string {
	if value := strings.TrimSpace(c.GetHeader(HeaderSessionID)); value != "" {
		return value
	}
	if value := strings.TrimSpace(c.GetHeader("x-admin-session")); value != "" {
		return value
	}
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(auth, "Session ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Session "))
	}
	return strings.TrimSpace(c.GetHeader("x-token"))
}

func Key(sessionID string) string {
	return "admin:session:" + sessionID
}

func UserKey(userID uint) string {
	return "admin:user-session:" + formatUint(userID)
}

func Save(ctx context.Context, client *redis.Client, data Data, ttl time.Duration) error {
	if client == nil {
		return errors.New("redis client is nil")
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	userKey := UserKey(data.UserID)
	oldSessionID, _ := client.Get(ctx, userKey).Result()

	pipe := client.TxPipeline()
	if oldSessionID != "" && oldSessionID != data.SessionID {
		pipe.Del(ctx, Key(oldSessionID))
	}
	pipe.Set(ctx, Key(data.SessionID), raw, ttl)
	pipe.Set(ctx, userKey, data.SessionID, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func Load(ctx context.Context, client *redis.Client, sessionID string) (Data, error) {
	var data Data
	if strings.TrimSpace(sessionID) == "" {
		return data, redis.Nil
	}
	raw, err := client.Get(ctx, Key(sessionID)).Result()
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return data, err
	}
	if data.SessionID == "" {
		data.SessionID = sessionID
	}
	return data, nil
}

func Validate(ctx context.Context, client *redis.Client, sessionID string, ttl time.Duration) (Data, error) {
	data, err := Load(ctx, client, sessionID)
	if err != nil {
		return Data{}, err
	}
	currentSessionID, err := client.Get(ctx, UserKey(data.UserID)).Result()
	if err != nil {
		return Data{}, err
	}
	if currentSessionID != sessionID {
		return Data{}, redis.Nil
	}
	pipe := client.TxPipeline()
	pipe.Expire(ctx, Key(sessionID), ttl)
	pipe.Expire(ctx, UserKey(data.UserID), ttl)
	_, _ = pipe.Exec(ctx)
	return data, nil
}

func formatUint(value uint) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

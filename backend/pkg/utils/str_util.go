package utils

import (
	"strings"
)

// DesensitizeEmail 邮箱脱敏处理
// 例如: 1503464068@qq.com -> 150******8@qq.com
func DesensitizeEmail(email string) string {
	if email == "" {
		return ""
	}
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 {
		return email
	}
	username := email[:atIndex]
	domain := email[atIndex:]

	if len(username) <= 2 {
		return "***" + domain
	}

	return username[:3] + "******" + username[len(username)-1:] + domain
}

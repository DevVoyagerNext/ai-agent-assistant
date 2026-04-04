package utils

import (
	"backend/config"
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

// Email 邮件发送结构体
type Email struct {
	To      []string // 收件人列表
	Subject string   // 邮件主题
	Body    string   // 邮件内容
}

// SendEmail 发送邮件
// to: 收件人列表
// subject: 邮件主题
// body: 邮件内容
func SendEmail(cfg config.Email, to []string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(cfg.From, cfg.Nickname))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Secret)
	if cfg.IsSSL {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return d.DialAndSend(m)
}

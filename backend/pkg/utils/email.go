package utils

import (
	"backend/global"
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
func SendEmail(to []string, subject string, body string) error {
	e := global.GVA_CONFIG.Email
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.From, e.Nickname))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.Host, e.Port, e.From, e.Secret)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}

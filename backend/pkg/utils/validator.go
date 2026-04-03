package utils

import (
	"regexp"
)

// IsQQEmail 校验是否为 QQ 邮箱格式
func IsQQEmail(email string) bool {
	// QQ 邮箱正则：数字(5-11位)@qq.com
	pattern := `^[1-9]\d{4,10}@qq\.com$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// IsUsername 校验用户名：不能是空，不能超过10个，支持中英文下划线
func IsUsername(username string) bool {
	if username == "" || len([]rune(username)) > 10 {
		return false
	}
	// 中英文下划线
	pattern := `^[\p{Han}\w]+$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(username)
}

// IsPassword 校验密码：8到20位，英文字母，特殊符号，数字，至少包含其中两种
func IsPassword(password string) bool {
	if len(password) < 8 || len(password) > 20 {
		return false
	}
	// 校验支持的字符集：英文字母、数字、常见特殊字符
	match, _ := regexp.MatchString(`^[a-zA-Z0-9!@#$%^&*()_+=-]+$`, password)
	if !match {
		return false
	}

	// 统计包含的类型
	types := 0
	if match, _ := regexp.MatchString(`[a-zA-Z]`, password); match {
		types++
	}
	if match, _ := regexp.MatchString(`[0-9]`, password); match {
		types++
	}
	if match, _ := regexp.MatchString(`[!@#$%^&*()_+=-]`, password); match {
		types++
	}

	return types >= 2
}

// IsSignature 校验个性签名：不能超过30个字，支持中文，英文，数字，常用写作符号
func IsSignature(signature string) bool {
	if len([]rune(signature)) > 30 {
		return false
	}
	// 中文、英文、数字、常用标点符号 (,.?;:!，。？；：！)
	pattern := `^[\p{Han}\w\s\.,\?;:!，。？；：！]*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(signature)
}

// IsCode 校验验证码格式：4位数字或英文字母
func IsCode(code string) bool {
	pattern := `^[a-zA-Z0-9]{4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(code)
}

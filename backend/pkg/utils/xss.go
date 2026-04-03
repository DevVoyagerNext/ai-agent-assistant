package utils

import (
	"html"
)

// XSSFilter 简单的 XSS 过滤
func XSSFilter(s string) string {
	return html.EscapeString(s)
}

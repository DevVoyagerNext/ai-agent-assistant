package utils

import (
	"backend/global"
	"bytes"
	"context"
	"fmt"
	neturl "net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// UploadToQiniu 上传文件到七牛云，返回 fileKey
func UploadToQiniu(fileBytes []byte, originalName string, prefix string) (string, error) {
	q := global.GVA_CONFIG.Qiniu
	accessKey := CleanQiniuFileURL(q.AccessKey)
	secretKey := CleanQiniuFileURL(q.SecretKey)
	bucket := CleanQiniuFileURL(q.Bucket)
	zone := strings.ToLower(CleanQiniuFileURL(q.Zone))
	if accessKey == "" || secretKey == "" || bucket == "" {
		return "", fmt.Errorf("七牛云配置缺失")
	}

	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	// 华南-广东 z2
	if zone == "z2" {
		cfg.Region = &storage.ZoneHuanan
	} else {
		// 默认自动识别
		cfg.Region = &storage.ZoneHuanan
	}
	cfg.UseHTTPS = q.UseHTTPS
	cfg.UseCdnDomains = false

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".file"
	}

	// Agent/前缀或者其他前缀
	fileKey := fmt.Sprintf("%s%d%s", prefix, time.Now().UnixNano(), ext)

	err := formUploader.Put(context.Background(), &ret, upToken, fileKey, bytes.NewReader(fileBytes), int64(len(fileBytes)), nil)
	if err != nil {
		return "", err
	}

	return ret.Key, nil
}

// GetQiniuDownloadURL 获取七牛云文件的下载链接
func GetQiniuDownloadURL(fileKey string) string {
	q := global.GVA_CONFIG.Qiniu
	if CleanQiniuFileURL(q.Domain) == "" {
		return ""
	}

	cleanKey := CleanQiniuFileURL(fileKey)
	if cleanKey == "" {
		return ""
	}

	domain := buildQiniuBaseDomain(q.Domain)
	if domain == "" {
		return ""
	}

	mac := auth.New(CleanQiniuFileURL(q.AccessKey), CleanQiniuFileURL(q.SecretKey))
	deadline := time.Now().Add(2 * time.Hour).Unix()
	return storage.MakePrivateURL(mac, domain, cleanKey, deadline)
}

// ExtractQiniuKey 从完整的七牛云下载链接中提取 fileKey
func ExtractQiniuKey(fileURL string) string {
	q := global.GVA_CONFIG.Qiniu
	cleanURL := CleanQiniuFileURL(fileURL)
	if cleanURL == "" {
		return ""
	}
	if CleanQiniuFileURL(q.Domain) == "" {
		return cleanURL
	}

	domain := buildQiniuBaseDomain(q.Domain)
	if domain == "" {
		return cleanURL
	}

	if !strings.HasSuffix(domain, "/") {
		domain += "/"
	}

	if strings.HasPrefix(cleanURL, domain) {
		return strings.TrimPrefix(cleanURL, domain)
	}

	if parsed, err := neturl.Parse(cleanURL); err == nil && parsed.Host != "" {
		return strings.TrimPrefix(parsed.Path, "/")
	}

	return cleanURL
}

// CleanQiniuFileURL 清洗七牛云文件 URL / Key，去掉空格、反引号和多余包装
func CleanQiniuFileURL(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.Trim(cleaned, "`")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.Trim(cleaned, "\"'")
	return strings.TrimSpace(cleaned)
}

func buildQiniuBaseDomain(domain string) string {
	cleaned := CleanQiniuFileURL(domain)
	if cleaned == "" {
		return ""
	}
	if !strings.HasPrefix(cleaned, "http://") && !strings.HasPrefix(cleaned, "https://") {
		if global.GVA_CONFIG.Qiniu.UseHTTPS {
			cleaned = "https://" + cleaned
		} else {
			cleaned = "http://" + cleaned
		}
	}
	return strings.TrimRight(cleaned, "/")
}

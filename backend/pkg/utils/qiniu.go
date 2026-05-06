package utils

import (
	"backend/global"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// UploadToQiniu 上传文件到七牛云，返回 fileKey
func UploadToQiniu(fileBytes []byte, originalName string, prefix string) (string, error) {
	q := global.GVA_CONFIG.Qiniu
	if q.AccessKey == "" || q.SecretKey == "" {
		return "", fmt.Errorf("七牛云配置缺失")
	}

	putPolicy := storage.PutPolicy{
		Scope: q.Bucket,
	}
	mac := qbox.NewMac(q.AccessKey, q.SecretKey)
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	// 华南-广东 z2
	if q.Zone == "z2" {
		cfg.Region = &storage.ZoneHuanan
	} else {
		// 默认自动识别
		cfg.Region = &storage.ZoneHuanan
	}
	cfg.UseHTTPS = true
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
	if q.Domain == "" {
		return ""
	}

	domain := q.Domain
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "http://" + domain // 默认使用 http 或 https，根据实际情况
	}

	return fmt.Sprintf("%s/%s", domain, fileKey)
}

// ExtractQiniuKey 从完整的七牛云下载链接中提取 fileKey
func ExtractQiniuKey(fileURL string) string {
	q := global.GVA_CONFIG.Qiniu
	if q.Domain == "" {
		return fileURL
	}

	domain := q.Domain
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "http://" + domain
	}

	if !strings.HasSuffix(domain, "/") {
		domain += "/"
	}

	if strings.HasPrefix(fileURL, domain) {
		return strings.TrimPrefix(fileURL, domain)
	}

	return fileURL
}

package middleware

import (
	"backend/global"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const requestIDKey = "requestId"

// bodyLogWriter 用于劫持响应体内容
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = c.GetHeader("x-request-id")
		}
		if requestID == "" {
			buf := make([]byte, 16)
			if _, err := rand.Read(buf); err == nil {
				requestID = hex.EncodeToString(buf)
			}
		}
		if requestID != "" {
			c.Set(requestIDKey, requestID)
			c.Header("X-Request-Id", requestID)
		}
		c.Next()
	}
}

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. 记录并还原请求体 (增加对二进制文件上传的过滤)
		var requestBody []byte
		contentType := c.Request.Header.Get("Content-Type")
		isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

		if c.Request.Body != nil && !isMultipart {
			// 限制读取长度，防止超大请求导致内存溢出
			// 仅在非文件上传请求时读取 Body
			requestBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, 1024*10)) // 最多读取 10KB
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 2. 劫持响应体
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		logger := global.GVA_LOG

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		requestID, _ := c.Get(requestIDKey)
		requestIDStr, _ := requestID.(string)

		errorMsg := strings.TrimSpace(c.Errors.String())
		status := c.Writer.Status()

		// 构建请求体日志内容
		var requestBodyStr string
		if isMultipart {
			requestBodyStr = "[Multipart Form Data - Binary Content Omitted]"
		} else if len(requestBody) > 0 {
			requestBodyStr = string(requestBody)
			if len(requestBodyStr) > 2000 {
				requestBodyStr = requestBodyStr[:2000] + "...(truncated)"
			}
		}

		// 构建响应体日志内容 (同样增加截断防止日志过大)
		responseBodyStr := blw.body.String()
		if len(responseBodyStr) > 2000 {
			responseBodyStr = responseBodyStr[:2000] + "...(truncated)"
		}

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("ua", c.Request.UserAgent()),
			zap.String("request_id", requestIDStr),
			zap.Duration("latency", time.Since(start)),
			zap.Int("bytes", c.Writer.Size()),
			zap.String("request_body", requestBodyStr),
			zap.String("response_body", responseBodyStr),
		}

		if errorMsg != "" {
			fields = append(fields, zap.String("error", errorMsg))
		}

		switch {
		case status >= http.StatusInternalServerError || errorMsg != "":
			logger.Error("HTTP 请求", fields...)
		case status >= http.StatusBadRequest:
			logger.Warn("HTTP 请求", fields...)
		default:
			logger.Info("HTTP 请求", fields...)
		}
	}
}

func ZapRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger := global.GVA_LOG

				requestID, _ := c.Get(requestIDKey)
				requestIDStr, _ := requestID.(string)

				path := c.FullPath()
				if path == "" {
					path = c.Request.URL.Path
				}

				logger.Error("HTTP panic",
					zap.Any("recover", rec),
					zap.ByteString("stack", debug.Stack()),
					zap.String("method", c.Request.Method),
					zap.String("path", path),
					zap.String("query", c.Request.URL.RawQuery),
					zap.String("ip", c.ClientIP()),
					zap.String("ua", c.Request.UserAgent()),
					zap.String("request_id", requestIDStr),
				)

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

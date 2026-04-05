package middleware

import (
	"backend/global"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const requestIDKey = "requestId"

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

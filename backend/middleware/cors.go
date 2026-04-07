package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
		"http://localhost:8080": {},
		"http://127.0.0.1:8080": {},
		"http://localhost:3000": {},
		"http://127.0.0.1:3000": {},
		"http://localhost:4173": {},
		"http://127.0.0.1:4173": {},
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := false
		if origin != "" {
			if gin.Mode() != gin.ReleaseMode {
				allowOrigin = true
			} else {
				_, allowOrigin = allowedOrigins[origin]
			}
		}

		if allowOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With,X-Token,x-token,X-Request-Id,x-request-id")
			c.Header("Access-Control-Expose-Headers", "Content-Length,Content-Type,New-Token,New-Expires-At,X-Request-Id")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

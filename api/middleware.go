package api

import (
	"fmt"
	"github.com/trae2api/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trae2api/config"
)

// AuthMiddleware 验证请求的Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未设置 AuthToken，则不启用鉴权
		if config.AppConfig.AuthToken == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Log.Error("Authorization is empty")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 支持 "Bearer <token>" 格式
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		if token != config.AppConfig.AuthToken {
			logger.Log.Error(fmt.Sprintf("Invalid authorization token:%s", token))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

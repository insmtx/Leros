package mcp

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	apiKeyHeader        = "X-API-Key"
	bearerPrefix        = "Bearer "

	// TODO: 将此固定开发令牌替换为基于数据库的令牌验证。
	defaultAuthToken = "leros-mcp-token"
)

// DefaultAuthToken 返回当前 MCP 授权令牌。
func DefaultAuthToken() string {
	return defaultAuthToken
}

func requireToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		if validateToken(tokenFromRequest(c)) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}

func validateToken(token string) bool {
	// TODO: 将此固定开发令牌替换为基于数据库的令牌验证。
	return tokenMatches(token, defaultAuthToken)
}

func tokenFromRequest(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader(authorizationHeader))
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	}
	if authHeader != "" {
		return authHeader
	}
	return strings.TrimSpace(c.GetHeader(apiKeyHeader))
}

func tokenMatches(actual string, expected string) bool {
	if actual == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

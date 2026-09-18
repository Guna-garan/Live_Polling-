package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func extractToken(c *gin.Context, cookieName string) string {
	token, err := c.Cookie(cookieName)
	if err == nil && token != "" {
		return token
	}
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

// RequireAuth reads the JWT from HttpOnly cookie or Authorization Bearer header.
func RequireAuth(parse func(token string) (interface{}, error), cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c, cookieName)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required"},
			})
			return
		}

		userID, err := parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid or expired session"},
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuth behaves like RequireAuth but never aborts the request.
func OptionalAuth(parse func(token string) (interface{}, error), cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c, cookieName)
		if token != "" {
			if userID, err := parse(token); err == nil {
				c.Set("userID", userID)
			}
		}
		c.Next()
	}
}

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAuth reads the JWT from the HttpOnly cookie, validates it, and
// sets "userID" in the Gin context for downstream handlers. It never
// trusts any user/creator ID supplied by the client in the request body.
func RequireAuth(parse func(token string) (interface{}, error), cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err != nil || token == "" {
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

// OptionalAuth behaves like RequireAuth but never aborts the request; if
// a valid cookie is present it sets "userID", otherwise the request
// proceeds unauthenticated. Used on public endpoints (like GET poll)
// that behave slightly differently for the poll's own creator.
func OptionalAuth(parse func(token string) (interface{}, error), cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err == nil && token != "" {
			if userID, err := parse(token); err == nil {
				c.Set("userID", userID)
			}
		}
		c.Next()
	}
}

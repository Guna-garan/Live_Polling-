package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS restricts cross-origin requests to exactly the configured
// frontend origin. Credentials (the auth cookie) are allowed only for
// that origin — never with a wildcard, since browsers refuse
// Access-Control-Allow-Origin: * combined with credentials anyway, and
// allowing it would be a real vulnerability if that ever changed.
func CORS(frontendURL string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

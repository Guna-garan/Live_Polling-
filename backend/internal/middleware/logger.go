package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger logs method, route, status and duration for every
// request. It deliberately never logs request/response bodies, headers,
// or cookies, so credentials and tokens can never end up in logs.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Printf("%s %s %d %s", c.Request.Method, path, c.Writer.Status(), time.Since(start))
	}
}

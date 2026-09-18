package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter tracks one token-bucket limiter per client IP. It's a simple
// in-process limiter — good enough to blunt brute-force/spam on a single
// instance; a multi-instance deployment would want a shared store (e.g.
// Redis) instead, which is straightforward to swap in later since this
// is isolated behind the RateLimit() constructor.
type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

func newIPLimiter(r rate.Limit, burst int) *ipLimiter {
	return &ipLimiter{visitors: make(map[string]*rate.Limiter), r: r, burst: burst}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	lim, ok := l.visitors[ip]
	if !ok {
		lim = rate.NewLimiter(l.r, l.burst)
		l.visitors[ip] = lim
	}
	return lim
}

// RateLimit returns middleware allowing `burst` requests immediately and
// `perSecond` requests/second thereafter, per client IP. Intended for
// sensitive endpoints: /signup, /login, /vote.
func RateLimit(perSecond rate.Limit, burst int) gin.HandlerFunc {
	limiter := newIPLimiter(perSecond, burst)
	return func(c *gin.Context) {
		if !limiter.get(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "Too many requests, please slow down"},
			})
			return
		}
		c.Next()
	}
}

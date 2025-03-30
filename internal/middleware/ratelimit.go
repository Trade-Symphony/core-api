package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	sync.Mutex
	lastAccess map[string]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		lastAccess: make(map[string]time.Time),
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.Lock()
		lastTime, exists := rl.lastAccess[ip]
		now := time.Now()
		rl.lastAccess[ip] = now
		rl.Unlock()

		if exists && now.Sub(lastTime) < time.Second {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded. Please try again in a moment.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

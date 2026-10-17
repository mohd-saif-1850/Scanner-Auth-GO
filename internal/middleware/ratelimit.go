package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"scannerAuth/internal/dto"
)

type RateLimiter struct {
	mu    sync.Mutex
	store map[string]*clientLimit
}

type clientLimit struct {
	count     int
	resetTime int64
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		store: make(map[string]*clientLimit),
	}
}

// RateLimit returns a middleware that limits requests to max per 60-second window
func (rl *RateLimiter) RateLimit(max int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now().Unix()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		limit, exists := rl.store[clientIP]
		if !exists || now >= limit.resetTime {
			rl.store[clientIP] = &clientLimit{
				count:     1,
				resetTime: now + 60,
			}
			c.Next()
			return
		}

		if limit.count >= max {
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Code:    "RATE_LIMIT_EXCEEDED",
				Message: "too many requests",
			})
			c.Abort()
			return
		}

		limit.count++
		c.Next()
	}
}

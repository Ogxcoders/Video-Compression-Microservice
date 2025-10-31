package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl.limit == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		client, exists := rl.requests[clientIP]
		now := time.Now()

		if !exists || now.After(client.resetTime) {
			rl.requests[clientIP] = &clientInfo{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			c.Next()
			return
		}

		if client.count >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded",
				"retry_after": client.resetTime.Sub(now).Seconds(),
			})
			c.Abort()
			return
		}

		client.count++
		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.requests {
			if now.After(client.resetTime.Add(5 * time.Minute)) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

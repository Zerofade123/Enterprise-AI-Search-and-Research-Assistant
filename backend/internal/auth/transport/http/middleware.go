package http

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ratelimitMu sync.Mutex
	lastReset   = time.Now()
	counts      = map[string]int{}
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("request_id", time.Now().UTC().Format(time.RFC3339Nano))
		c.Next()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()
		ratelimitMu.Lock()
		defer ratelimitMu.Unlock()
		if time.Since(lastReset) > time.Minute {
			counts = map[string]int{}
			lastReset = time.Now()
		}
		counts[key]++
		if counts[key] > 60 {
			c.AbortWithStatusJSON(429, gin.H{"success": false, "error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

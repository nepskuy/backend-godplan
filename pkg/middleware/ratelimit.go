package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages rate limiters for each IP address
type IPRateLimiter struct {
	ips map[string]*client
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new rate limiter manager
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*client),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Start background cleanup
	go i.cleanup()

	return i
}

// cleanup removes old entries to prevent memory leaks
func (i *IPRateLimiter) cleanup() {
	for {
		time.Sleep(1 * time.Minute)

		i.mu.Lock()
		for ip, client := range i.ips {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

// GetLimiter returns the rate limiter for the provided IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	c, exists := i.ips[ip]
	if !exists {
		limiter := rate.NewLimiter(i.r, i.b)
		c = &client{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		i.ips[ip] = c
	} else {
		c.lastSeen = time.Now()
	}

	return c.limiter
}

// RateLimitMiddleware creates a middleware for rate limiting based on IP
func RateLimitMiddleware() gin.HandlerFunc {
	// Limit: 5 requests per second, burst of 10
	limiter := NewIPRateLimiter(5, 10)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

package middlewares

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type IPRateLimiter struct {
	ips    map[string]*ClientInfo
	mu     *sync.RWMutex
	rate   int           // maksimalan broj zahteva
	window time.Duration // vremenski prozor
}

type ClientInfo struct {
	count     int
	firstSeen time.Time
}

func NewIPRateLimiter(rate int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:    make(map[string]*ClientInfo),
		mu:     &sync.RWMutex{},
		rate:   rate,
		window: window,
	}

	go limiter.cleanupOldEntries()

	return limiter
}

func (i *IPRateLimiter) cleanupOldEntries() {
	ticker := time.NewTicker(i.window)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		now := time.Now()
		for ip, info := range i.ips {
			if now.Sub(info.firstSeen) > i.window {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

func (i *IPRateLimiter) Allow(ip string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()

	if info, exists := i.ips[ip]; exists {
		if now.Sub(info.firstSeen) > i.window {
			i.ips[ip] = &ClientInfo{
				count:     1,
				firstSeen: now,
			}
			return true
		}

		if info.count >= i.rate {
			log.Printf("Rate limit exceeded for IP: %s (requests: %d)", ip, info.count)
			return false
		}

		info.count++
		return true
	}

	i.ips[ip] = &ClientInfo{
		count:     1,
		firstSeen: now,
	}
	return true
}

func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Gin koristi ClientIP() koji POVERUJE X-Forwarded-For headeru
		// ako SetTrustedProxies() nije pozvan!
		ip := c.ClientIP()

		log.Printf("Request from IP: %s (Path: %s)", ip, c.Request.URL.Path)

		if !limiter.Allow(ip) {
			log.Printf("BLOCKED: Too many requests from IP: %s", ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
				"ip":    ip,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

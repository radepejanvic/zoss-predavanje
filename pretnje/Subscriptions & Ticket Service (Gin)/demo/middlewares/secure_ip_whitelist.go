package middlewares

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SecureIPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	// Parse allowed IPs
	allowed := make([]net.IP, 0, len(allowedIPs))
	for _, ipStr := range allowedIPs {
		if ip := net.ParseIP(ipStr); ip != nil {
			allowed = append(allowed, ip)
		}
	}

	return func(c *gin.Context) {
		// PRAVILNO: Čitamo STVARNU IP adresu
		// RemoteAddr sadrži IP:port stvarne konekcije
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			host = c.Request.RemoteAddr
		}

		realIP := net.ParseIP(host)
		if realIP == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid IP address",
			})
			c.Abort()
			return
		}

		// Proveri da li je STVARNA IP u whitelisti
		isAllowed := false
		for _, allowedIP := range allowed {
			if realIP.Equal(allowedIP) {
				isAllowed = true
				break
			}
		}

		// Dozvoli localhost za testiranje
		if realIP.IsLoopback() {
			isAllowed = true
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":        "Access denied - IP not whitelisted",
				"your_real_ip": host,
				"note":         "We check REAL IP, not X-Forwarded-For headers",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

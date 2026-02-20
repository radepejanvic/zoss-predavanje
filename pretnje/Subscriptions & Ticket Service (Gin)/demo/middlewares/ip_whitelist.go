package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func VulnerableIPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//  GREŠKA: Koristimo c.ClientIP() bez SetTrustedProxies()!
		// ClientIP() čita X-Forwarded-For, X-Real-IP, itd. bez validacije
		clientIP := c.ClientIP()

		// Proveri da li je "client IP" (koji može biti lažiran!) u whitelisti
		isAllowed := false
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied - Only proxy is allowed",
				"message": "Requests are only accepted from the reverse proxy",
				"your_ip": clientIP,
				"hint":    "X-Forwarded-For header determines access",
			})
			c.Abort()
			return
		}

		// Zahtev je "prošao" jer ima odgovarajući X-Forwarded-For header!
		c.Next()
	}
}

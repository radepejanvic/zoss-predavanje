package main

import (
	"gin-subscription-service/db"
	"gin-subscription-service/handlers"
	"gin-subscription-service/middlewares"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Odaberi režim rada:
	// false = Direktan pristup (bez proxy-ja)
	// true  = Iza reverse proxy-ja
	USE_PROXY_MODE_ = true

	// IP adresa reverse proxy-ja (samo ako USE_PROXY_MODE = true)
	TRUSTED_PROXY_IP_ = "203.0.113.50"

	SERVER_PORT_         = ":8080"
	RATE_LIMIT_REQUESTS_ = 10
	RATE_LIMIT_WINDOW_   = time.Minute
)

func main() {
	db.InitDatabase()
	router := gin.Default()

	if USE_PROXY_MODE_ {
		router.SetTrustedProxies([]string{TRUSTED_PROXY_IP_})

		log.Printf(" SECURE MODE: Behind reverse proxy")
		log.Printf(" SetTrustedProxies([%s]) configured", TRUSTED_PROXY_IP_)
		log.Printf(" X-Forwarded-For trusted ONLY from proxy %s", TRUSTED_PROXY_IP_)
		log.Printf(" Spoofed headers from other IPs will be ignored")
	} else {
		// MODE 1: Direktan pristup (bez proxy-ja)
		// Ne verujemo nijednom X-Forwarded-For headeru
		router.SetTrustedProxies(nil)

		log.Printf(" SECURE MODE: Direct access (no proxy)")
		log.Printf("SetTrustedProxies(nil) configured")
		log.Printf("ALL X-Forwarded-For headers will be ignored")
		log.Printf("Using real RemoteAddr IP for rate limiting")
	}

	rateLimiter := middlewares.NewIPRateLimiter(RATE_LIMIT_REQUESTS_, RATE_LIMIT_WINDOW_)
	router.Use(middlewares.RateLimitMiddleware(rateLimiter))

	subscriptionGroup := router.Group("/api/subscriptions")
	{
		subscriptionGroup.POST("/", handlers.CreateSubscription)
		subscriptionGroup.GET("/", handlers.GetSubscriptions)
		subscriptionGroup.GET("/by-email", handlers.GetSubscriptionsByEmail)
		subscriptionGroup.GET("/:id", handlers.GetSubscription)
		subscriptionGroup.PUT("/:id", handlers.UpdateSubscription)
		subscriptionGroup.DELETE("/:id", handlers.DeleteSubscription)
	}

	ticketGroup := router.Group("/api/tickets")
	{
		ticketGroup.POST("/", handlers.CreateTicket)
		ticketGroup.GET("/", handlers.GetTickets)
		ticketGroup.GET("/by-email", handlers.GetTicketsByEmail)
		ticketGroup.GET("/by-subscription", handlers.GetTicketsBySubscription)
		ticketGroup.GET("/:id", handlers.GetTicket)
		ticketGroup.PUT("/:id", handlers.UpdateTicket)
		ticketGroup.DELETE("/:id", handlers.DeleteTicket)
	}

	if err := router.Run(SERVER_PORT_); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

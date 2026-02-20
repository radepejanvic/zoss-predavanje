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
	// IP adresa našeg reverse proxy-ja (simulacija)
	TRUSTED_PROXY_IP = "203.0.113.50"

	// Port na kojem radi aplikacija
	SERVER_PORT = ":8080"

	// Rate limit: maksimalno zahteva po IP adresi
	RATE_LIMIT_REQUESTS = 10
	RATE_LIMIT_WINDOW   = time.Minute
)

func main() {
	db.InitDatabase()

	router := gin.Default()

	// Aplikacija želi da prima zahteve SAMO od proxy-ja (203.0.113.50).
	// ALI middleware koristi c.ClientIP() bez SetTrustedProxies()!
	// Rezultat: Veruje X-Forwarded-For od bilo kog izvora = RANJIVOST!
	router.Use(middlewares.VulnerableIPWhitelistMiddleware([]string{TRUSTED_PROXY_IP}))

	rateLimiter := middlewares.NewIPRateLimiter(RATE_LIMIT_REQUESTS, RATE_LIMIT_WINDOW)

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

	log.Printf("Simulating: Application behind reverse proxy at %s", TRUSTED_PROXY_IP)
	log.Printf("Starting server on %s", SERVER_PORT)
	log.Printf("Rate limit: %d requests per %v per IP", RATE_LIMIT_REQUESTS, RATE_LIMIT_WINDOW)

	if err := router.Run(SERVER_PORT); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

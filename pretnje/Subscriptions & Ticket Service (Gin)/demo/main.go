package main

import (
	"gin-subscription-service/db"
	"gin-subscription-service/handlers"
	"gin-subscription-service/middlewares"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDatabase()

	router := gin.Default()

	// NAMERNO ne pozivamo router.SetTrustedProxies()
	// Ovo omogućava da napadač lažira X-Forwarded-For header i izvede DoS napad

	// Kreiraj rate limiter: maksimalno 10 zahteva po IP adresi u 1 minuti
	rateLimiter := middlewares.NewIPRateLimiter(10, time.Minute)

	// Primeni rate limiter kao globalni middleware
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

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

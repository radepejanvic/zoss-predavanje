package handlers

import (
	"gin-subscription-service/db"
	"gin-subscription-service/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTicket(c *gin.Context) {
	var ticket models.Ticket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket.Status = "open"
	if ticket.Priority == "" {
		ticket.Priority = "medium"
	}

	if err := db.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func GetTickets(c *gin.Context) {
	var tickets []models.Ticket
	if err := db.DB.Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func GetTicket(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket

	if err := db.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func UpdateTicket(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket

	if err := db.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	var updateData models.Ticket
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateData.Subject != "" {
		ticket.Subject = updateData.Subject
	}
	if updateData.Description != "" {
		ticket.Description = updateData.Description
	}
	if updateData.Status != "" {
		ticket.Status = updateData.Status
	}
	if updateData.Priority != "" {
		ticket.Priority = updateData.Priority
	}

	if err := db.DB.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ticket"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket

	if err := db.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if err := db.DB.Delete(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted successfully"})
}

func GetTicketsByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter required"})
		return
	}

	var tickets []models.Ticket
	if err := db.DB.Where("user_email = ?", email).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func GetTicketsBySubscription(c *gin.Context) {
	subscriptionID := c.Query("subscription_id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID parameter required"})
		return
	}

	var tickets []models.Ticket
	if err := db.DB.Where("subscription_id = ?", subscriptionID).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

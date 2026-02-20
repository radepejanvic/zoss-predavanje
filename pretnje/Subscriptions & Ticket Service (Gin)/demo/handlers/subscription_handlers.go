package handlers

import (
	"gin-subscription-service/db"
	"gin-subscription-service/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateSubscription(c *gin.Context) {
	var subscription models.Subscription
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription.Status = "active"
	subscription.StartDate = time.Now()
	subscription.EndDate = time.Now().AddDate(0, 1, 0) // 1 mesec

	if err := db.DB.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

func GetSubscriptions(c *gin.Context) {
	var subscriptions []models.Subscription
	if err := db.DB.Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

func GetSubscription(c *gin.Context) {
	id := c.Param("id")
	var subscription models.Subscription

	if err := db.DB.First(&subscription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func UpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	var subscription models.Subscription

	if err := db.DB.First(&subscription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	var updateData models.Subscription
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription.UserEmail = updateData.UserEmail
	subscription.PlanType = updateData.PlanType
	subscription.UpdatedAt = time.Now()

	if err := db.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func DeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	var subscription models.Subscription

	if err := db.DB.First(&subscription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if err := db.DB.Delete(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully"})
}

func GetSubscriptionsByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter required"})
		return
	}

	var subscriptions []models.Subscription
	if err := db.DB.Where("user_email = ?", email).Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

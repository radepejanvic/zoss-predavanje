package models

import (
	"time"

	"gorm.io/gorm"
)

// Subscription model - predstavlja pretplatu korisnika
type Subscription struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserEmail string         `json:"user_email" binding:"required"`
	PlanType  string         `json:"plan_type" binding:"required"` // "basic", "premium", "enterprise"
	Status    string         `json:"status"`                       // "active", "inactive", "cancelled"
	StartDate time.Time      `json:"start_date"`
	EndDate   time.Time      `json:"end_date"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Ticket model - predstavlja support tiket
type Ticket struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	UserEmail      string         `json:"user_email" binding:"required"`
	SubscriptionID uint           `json:"subscription_id"`
	Subject        string         `json:"subject" binding:"required"`
	Description    string         `json:"description" binding:"required"`
	Status         string         `json:"status"`   // "open", "in_progress", "resolved", "closed"
	Priority       string         `json:"priority"` // "low", "medium", "high", "urgent"
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

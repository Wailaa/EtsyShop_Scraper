package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	ID                     uuid.UUID `gorm:"primaryKey;type:uuid"`
	FirstName              string    `gorm:"type:varchar(50);not null"`
	LastName               string    `gorm:"type:varchar(50);not null"`
	Email                  string    `gorm:"type:varchar(255) ;uniqueIndex;not null"`
	PasswordHashed         string    `gorm:"type:varchar(155)"`
	SubscriptionType       string    `gorm:"type:varchar(55)"`
	EmailVerified          bool      `gorm:"default:false"`
	EmailVerificationToken string    `gorm:"type:varchar(255)"`
	LastTimeLoggedIn       time.Time `gorm:"type.TIMESTAMP"`
	LastTimeLoggedOut      time.Time `gorm:"type.TIMESTAMP"`
	ShopsFollowing         []Shop    `gorm:"many2many:account_shop_following;"`
}

type RegisterAccount struct {
	FirstName        string `json:"first_name" binding:"required"`
	LastName         string `json:"last_name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Password         string `json:"password" binding:"required,min=8"`
	PasswordConfirm  string `json:"password_confirm" binding:"required"`
	SubscriptionType string `json:"subscription_type"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  *Token `json:"access_token"`
	RefreshToken *Token `json:"refresh_token"`
}

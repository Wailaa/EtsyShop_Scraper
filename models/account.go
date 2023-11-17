package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model

	FirstName        string `gorm:"type:varchar(50);not null"`
	LastName         string `gorm:"type:varchar(50);not null"`
	Email            string `gorm:"type:varchar(255) ;uniqueIndex;not null"`
	PasswordHashed   string `gorm:"type:varchar(155)"`
	SubscriptionType string `gorm:"type:varchar(55)"`
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

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

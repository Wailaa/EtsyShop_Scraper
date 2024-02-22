package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	ID                     uuid.UUID     `gorm:"primaryKey;type:uuid"`
	FirstName              string        `gorm:"type:varchar(50);not null"`
	LastName               string        `gorm:"type:varchar(50);not null"`
	Email                  string        `gorm:"type:varchar(255) ;uniqueIndex;not null"`
	PasswordHashed         string        `gorm:"type:varchar(155)"`
	SubscriptionType       string        `gorm:"type:varchar(55)"`
	EmailVerified          bool          `gorm:"default:false"`
	EmailVerificationToken string        `gorm:"type:varchar(255)"`
	LastTimeLoggedIn       time.Time     `gorm:"type.TIMESTAMP"`
	LastTimeLoggedOut      time.Time     `gorm:"type.TIMESTAMP"`
	ShopsFollowing         []Shop        `gorm:"many2many:account_shop_following;"`
	Requests               []ShopRequest `gorm:"foreignKey:AccountID;references:ID"`
}

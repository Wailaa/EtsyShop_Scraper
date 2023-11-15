package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type User struct {
	DB *gorm.DB
}

func NewUserController(DB *gorm.DB) *User {
	return &User{DB}
}

func (s *User) RegisterUser(ctx *gin.Context) {

	var account *models.RegisterAccount

	if err := ctx.ShouldBindJSON(&account); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if account.Password != account.PasswordConfirm {
		message := "Your password and confirmation password do not match"
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": message})
		return
	}

	newAccount := models.Account{
		FirstName:        account.FirstName,
		LastName:         account.LastName,
		Email:            account.Email,
		PasswordHashed:   utils.HashPass(account.Password),
		SubscriptionType: account.SubscriptionType,
	}

	res := s.DB.Create(&newAccount)

	if res.Error != nil {
		if strings.Contains(res.Error.Error(), "email") {
			message := "this email is already in use"
			ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
			return
		}
	}

	message := "user created"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})

}

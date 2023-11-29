package controllers

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"log"
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

	passwardHashed, err := utils.HashPass(account.Password)
	if err != nil {
		log.Fatal(err)
		message := "error while hashing password"
		ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
		return
	}

	EmailVerificationToken, err := utils.CreateVerificationString()
	if err != nil {
		log.Fatal(err)
		message := "error while creating the User"
		ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
		return
	}

	newAccount := &models.Account{
		FirstName:              account.FirstName,
		LastName:               account.LastName,
		Email:                  account.Email,
		PasswordHashed:         passwardHashed,
		SubscriptionType:       account.SubscriptionType,
		EmailVerificationToken: EmailVerificationToken,
	}

	res := s.DB.Create(&newAccount)

	if res.Error != nil {
		if strings.Contains(res.Error.Error(), "email") {
			message := "this email is already in use"
			ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
			return
		}
	}
	utils.SendVerificationEmail(newAccount)
	message := "thank you for registering, please check your email inbox"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})

}
func (s *User) GetAccountByEmail(email string) *models.Account {
	account := &models.Account{}

	result := s.DB.Where("email = ?", email).First(&account)
	if result.Error != nil {

		return account

	}
	newAccount := &models.Account{
		FirstName:        account.FirstName,
		LastName:         account.LastName,
		Email:            account.Email,
		PasswordHashed:   account.PasswordHashed,
		SubscriptionType: account.SubscriptionType,
	}
	return newAccount
}

func (s *User) LoginAccount(ctx *gin.Context) {

	var loginDetails *models.LoginRequest
	config, err := initializer.LoadProjConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)

	}

	if err := ctx.ShouldBindJSON(&loginDetails); err != nil {
		message := "failed to fetch login details"

		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	result := s.GetAccountByEmail(loginDetails.Email)

	if *result == (models.Account{}) {
		message := "user not found"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	if !utils.IsPassVerified(loginDetails.Password, result.PasswordHashed) {
		message := "password is incorrect"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	accessToken, err := utils.CreateJwtToken(config.AccTokenExp)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed", "message": err.Error()})
	}

	refreshToken, err := utils.CreateJwtToken(config.RefTokenExp)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed", "message": err.Error()})
	}
	loginResponse := &models.LoginResponse{
		TokenType:    "Bearer",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	ctx.SetCookie("accessToken", accessToken, 86400, "/", "localhost", false, true)
	ctx.SetCookie("refreshToken", refreshToken, 604800, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, loginResponse)

}

func (s *User) VerifyAccount(ctx *gin.Context) {
	TranID := ctx.Query("TranID")

	VerifyUser := &models.Account{}

	DBCheck := s.DB.Where("email_verification_token = ?", TranID).Find(&VerifyUser).Limit(1)
	if DBCheck.Error != nil {
		log.Fatalln(DBCheck.Error)
		message := "something went wrong while verifing email"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if *VerifyUser == (models.Account{}) {
		message := "Invalid verification code or account does not exists"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if VerifyUser.EmailVerified {
		message := "this link is not valid anymore"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	s.DB.Model(VerifyUser).Updates(map[string]interface{}{"email_verified": true, "email_verification_token": ""})

	message := "Email has been verified"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
}

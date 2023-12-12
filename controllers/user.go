package controllers

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	newUUID := uuid.New()

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
		log.Println(err)
		message := "error while hashing password"
		ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
		return
	}

	EmailVerificationToken, err := utils.CreateVerificationString()
	if err != nil {
		log.Println(err)
		message := "error while creating the User"
		ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
		return
	}

	newAccount := &models.Account{
		ID:                     newUUID,
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
		message := "internal issue"
		ctx.JSON(http.StatusConflict, gin.H{"status": "failed", "message": message})
		return
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
		ID:               account.ID,
		FirstName:        account.FirstName,
		LastName:         account.LastName,
		Email:            account.Email,
		PasswordHashed:   account.PasswordHashed,
		SubscriptionType: account.SubscriptionType,
	}
	return newAccount
}

func (s *User) LoginAccount(ctx *gin.Context) {
	now := time.Now().UTC()
	var loginDetails *models.LoginRequest
	config := initializer.LoadProjConfig(".")

	if err := ctx.ShouldBindJSON(&loginDetails); err != nil {
		message := "failed to fetch login details"
		log.Println(message)
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

	accessToken, err := utils.CreateJwtToken(config.AccTokenExp, result.ID)
	if err != nil {
		log.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed", "message": err.Error()})
		return
	}

	refreshToken, err := utils.CreateJwtToken(config.RefTokenExp, result.ID)
	if err != nil {
		log.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed", "message": err.Error()})
		return
	}

	if err = s.DB.Model(result).Where("id = ?", result.ID).Update("last_time_logged_in", now).Error; err != nil {
		log.Println(err)
	}

	loginResponse := &models.LoginResponse{
		TokenType:    "Bearer",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	ctx.SetCookie("accessToken", string(*accessToken), 86400, "/", "localhost", false, true)
	ctx.SetCookie("refreshToken", string(*refreshToken), 604800, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, loginResponse)

}

func (s *User) LogOutAccount(ctx *gin.Context) {
	now := time.Now().UTC()
	accessToken, _ := ctx.Cookie("accessToken")
	refreshToken, _ := ctx.Cookie("refreshToken")

	if accessToken == "" || refreshToken == "" {
		message := "incorrect token or the session has expired"
		log.Println(message)
		ctx.JSON(http.StatusConflict, gin.H{"status": "failed", "message": message})
		return
	}

	tokenClaims, err := utils.ValidateJWT(accessToken)
	if err != nil {
		log.Println(err)
		return
	}

	if err = s.DB.Model(&models.Account{}).Where("id = ?", tokenClaims.UserUUID).Update("last_time_logged_out", now).Error; err != nil {
		log.Println(err)
		return
	}

	err = utils.BlacklistJWT(accessToken)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = utils.BlacklistJWT(refreshToken)
	if err != nil {
		log.Println(err.Error())
		return
	}

	ctx.SetCookie("accessToken", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("refreshToken", "", -1, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "user logged out successfully"})

}

func (s *User) VerifyAccount(ctx *gin.Context) {
	TranID := ctx.Query("TranID")

	VerifyUser := &models.Account{}

	DBCheck := s.DB.Where("email_verification_token = ?", TranID).Find(&VerifyUser).Limit(1)
	if DBCheck.Error != nil {
		log.Println(DBCheck.Error)
		message := "something went wrong while verifing email"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if *VerifyUser == (models.Account{}) {
		message := "Invalid verification code or account does not exists"
		log.Println(message)
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if VerifyUser.EmailVerified {
		message := "this link is not valid anymore"
		log.Println(message)
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	s.DB.Model(VerifyUser).Updates(map[string]interface{}{"email_verified": true, "email_verification_token": ""})

	message := "Email has been verified"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
}

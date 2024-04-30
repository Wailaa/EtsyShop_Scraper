package controllers

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	DB    *gorm.DB
	utils utils.UtilsProcess
}

func NewUserController(DB *gorm.DB, Process utils.UtilsProcess) *User {
	return &User{
		DB:    DB,
		utils: Process,
	}
}
func NewUserUtilsAccess(Process utils.UtilsProcess) *User {
	return &User{utils: Process}
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
	TokenType    string        `json:"token_type"`
	AccessToken  *models.Token `json:"access_token"`
	RefreshToken *models.Token `json:"refresh_token"`
	User         UserData
}

type UserData struct {
	Name  string
	Email string
	Shops []models.Shop
}

type ReqPassChange struct {
	CurrentPass string `json:"current_password"`
	NewPass     string `json:"new_password"`
	ConfirmPass string `json:"confirm_password"`
}

type UserReqPassChange struct {
	RCP         string `json:"rcp"`
	NewPass     string `json:"new_password"`
	ConfirmPass string `json:"confirm_password"`
}

type UserReqForgotPassword struct {
	Email string `json:"email_account"`
}
type UserController interface {
	GetAccountByID(ID uuid.UUID) (account *models.Account, err error)
}

func (s *User) RegisterUser(ctx *gin.Context) {

	var account *RegisterAccount
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

	passwardHashed, err := s.utils.HashPass(account.Password)
	if err != nil {
		log.Println(err)
		message := "error while hashing password"
		ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
		return
	}

	EmailVerificationToken, err := s.utils.CreateVerificationString()
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

	err = s.utils.SendVerificationEmail(newAccount)
	if err != nil {
		log.Println(err)
		message := "internal error"
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "registraition rejected", "message": message})
		return
	}
	message := "thank you for registering, please check your email inbox"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})

}

func (s *User) GetAccountByID(ID uuid.UUID) (account *models.Account, err error) {

	if err := s.DB.Where("ID = ?", ID).First(&account).Error; err != nil {
		log.Println("no account was Found ,error :", err)
		return nil, err
	}
	return
}

func (s *User) GetAccountByEmail(email string) *models.Account {
	account := &models.Account{}

	result := s.DB.Where("email = ?", email).First(&account)
	if result.Error != nil {

		return account

	}

	return account
}

func (s *User) LoginAccount(ctx *gin.Context) {

	var loginDetails *LoginRequest
	config := initializer.LoadProjConfig(".")

	if err := ctx.ShouldBindJSON(&loginDetails); err != nil {
		message := "failed to fetch login details"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	result := s.GetAccountByEmail(loginDetails.Email)

	if reflect.DeepEqual(*result, models.Account{}) {
		message := "user not found"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	if !s.utils.IsPassVerified(loginDetails.Password, result.PasswordHashed) {
		message := "password is incorrect"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	accessToken, err := s.utils.CreateJwtToken(config.AccTokenExp, result.ID)
	if err != nil {
		log.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed", "message": err.Error()})
		return
	}

	refreshToken, err := s.utils.CreateJwtToken(config.RefTokenExp, result.ID)
	if err != nil {
		log.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed", "message": err.Error()})
		return
	}

	if err = s.UpdateLastTimeLoggedIn(result); err != nil {
		log.Println(err)
	}

	if err := s.JoinShopFollowing(result); err != nil {
		log.Println(err)
		return
	}

	loginResponse := s.GenerateLoginResponce(result, accessToken, refreshToken)

	ctx.SetCookie("access_token", string(*accessToken), int(config.AccTokenExp.Seconds()), "/", "localhost", false, true)
	ctx.SetCookie("refresh_token", string(*refreshToken), int(config.RefTokenExp.Seconds()), "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, loginResponse)

}

func GetTokens(ctx *gin.Context) (map[string]*models.Token, error) {
	tokens := make(map[string]*models.Token)
	if accessToken, err := ctx.Cookie("access_token"); err == nil {
		tokens["access_token"] = models.NewToken(accessToken)
	}
	if refreshToken, err := ctx.Cookie("refresh_token"); err == nil {
		tokens["refresh_token"] = models.NewToken(refreshToken)
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("failed to retrieve both tokens ")
	}
	return tokens, nil
}

func (s *User) LogOutAccount(ctx *gin.Context) {
	now := time.Now().UTC()
	var userUUID uuid.UUID

	tokenList, err := GetTokens(ctx)
	if err != nil {
		log.Println(err.Error())
		return
	}

	for tokenName, token := range tokenList {
		if reflect.ValueOf(userUUID).IsZero() {
			tokenClaims, err := s.utils.ValidateJWT(token)
			if err != nil {
				log.Println(err)
				return
			}

			userUUID = tokenClaims.UserUUID

			if err = s.DB.Model(&models.Account{}).Where("id = ?", userUUID).Update("last_time_logged_out", now).Error; err != nil {
				log.Println(err)
				ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "failed to update logout details"})
				return
			}

		}

		err = s.utils.BlacklistJWT(token)
		if err != nil {
			log.Println(err.Error())
		}

		ctx.SetCookie(tokenName, "", -1, "/", "localhost", false, true)
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "user logged out successfully"})

}

func (s *User) VerifyAccount(ctx *gin.Context) {
	TranID := ctx.Query("TranID")

	VerifyUser := &models.Account{}

	DBCheck := s.DB.Where("email_verification_token = ?", TranID).Find(&VerifyUser).Limit(1)
	if DBCheck.Error != nil {
		log.Println(DBCheck.Error)
		message := "something went wrong while verifying email"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if reflect.DeepEqual(*VerifyUser, models.Account{}) {
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

func (s *User) ChangePass(ctx *gin.Context) {

	reqPassChange := ReqPassChange{}
	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

	if err := ctx.ShouldBindJSON(&reqPassChange); err != nil {
		message := "failed to fetch change password request"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	Account, err := s.GetAccountByID(currentUserUUID)
	if err != nil {
		message := "failed to fetch change password request"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	if reflect.DeepEqual(Account, &models.Account{}) {
		message := "user not found"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	if !s.utils.IsPassVerified(reqPassChange.CurrentPass, Account.PasswordHashed) {
		message := "password is incorrect"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	if reqPassChange.NewPass == reqPassChange.ConfirmPass {
		passwardHashed, err := s.utils.HashPass(reqPassChange.NewPass)
		if err != nil {
			log.Println(err)
			message := "error while hashing password"
			ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
			return
		}
		s.DB.Model(Account).Update("password_hashed", passwardHashed)
	} else {
		message := "new password is not confirmed"
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": message})
		ctx.Abort()
		return
	}
	message := "password changed"
	ctx.JSON(http.StatusOK, gin.H{"status": "registraition rejected", "message": message})
	s.LogOutAccount(ctx)
}

func (s *User) ForgotPassReq(ctx *gin.Context) {
	ForgotAccountPass := &UserReqForgotPassword{}
	if err := ctx.ShouldBindJSON(&ForgotAccountPass); err != nil {
		message := "failed to fetch change password request"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	Account := s.GetAccountByEmail(ForgotAccountPass.Email)
	if reflect.DeepEqual(Account, &models.Account{}) {
		message := "reset password request denied , no account associated  "
		log.Println(message)
		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}
	ResetPassToken, err := s.utils.CreateVerificationString()
	if err != nil {
		message := "failed to create resetpassword token"
		log.Println(message, "error :", err)
		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
		return
	}
	Account.RequestChangePass = true
	Account.AccountPassResetToken = ResetPassToken
	s.DB.Save(Account)

	go s.utils.SendResetPassEmail(Account)
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})

}

func (s *User) ResetPass(ctx *gin.Context) {

	reqChangePass := UserReqPassChange{}
	VerifyUser := &models.Account{}

	if err := ctx.ShouldBindJSON(&reqChangePass); err != nil {
		message := "failed to fetch change password request"
		log.Println(message)
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": message})
		return
	}

	if reqChangePass.NewPass != reqChangePass.ConfirmPass {
		message := "Passwords are not the same"
		log.Println(message)
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	DBCheck := s.DB.Where("account_pass_reset_token = ?", reqChangePass.RCP).Find(&VerifyUser).Limit(1)
	if DBCheck.Error != nil {
		log.Println(DBCheck.Error)
		message := "something went wrong while resetting password"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if reflect.DeepEqual(*VerifyUser, models.Account{}) {
		message := "Invalid verification code or account does not exists"
		log.Println(message)
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}
	if VerifyUser.AccountPassResetToken != reqChangePass.RCP {
		message := "failed to change password"
		log.Println(message, "rcp and AccountPassResetToken no match")
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	if VerifyUser.AccountPassResetToken == "" {
		message := "this link is not valid anymore"
		log.Println(message)
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	newPasswardHashed, err := s.utils.HashPass(reqChangePass.NewPass)
	if err != nil {
		log.Println(err)
		message := "error while hashing password"
		ctx.JSON(http.StatusConflict, gin.H{"status": "registraition rejected", "message": message})
		return
	}

	s.DB.Model(VerifyUser).Updates(map[string]interface{}{"request_change_pass": false, "account_pass_reset_token": "", "password_hashed": newPasswardHashed})

	message := "Password changed successfully"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
}

func (s *User) UpdateLastTimeLoggedIn(Account *models.Account) error {
	now := time.Now()
	if err := s.DB.Model(Account).Where("id = ?", Account.ID).Update("last_time_logged_in", now).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *User) JoinShopFollowing(Account *models.Account) error {

	if err := s.DB.Preload("ShopsFollowing").First(Account, Account.ID).Error; err != nil {
		log.Println(err)
		return err
	}

	for i := range Account.ShopsFollowing {
		if err := s.DB.Preload("ShopMenu").Preload("Reviews").Preload("Member").First(&Account.ShopsFollowing[i]).Error; err != nil {
			log.Println(err)
			return err
		}
	}

	return nil

}

func (s *User) GenerateLoginResponce(Account *models.Account, AccessToken, RefreshToken *models.Token) *LoginResponse {

	user := UserData{
		Name:  Account.FirstName,
		Email: Account.Email,
		Shops: Account.ShopsFollowing,
	}

	loginResponse := &LoginResponse{
		TokenType:    "Bearer",
		AccessToken:  AccessToken,
		RefreshToken: RefreshToken,
		User:         user,
	}

	return loginResponse
}

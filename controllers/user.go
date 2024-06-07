package controllers

import (
	"errors"
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/repository"
	"EtsyScraper/utils"
)

type User struct {
	DB    *gorm.DB
	utils utils.UtilsProcess
	User  repository.UserRepository
}

func NewUserController(DB *gorm.DB, Process utils.UtilsProcess, UserDB repository.UserRepository) *User {
	return &User{
		DB:    DB,
		utils: Process,
		User:  UserDB,
	}
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

func (s *User) RegisterUser(ctx *gin.Context) {

	var account *RegisterAccount

	if err := ctx.ShouldBindJSON(&account); err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if account.Password != account.PasswordConfirm {
		HandleResponse(ctx, nil, http.StatusBadRequest, "Your password and confirmation password do not match", nil)
		return
	}

	passwardHashed, err := s.utils.HashPass(account.Password)
	if err != nil {
		HandleResponse(ctx, err, http.StatusConflict, "error while hashing password", nil)
		return
	}

	EmailVerificationToken, err := s.utils.CreateVerificationString()
	if err != nil {
		HandleResponse(ctx, err, http.StatusConflict, "error while creating the User", nil)
		return
	}

	newAccount, err := s.CreateNewAccountRecord(account, passwardHashed, EmailVerificationToken)
	if err != nil {
		HandleResponse(ctx, err, http.StatusConflict, err.Error(), nil)
		return
	}

	err = s.utils.SendVerificationEmail(newAccount)
	if err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "internal error", nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "thank you for registering, please check your email inbox", nil)

}

func (s *User) LoginAccount(ctx *gin.Context) {

	var loginDetails *LoginRequest
	config := initializer.LoadProjConfig(".")

	if err := ctx.ShouldBindJSON(&loginDetails); err != nil {
		HandleResponse(ctx, err, http.StatusNotFound, "failed to fetch login details", nil)
		return
	}

	result := s.User.GetAccountByEmail(loginDetails.Email)

	if reflect.DeepEqual(*result, models.Account{}) {
		HandleResponse(ctx, nil, http.StatusNotFound, "user not found", nil)
		return
	}

	if !s.utils.IsPassVerified(loginDetails.Password, result.PasswordHashed) {
		HandleResponse(ctx, nil, http.StatusNotFound, "password is incorrect", nil)
		return
	}

	accessToken, err := s.utils.CreateJwtToken(config.AccTokenExp, result.ID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	refreshToken, err := s.utils.CreateJwtToken(config.RefTokenExp, result.ID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err = s.User.UpdateLastTimeLoggedIn(result); err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "internal error", nil)
		return
	}

	if err := s.User.JoinShopFollowing(result); err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "internal error", nil)
		return
	}

	loginResponse := s.GenerateLoginResponse(result, accessToken, refreshToken)

	ctx.SetCookie("access_token", string(*accessToken), int(config.AccTokenExp.Seconds()), "/", config.ClientOrigin, false, true)
	ctx.SetCookie("refresh_token", string(*refreshToken), int(config.RefTokenExp.Seconds()), "/", config.ClientOrigin, false, true)

	HandleResponse(ctx, nil, http.StatusOK, "", loginResponse)

}

func (s *User) LogOutAccount(ctx *gin.Context) {
	var userUUID uuid.UUID

	tokenList, err := s.utils.GetTokens(ctx)
	if err != nil {
		HandleResponse(ctx, err, http.StatusOK, "", nil)
		return
	}

	for tokenName, token := range tokenList {
		if userUUID == uuid.Nil {
			tokenClaims, err := s.utils.ValidateJWT(token)
			if err != nil {
				HandleResponse(ctx, err, http.StatusOK, "", nil)
				return
			}

			userUUID = tokenClaims.UserUUID

			if err = s.User.UpdateLastTimeLoggedOut(userUUID); err != nil {
				HandleResponse(ctx, err, http.StatusForbidden, "failed to update logout details", nil)
				return
			}

		}

		err = s.utils.BlacklistJWT(token)
		if err != nil {
			log.Println(err.Error())
		}

		ctx.SetCookie(tokenName, "", -1, "/", "localhost", false, true)
	}
	HandleResponse(ctx, nil, http.StatusOK, "user logged out successfully", nil)

}

func (s *User) VerifyAccount(ctx *gin.Context) {
	TranID := ctx.Query("TranID")

	VerifyUser := &models.Account{}

	if err := s.DB.Where("email_verification_token = ?", TranID).Find(&VerifyUser).Limit(1).Error; err != nil { //create method
		HandleResponse(ctx, err, http.StatusForbidden, "something went wrong while verifying email", nil)
		return
	}

	if reflect.DeepEqual(*VerifyUser, models.Account{}) {
		HandleResponse(ctx, nil, http.StatusForbidden, "Invalid verification code or account does not exists", nil)
		return
	}

	if VerifyUser.EmailVerified {
		HandleResponse(ctx, nil, http.StatusForbidden, "this link is not valid anymore", nil)
		return
	}

	if err := s.User.UpdateAccountAfterVerify(VerifyUser); err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "internal error", nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "Email has been verified", nil)

}

func (s *User) ChangePass(ctx *gin.Context) {

	reqPassChange := ReqPassChange{}
	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

	if err := ctx.ShouldBindJSON(&reqPassChange); err != nil {
		HandleResponse(ctx, err, http.StatusNotFound, "failed to fetch change password request", nil)
	}

	Account, err := s.User.GetAccountByID(currentUserUUID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusNotFound, "failed to fetch change password request", nil)
		return
	}

	if reflect.DeepEqual(Account, &models.Account{}) {
		HandleResponse(ctx, nil, http.StatusNotFound, "user not found", nil)
		return
	}

	if !s.utils.IsPassVerified(reqPassChange.CurrentPass, Account.PasswordHashed) {
		HandleResponse(ctx, nil, http.StatusNotFound, "password is incorrect", nil)
		return
	}

	if reqPassChange.NewPass != reqPassChange.ConfirmPass {
		HandleResponse(ctx, nil, http.StatusUnauthorized, "new password is not confirmed", nil)
		ctx.Abort()
		return
	}

	passwardHashed, err := s.utils.HashPass(reqPassChange.NewPass)
	if err != nil {
		HandleResponse(ctx, err, http.StatusConflict, "error while hashing password", nil)
		return
	}

	if err := s.User.UpdateAccountNewPass(Account, passwardHashed); err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "internal error", nil)
		return
	}

	HandleResponse(ctx, err, http.StatusOK, "password changed", nil)

	s.LogOutAccount(ctx)
}

func (s *User) ForgotPassReq(ctx *gin.Context) {
	ForgotAccountPass := &UserReqForgotPassword{}
	if err := ctx.ShouldBindJSON(&ForgotAccountPass); err != nil {
		HandleResponse(ctx, err, http.StatusNotFound, "failed to fetch change password request", nil)
		return
	}

	Account := s.User.GetAccountByEmail(ForgotAccountPass.Email)
	if reflect.DeepEqual(Account, &models.Account{}) {
		err := errors.New("reset password request denied , no account associated  ")
		HandleResponse(ctx, err, http.StatusOK, "", nil)
		return
	}
	ResetPassToken, err := s.utils.CreateVerificationString()
	if err != nil {
		err := errors.New("failed to create resetpassword token")
		HandleResponse(ctx, err, http.StatusOK, "", nil)
		return
	}
	Account.RequestChangePass = true
	Account.AccountPassResetToken = ResetPassToken

	if err := s.User.SaveAccount(Account); err != nil {
		HandleResponse(ctx, err, http.StatusInternalServerError, "Failed to save account", nil)
		return
	}

	go s.utils.SendResetPassEmail(Account)

	HandleResponse(ctx, err, http.StatusOK, "", nil)

}

func (s *User) ResetPass(ctx *gin.Context) {

	reqChangePass := UserReqPassChange{}
	VerifyUser := &models.Account{}

	if err := ctx.ShouldBindJSON(&reqChangePass); err != nil {
		HandleResponse(ctx, err, http.StatusNotFound, "failed to fetch change password request", nil)
		return
	}

	if reqChangePass.NewPass != reqChangePass.ConfirmPass {
		err := errors.New("passwords are not the same")
		HandleResponse(ctx, err, http.StatusForbidden, err.Error(), nil)
		return
	}

	if err := s.DB.Where("account_pass_reset_token = ?", reqChangePass.RCP).Find(&VerifyUser).Limit(1).Error; err != nil {
		HandleResponse(ctx, err, http.StatusForbidden, "something went wrong while resetting password", nil)
		return
	}

	if reflect.DeepEqual(*VerifyUser, models.Account{}) || VerifyUser.AccountPassResetToken == "" {
		err := errors.New("invalid verification code or account does not exists")
		HandleResponse(ctx, err, http.StatusForbidden, "something went wrong while resetting password", nil)
		return
	}
	if VerifyUser.AccountPassResetToken != reqChangePass.RCP {
		err := errors.New("rcp and AccountPassResetToken no match")
		HandleResponse(ctx, err, http.StatusForbidden, "failed to change password", nil)
		return
	}

	newPasswardHashed, err := s.utils.HashPass(reqChangePass.NewPass)
	if err != nil {
		HandleResponse(ctx, err, http.StatusConflict, "error while hashing password", nil)
		return
	}

	if err = s.UpdateAccountAfterResetPass(VerifyUser, newPasswardHashed); err != nil {
		HandleResponse(ctx, err, http.StatusConflict, "internal error", nil)
		return
	}

	HandleResponse(ctx, nil, http.StatusOK, "Password changed successfully", nil)
}

func (s *User) GenerateLoginResponse(Account *models.Account, AccessToken, RefreshToken *models.Token) *LoginResponse {

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

func (s *User) UpdateAccountAfterResetPass(Account *models.Account, newPasswardHashed string) error {

	err := s.DB.Model(Account).Updates(map[string]interface{}{"request_change_pass": false, "account_pass_reset_token": "", "password_hashed": newPasswardHashed}).Error
	if err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *User) CreateNewAccountRecord(account *RegisterAccount, passwardHashed, EmailVerificationToken string) (*models.Account, error) {
	newUUID := uuid.New()

	newAccount := &models.Account{
		ID:                     newUUID,
		FirstName:              account.FirstName,
		LastName:               account.LastName,
		Email:                  account.Email,
		PasswordHashed:         passwardHashed,
		SubscriptionType:       account.SubscriptionType,
		EmailVerificationToken: EmailVerificationToken,
	}

	if err := s.DB.Create(newAccount).Error; err != nil {
		if utils.StringContains(err.Error(), "email") {
			message := errors.New("this email is already in use")
			return newAccount, utils.HandleError(message)
		}

		return newAccount, utils.HandleError(err)
	}
	return newAccount, nil
}

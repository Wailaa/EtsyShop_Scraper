package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/repository"
	"EtsyScraper/utils"
)

func AuthMiddleWare(Process utils.UtilsProcess, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		config := initializer.LoadProjConfig(".")

		JwtTokens, err := Process.GetTokens(ctx)
		if err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, err.Error()+" please login again", nil)
			return
		}

		accessToken, accHasValue := JwtTokens["access_token"]
		refreshToken, refHasValue := JwtTokens["refresh_token"]

		if accHasValue {
			if ok := IsAuthorized(ctx, Process, accessToken, userRepo); ok {
				ctx.Next()
				return
			}
		}

		if !refHasValue {
			HandleResponse(ctx, err, http.StatusUnauthorized, "no refreshToken found", nil)
			return
		}

		if ok := IsAuthorized(ctx, Process, refreshToken, userRepo); ok {

			accessToken, err = Process.RefreshAccToken(refreshToken)
			if err != nil {
				HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
			}
			ctx.SetCookie("access_token", string(*accessToken), int(config.AccTokenExp.Seconds()), "/", "localhost", false, true)
			ctx.Next()
			return
		}

		HandleResponse(ctx, err, http.StatusUnauthorized, "login required ", nil)
	}
}

func Authorization(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
		Account, err := userRepo.GetAccountByID(currentUserUUID)
		if err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
			return
		}

		if !Account.EmailVerified {
			err := errors.New("email not verified")
			HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		ctx.Next()

	}
}
func IsAccountFollowingShop(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
		ShopID := ctx.Param("shopID")

		ShopIDToUint, err := utils.StringToUint(ShopID)
		if err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, "failed to get Shop id", nil)
			return
		}

		Account, err := userRepo.GetAccountWithShops(currentUserUUID)
		if err != nil {
			HandleResponse(ctx, err, http.StatusInternalServerError, "internal error", nil)
			return
		}

		isFollow := false
		for _, shop := range Account.ShopsFollowing {
			if shop.ID == ShopIDToUint {
				isFollow = true
				break
			}
		}

		if !isFollow {
			HandleResponse(ctx, err, http.StatusUnauthorized, "no permission", nil)
			return
		}

		ctx.Next()
	}
}

func IsAuthorized(ctx *gin.Context, Process utils.UtilsProcess, Token *models.Token, userRepo repository.UserRepository) bool {

	userClaims, err := Process.ValidateJWT(Token)
	if err != nil {
		return false
	}

	isBlackListed, err := Process.IsJWTBlackListed(Token)
	if err != nil || isBlackListed {
		return false
	}

	_, err = userRepo.GetAccountByID(userClaims.UserUUID)
	if err != nil {
		HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
		return false
	}
	ctx.Set("currentUserUUID", userClaims.UserUUID)

	return true
}

package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
)

func AuthMiddleWare(Process utils.UtilsProcess) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		config := initializer.LoadProjConfig(".")

		user := &models.Account{}
		JwtTokens, err := GetTokens(ctx)
		if err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, err.Error()+" please login again", nil)
			return
		}

		accessToken, accHasValue := JwtTokens["access_token"]
		refreshToken, refHasValue := JwtTokens["refresh_token"]

		if accHasValue {
			userClaims, errClaims := Process.ValidateJWT(accessToken)
			isBlackListed, err := Process.IsJWTBlackListed(accessToken)
			if !isBlackListed && err == nil && errClaims == nil {

				if err := initializer.DB.Where("id = ?", userClaims.UserUUID).First(user).Error; err != nil {
					HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
					return
				}
				ctx.Set("currentUserUUID", user.ID)
				ctx.Next()
				return
			}
		}
		if !refHasValue {
			HandleResponse(ctx, err, http.StatusUnauthorized, "no refreshToken found", nil)
			return
		}
		accessToken, err = Process.RefreshAccToken(refreshToken)
		if err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		userClaims, errClaims := Process.ValidateJWT(accessToken)
		if errClaims != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, "auth failed because of "+errClaims.Error(), nil)
			return
		}

		ctx.SetCookie("access_token", string(*accessToken), int(config.AccTokenExp.Seconds()), "/", "localhost", false, true)

		if err := initializer.DB.Where("id = ?", userClaims.UserUUID).First(user).Error; err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, err.Error(), nil)
			return
		}

		ctx.Set("currentUserUUID", user.ID)
		ctx.Next()
	}
}

func Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		Account := models.Account{}
		currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

		if err := initializer.DB.Where("id = ?", currentUserUUID).First(&Account).Error; err != nil {
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
func IsAccountFollowingShop() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
		ShopID := ctx.Param("shopID")

		ShopIDToUint, err := utils.StringToUint(ShopID)
		if err != nil {
			HandleResponse(ctx, err, http.StatusUnauthorized, "failed to get Shop id", nil)
			return
		}

		Account := models.Account{}

		if err := initializer.DB.Preload("ShopsFollowing").First(&Account, "id = ?", currentUserUUID).Error; err != nil {
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

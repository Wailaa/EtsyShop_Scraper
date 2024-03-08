package controllers

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		config := initializer.LoadProjConfig(".")
		user := &models.Account{}
		JwtTokens, err := GetTokens(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": err.Error() + " please login again"})
			ctx.Abort()
			return
		}

		accessToken, accHasValue := JwtTokens["access_token"]
		refreshToken, refHasValue := JwtTokens["refresh_token"]

		if accHasValue {
			userClaims, errClaims := utils.ValidateJWT(accessToken)
			isBlackListed, err := utils.IsJWTBlackListed(accessToken)
			if !isBlackListed && err == nil && errClaims == nil {

				result := initializer.DB.Where("id = ?", userClaims.UserUUID).First(user)
				if result.Error != nil {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": result.Error})
					ctx.Abort()
					return
				}
				ctx.Set("currentUserUUID", user.ID)
				ctx.Next()
				return
			}
		}
		if !refHasValue {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "no refreshToken found"})
			ctx.Abort()
			return
		}
		accessToken, err = utils.RefreshAccToken(refreshToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": err.Error()})
			ctx.Abort()
			return
		}
		userClaims, errClaims := utils.ValidateJWT(accessToken)
		if errClaims != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "auth failed because of " + errClaims.Error()})
			ctx.Abort()
			return
		}

		ctx.SetCookie("access_token", string(*accessToken), int(config.AccTokenExp.Seconds()), "/", "localhost", false, true)

		result := initializer.DB.Where("id = ?", userClaims.UserUUID).First(user)
		if result.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": result.Error})
			ctx.Abort()
			return
		}

		ctx.Set("currentUserUUID", user.ID)
		ctx.Next()
	}
}

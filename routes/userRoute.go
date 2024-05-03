package routes

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/utils"

	"github.com/gin-gonic/gin"
)

type UserRoute struct {
}

func (ur *UserRoute) GeneraluserRoutes(server *gin.Engine) {

	utils := &utils.Utils{}

	router := server.Group("/auth")

	register := controllers.NewUserController(initializer.DB, utils).RegisterUser
	confirmEmail := controllers.NewUserController(initializer.DB, utils).VerifyAccount
	login := controllers.NewUserController(initializer.DB, utils).LoginAccount
	logOut := controllers.NewUserController(initializer.DB, utils).LogOutAccount
	forgotPass := controllers.NewUserController(initializer.DB, utils).ForgotPassReq
	changePass := controllers.NewUserController(initializer.DB, utils).ChangePass
	resetPass := controllers.NewUserController(initializer.DB, utils).ResetPass

	router.POST("/register", register)
	router.POST("/login", login)
	router.GET("/logout", logOut)
	router.GET("/verifyaccount", confirmEmail)
	router.POST("/forgotpassword", forgotPass)
	router.POST("/resetpassword", resetPass)
	router.POST("/changepassword", controllers.AuthMiddleWare(utils), controllers.Authorization(), changePass)
}

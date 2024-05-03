package routes

import (
	"EtsyScraper/controllers"
	"EtsyScraper/utils"

	"github.com/gin-gonic/gin"
)

type UserRoute struct {
	UserController UserControllerInterface
}

type UserControllerInterface interface {
	RegisterUser(c *gin.Context)
	VerifyAccount(c *gin.Context)
	LoginAccount(c *gin.Context)
	LogOutAccount(c *gin.Context)
	ForgotPassReq(c *gin.Context)
	ChangePass(c *gin.Context)
	ResetPass(c *gin.Context)
}

func (ur *UserRoute) GeneraluserRoutes(server *gin.Engine) {

	utils := &utils.Utils{}

	router := server.Group("/auth")

	register := ur.UserController.RegisterUser
	confirmEmail := ur.UserController.VerifyAccount
	login := ur.UserController.LoginAccount
	logOut := ur.UserController.LogOutAccount
	forgotPass := ur.UserController.ForgotPassReq
	changePass := ur.UserController.ChangePass
	resetPass := ur.UserController.ResetPass

	router.POST("/register", register)
	router.POST("/login", login)
	router.GET("/logout", logOut)
	router.GET("/verifyaccount", confirmEmail)
	router.POST("/forgotpassword", forgotPass)
	router.POST("/resetpassword", resetPass)
	router.POST("/changepassword", controllers.AuthMiddleWare(utils), controllers.Authorization(), changePass)
}

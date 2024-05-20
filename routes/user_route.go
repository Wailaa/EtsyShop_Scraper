package routes

import (
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

func NewUserRouteController(process UserControllerInterface) *UserRoute {
	return &UserRoute{UserController: process}
}

func (ur *UserRoute) GeneraluserRoutes(server *gin.Engine, authentication, authorization gin.HandlerFunc) {

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
	router.POST("/changepassword", authentication, authorization, changePass)
}
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTMLRoutes struct {
}

func NewHTMLRouter() *HTMLRoutes {
	return &HTMLRoutes{}
}

func (d *HTMLRoutes) GeneralHTMLRoutes(server *gin.Engine, authentication, authorization gin.HandlerFunc, path string) {

	server.Static("/static", "./static")
	server.LoadHTMLGlob(path)

	server.GET("/reset_password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "resetPass.html", nil)
	})

	server.GET("/change_password", authentication, authorization, func(c *gin.Context) {
		c.HTML(http.StatusOK, "changePass.html", nil)
	})

	server.GET("/log_in", func(c *gin.Context) {
		c.HTML(http.StatusOK, "logIn.html", nil)
	})

	server.GET("/verify_account", func(c *gin.Context) {
		c.HTML(http.StatusOK, "verifyAccount.html", nil)
	})

	server.GET("/stats", authentication, authorization, func(c *gin.Context) {
		c.HTML(http.StatusOK, "stats.html", nil)
	})

	server.GET("/", authentication, authorization, func(c *gin.Context) {
		c.HTML(http.StatusOK, "mainPage.html", nil)
	})
}

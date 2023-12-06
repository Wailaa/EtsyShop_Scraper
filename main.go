package main

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"net/http"

	"github.com/gin-gonic/gin"

	"fmt"
	"log"
)

var server *gin.Engine

func init() {
	config, err := initializer.LoadProjConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	initializer.DataBaseConnect(&config)
	initializer.RedisDBConnect(&config)

	initializer.DB.AutoMigrate(&models.Account{})
	fmt.Println("Migration is completed")

}
func main() {

	config, err := initializer.LoadProjConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)

	}

	server = gin.Default()

	router := server.Group("/auth")
	router.GET("/test", func(ctx *gin.Context) {
		message := "Welcome to EtsyScraper a"
		ctx.JSON(http.StatusOK, gin.H{"HTTPstatus": http.StatusOK, "message": message})
	})
	register := controllers.NewUserController(initializer.DB).RegisterUser
	confirmEmail := controllers.NewUserController(initializer.DB).VerifyAccount
	login := controllers.NewUserController(initializer.DB).LoginAccount
	logOut := controllers.NewUserController(initializer.DB).LogOutAccount
	router.POST("/register", register)
	router.POST("/login", login)
	router.POST("/logout", logOut)
	router.GET("/verifyaccount", confirmEmail)

	log.Fatal(server.Run(":" + config.ServerPort))
}

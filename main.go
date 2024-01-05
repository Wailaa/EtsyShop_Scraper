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
	config := initializer.LoadProjConfig(".")
	initializer.DataBaseConnect(&config)
	initializer.RedisDBConnect(&config)

	initializer.DB.AutoMigrate(models.ModelsGroup...)
	fmt.Println("Migration is completed")

}
func main() {

	config := initializer.LoadProjConfig(".")

	server = gin.Default()

	router := server.Group("/auth")
	router.GET("/test", controllers.AuthMiddleWare(), func(ctx *gin.Context) {
		message := "Welcome to EtsyScraper"
		ctx.JSON(http.StatusOK, gin.H{"HTTPstatus": http.StatusOK, "message": message})
	})
	register := controllers.NewUserController(initializer.DB).RegisterUser
	confirmEmail := controllers.NewUserController(initializer.DB).VerifyAccount
	login := controllers.NewUserController(initializer.DB).LoginAccount
	logOut := controllers.NewUserController(initializer.DB).LogOutAccount
	router.POST("/register", register)
	router.POST("/login", login)
	router.GET("/logout", logOut)
	router.GET("/verifyaccount", confirmEmail)

	shopRoute := server.Group("/shop")
	createShop := controllers.NewShopController(initializer.DB).CreateNewShop
	followShop := controllers.NewShopController(initializer.DB).FollowShop
	unFollowShop := controllers.NewShopController(initializer.DB).UnFollowShop

	shopRoute.GET("/create_shop", controllers.AuthMiddleWare(), createShop)
	shopRoute.GET("/follow_shop", controllers.AuthMiddleWare(), followShop)
	shopRoute.GET("/unfollow_shop", controllers.AuthMiddleWare(), unFollowShop)

	log.Fatal(server.Run(":" + config.ServerPort))
}

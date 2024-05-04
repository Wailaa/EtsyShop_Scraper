package main

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/routes"
	scheduleUpdates "EtsyScraper/scheduleUpdateTask"
	scrap "EtsyScraper/scraping"
	"EtsyScraper/utils"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"fmt"
	"log"
)

var server *gin.Engine

func init() {
	config := initializer.LoadProjConfig(".")
	initializer.DataBaseConnect(&config)
	initializer.RedisDBConnect(&config)
	scheduleUpdates.StartScheduleScrapUpdate()
	initializer.DB.AutoMigrate(models.ModelsGroup...)
	fmt.Println("Migration is completed")

}
func main() {

	config := initializer.LoadProjConfig(".")

	server = gin.Default()

	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowHeaders:     []string{"Origin, Content-Type, Accept"},
		AllowMethods:     []string{"GET, POST"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	utils := &utils.Utils{}
	Scraper := &scrap.Scraper{}
	Creatrors := &controllers.ShopCreators{DB: initializer.DB}
	implShop := controllers.Shop{DB: initializer.DB, Scraper: Scraper, Process: Creatrors}

	userRoutes := routes.NewUserRouteController(controllers.NewUserController(initializer.DB, utils))
	userRoutes.GeneraluserRoutes(server, controllers.AuthMiddleWare(utils), controllers.Authorization())

	shopRoutes := routes.NewShopRouteController(&implShop)
	shopRoutes.GeneralShopRoutes(server, controllers.AuthMiddleWare(utils), controllers.Authorization())

	server.Static("/static", "./static")
	server.LoadHTMLGlob("static/templates/*")
	server.GET("/reset_password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "resetPass.html", nil)
	})
	server.GET("/change_password", controllers.AuthMiddleWare(utils), controllers.Authorization(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "changePass.html", nil)
	})
	server.GET("/log_in", func(c *gin.Context) {
		c.HTML(http.StatusOK, "logIn.html", nil)
	})
	server.GET("/verify_account", func(c *gin.Context) {
		c.HTML(http.StatusOK, "verifyAccount.html", nil)
	})
	server.GET("/stats", controllers.AuthMiddleWare(utils), controllers.Authorization(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "stats.html", nil)
	})
	server.GET("/", controllers.AuthMiddleWare(utils), controllers.Authorization(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "mainPage.html", nil)
	})

	log.Fatal(server.Run(":" + config.ServerPort))

}

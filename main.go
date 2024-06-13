package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/repository"
	"EtsyScraper/routes"
	scheduleUpdates "EtsyScraper/scheduleUpdateTask"
	scrap "EtsyScraper/scraping"
	"EtsyScraper/utils"
)

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

	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.ClientOrigin + config.ServerPort},
		AllowHeaders:     []string{"Origin, Content-Type, Accept"},
		AllowMethods:     []string{"GET, POST"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	utils := &utils.Utils{}
	Scraper := &scrap.Scraper{}
	Repository := &repository.DataBase{DB: initializer.DB}
	implShop := controllers.Shop{Scraper: Scraper, User: Repository, Shop: Repository}
	implShop.Operations = &implShop

	userRoutes := routes.NewUserRouteController(controllers.NewUserController(utils, Repository))
	userRoutes.GeneraluserRoutes(server, controllers.AuthMiddleWare(utils), controllers.Authorization())

	shopRoutes := routes.NewShopRouteController(&implShop)
	shopRoutes.GeneralShopRoutes(server, controllers.AuthMiddleWare(utils), controllers.Authorization(), controllers.IsAccountFollowingShop())

	templatesFilesPath := "./static/templates/*"
	htmlRoutes := routes.NewHTMLRouter()
	htmlRoutes.GeneralHTMLRoutes(server, controllers.AuthMiddleWare(utils), controllers.Authorization(), templatesFilesPath)

	log.Fatal(server.Run(":" + config.ServerPort))

}

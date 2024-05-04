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
	"strconv"
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

	shopRoute := server.Group("/shop")
	createNewShopRequest := controllers.NewShopController(implShop).CreateNewShopRequest
	followShop := controllers.NewShopController(implShop).FollowShop
	unFollowShop := controllers.NewShopController(implShop).UnFollowShop
	getShopByID := controllers.NewShopController(implShop).HandleGetShopByID
	getAllItemsByShopID := controllers.NewShopController(implShop).HandleGetItemsByShopID
	getAllSoldItemsByShopID := controllers.NewShopController(implShop).GetSoldItemsByShopID
	getShopStats := controllers.NewShopController(implShop).ProcessStatsRequest
	getItemsCountByShopID := controllers.NewShopController(implShop).GetItemsCountByShopID

	shopRoute.GET("/create_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), createNewShopRequest)
	shopRoute.GET("/follow_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), followShop)
	shopRoute.GET("/unfollow_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), unFollowShop)
	shopRoute.GET("/:shopID", controllers.AuthMiddleWare(utils), controllers.Authorization(), getShopByID)
	shopRoute.GET("/:shopID/all_items", controllers.AuthMiddleWare(utils), controllers.Authorization(), getAllItemsByShopID)

	shopRoute.GET("/:shopID/all_sold_items", controllers.AuthMiddleWare(utils), controllers.Authorization(), func(ctx *gin.Context) {
		ShopID := ctx.Param("shopID")
		ShopIDToUint, err := strconv.ParseUint(ShopID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to get Shop id"})
			return
		}
		Items, err := getAllSoldItemsByShopID(uint(ShopIDToUint))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, Items)
	})

	shopRoute.GET("/:shopID/items_count", controllers.AuthMiddleWare(utils), controllers.Authorization(), func(ctx *gin.Context) {
		ShopID := ctx.Param("shopID")
		ShopIDToUint, err := strconv.ParseUint(ShopID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to get Shop id"})
			return
		}
		Items, err := getItemsCountByShopID(uint(ShopIDToUint))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, Items)
	})

	shopRoute.GET("/stats/:shopID/:period", controllers.AuthMiddleWare(utils), controllers.Authorization(), func(ctx *gin.Context) {
		ShopID := ctx.Param("shopID")
		Period := ctx.Param("period")
		ShopIDToUint, err := strconv.ParseUint(ShopID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to get Shop id"})
			return
		}
		err = getShopStats(ctx, uint(ShopIDToUint), Period)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}

	})

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

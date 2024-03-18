package main

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	scheduleUpdates "EtsyScraper/scheduleUpdateTask"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fmt"
	"log"
)

var server *gin.Engine

func init() {
	config := initializer.LoadProjConfig(".")
	initializer.DataBaseConnect(&config)
	initializer.RedisDBConnect(&config)
	scheduleUpdates.ScheduleScrapUpdate()
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
	createNewShopRequest := controllers.NewShopController(initializer.DB).CreateNewShopRequest
	followShop := controllers.NewShopController(initializer.DB).FollowShop
	unFollowShop := controllers.NewShopController(initializer.DB).UnFollowShop
	getShopByID := controllers.NewShopController(initializer.DB).GetShopByID
	getAllItemsByShopID := controllers.NewShopController(initializer.DB).GetItemsByShopID
	getAllSoldItemsByShopID := controllers.NewShopController(initializer.DB).GetSoldItemsByShopID
	getShopStats := controllers.NewShopController(initializer.DB).ProcessStatsRequest

	shopRoute.GET("/create_shop", controllers.AuthMiddleWare(), createNewShopRequest)
	shopRoute.GET("/follow_shop", controllers.AuthMiddleWare(), followShop)
	shopRoute.GET("/unfollow_shop", controllers.AuthMiddleWare(), unFollowShop)
	shopRoute.GET("/:shopID", controllers.AuthMiddleWare(), func(ctx *gin.Context) {
		ShopID := ctx.Param("shopID")
		ShopIDToUint, err := strconv.ParseUint(ShopID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to get Shop id"})
			return
		}
		Shop, err := getShopByID(uint(ShopIDToUint))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, Shop)
	})

	shopRoute.GET("/:shopID/all_items", controllers.AuthMiddleWare(), func(ctx *gin.Context) {
		ShopID := ctx.Param("shopID")
		ShopIDToUint, err := strconv.ParseUint(ShopID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to get Shop id"})
			return
		}
		Items, err := getAllItemsByShopID(uint(ShopIDToUint))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, Items)
	})

	shopRoute.GET("/:shopID/all_sold_items", controllers.AuthMiddleWare(), func(ctx *gin.Context) {
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

	shopRoute.GET("/stats/:shopID/:period", controllers.AuthMiddleWare(), func(ctx *gin.Context) {
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

	log.Fatal(server.Run(":" + config.ServerPort))

}

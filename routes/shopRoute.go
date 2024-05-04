package routes

import (
	"EtsyScraper/controllers"
	"EtsyScraper/utils"

	"github.com/gin-gonic/gin"
)

type ShopRoutes struct {
	ShopController ShopRoutesInterface
}

type ShopRoutesInterface interface {
	CreateNewShopRequest(ctx *gin.Context)
	FollowShop(ctx *gin.Context)
	UnFollowShop(ctx *gin.Context)
	HandleGetShopByID(ctx *gin.Context)
	HandleGetItemsByShopID(ctx *gin.Context)
	HandleGetSoldItemsByShopID(ctx *gin.Context)
	ProcessStatsRequest(ctx *gin.Context)
	HandleGetItemsCountByShopID(ctx *gin.Context)
}

func NewShopRouteController(process ShopRoutesInterface) *ShopRoutes {
	return &ShopRoutes{ShopController: process}
}

func (us *ShopRoutes) GeneralShopRoutes(server *gin.Engine) {
	utils := &utils.Utils{}

	shopRoute := server.Group("/shop")

	createNewShopRequest := us.ShopController.CreateNewShopRequest
	followShop := us.ShopController.FollowShop
	unFollowShop := us.ShopController.UnFollowShop
	getShopByID := us.ShopController.HandleGetShopByID
	getAllItemsByShopID := us.ShopController.HandleGetItemsByShopID
	getAllSoldItemsByShopID := us.ShopController.HandleGetSoldItemsByShopID
	getShopStats := us.ShopController.ProcessStatsRequest
	getItemsCountByShopID := us.ShopController.HandleGetItemsCountByShopID

	shopRoute.GET("/create_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), createNewShopRequest)
	shopRoute.GET("/follow_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), followShop)
	shopRoute.GET("/unfollow_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), unFollowShop)
	shopRoute.GET("/:shopID", controllers.AuthMiddleWare(utils), controllers.Authorization(), getShopByID)
	shopRoute.GET("/:shopID/all_items", controllers.AuthMiddleWare(utils), controllers.Authorization(), getAllItemsByShopID)
	shopRoute.GET("/:shopID/all_sold_items", controllers.AuthMiddleWare(utils), controllers.Authorization(), getAllSoldItemsByShopID)
	shopRoute.GET("/:shopID/items_count", controllers.AuthMiddleWare(utils), controllers.Authorization(), getItemsCountByShopID)
	shopRoute.GET("/stats/:shopID/:period", controllers.AuthMiddleWare(utils), controllers.Authorization(), getShopStats)

}

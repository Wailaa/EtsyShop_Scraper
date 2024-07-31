package routes

import (
	"EtsyScraper/controllers"

	"github.com/gin-gonic/gin"
)

type ShopRoutes struct {
	ShopController controllers.ShopRoutesInterface
}

func NewShopRouteController(process controllers.ShopRoutesInterface) *ShopRoutes {
	return &ShopRoutes{ShopController: process}
}

func (us *ShopRoutes) GeneralShopRoutes(server *gin.Engine, authentication, authorization, isfollowingShop gin.HandlerFunc) {

	shopRoute := server.Group("/shop")

	createNewShopRequest := us.ShopController.CreateNewShopRequest
	followShop := us.ShopController.FollowShop
	unFollowShop := us.ShopController.UnFollowShop
	getShopByID := us.ShopController.HandleGetShopByID
	getAllItemsByShopID := us.ShopController.HandleGetItemsByShopID
	getAllSoldItemsByShopID := us.ShopController.HandleGetSoldItemsByShopID
	getShopStats := us.ShopController.ProcessStatsRequest
	getItemsCountByShopID := us.ShopController.HandleGetItemsCountByShopID

	shopRoute.POST("/create_shop", authentication, authorization, createNewShopRequest)
	shopRoute.POST("/follow_shop", authentication, authorization, followShop)
	shopRoute.POST("/unfollow_shop", authentication, authorization, unFollowShop)
	shopRoute.GET("/:shopID", authentication, authorization, isfollowingShop, getShopByID)
	shopRoute.GET("/:shopID/all_items", authentication, authorization, isfollowingShop, getAllItemsByShopID)
	shopRoute.GET("/:shopID/all_sold_items", authentication, authorization, isfollowingShop, getAllSoldItemsByShopID)
	shopRoute.GET("/:shopID/items_count", authentication, authorization, isfollowingShop, getItemsCountByShopID)
	shopRoute.GET("/stats/:shopID/:period", authentication, authorization, isfollowingShop, getShopStats)

}

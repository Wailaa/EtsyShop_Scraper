package routes

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	scrap "EtsyScraper/scraping"
	"EtsyScraper/utils"

	"github.com/gin-gonic/gin"
)

type ShopRoutes struct {
}

func NewShopRoutesController() *ShopRoutes {
	return &ShopRoutes{}
}

func (us *ShopRoutes) GeneralShopRoutes(server *gin.Engine) {
	utils := &utils.Utils{}
	Scraper := &scrap.Scraper{}
	Creatrors := &controllers.ShopCreators{DB: initializer.DB}
	implShop := controllers.Shop{DB: initializer.DB, Scraper: Scraper, Process: Creatrors}

	shopRoute := server.Group("/shop")

	createNewShopRequest := controllers.NewShopController(implShop).CreateNewShopRequest
	followShop := controllers.NewShopController(implShop).FollowShop
	unFollowShop := controllers.NewShopController(implShop).UnFollowShop
	getShopByID := controllers.NewShopController(implShop).HandleGetShopByID
	getAllItemsByShopID := controllers.NewShopController(implShop).HandleGetItemsByShopID
	getAllSoldItemsByShopID := controllers.NewShopController(implShop).HandleGetSoldItemsByShopID
	getShopStats := controllers.NewShopController(implShop).ProcessStatsRequest
	getItemsCountByShopID := controllers.NewShopController(implShop).HandleGetItemsCountByShopID

	shopRoute.GET("/create_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), createNewShopRequest)
	shopRoute.GET("/follow_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), followShop)
	shopRoute.GET("/unfollow_shop", controllers.AuthMiddleWare(utils), controllers.Authorization(), unFollowShop)
	shopRoute.GET("/:shopID", controllers.AuthMiddleWare(utils), controllers.Authorization(), getShopByID)
	shopRoute.GET("/:shopID/all_items", controllers.AuthMiddleWare(utils), controllers.Authorization(), getAllItemsByShopID)
	shopRoute.GET("/:shopID/all_sold_items", controllers.AuthMiddleWare(utils), controllers.Authorization(), getAllSoldItemsByShopID)
	shopRoute.GET("/:shopID/items_count", controllers.AuthMiddleWare(utils), controllers.Authorization(), getItemsCountByShopID)
	shopRoute.GET("/stats/:shopID/:period", controllers.AuthMiddleWare(utils), controllers.Authorization(), getShopStats)

}

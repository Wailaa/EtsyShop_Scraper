package controllers

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"EtsyScraper/models"
	"EtsyScraper/repository"
	scrap "EtsyScraper/scraping"
)

type Shop struct {
	Scraper    scrap.ScrapeUpdateProcess
	Operations ShopOperations
	User       repository.UserRepository
	Shop       repository.ShopRepository
}

func NewShopController(implementSHOP Shop) *Shop {
	return &Shop{

		Scraper:    implementSHOP.Scraper,
		Operations: &implementSHOP,
		User:       implementSHOP.User,
		Shop:       implementSHOP.Shop,
	}
}

type NewShopRequest struct {
	ShopName string `json:"new_shop_name"`
}

type FollowShopRequest struct {
	FollowShopName string `json:"follow_shop"`
}

type UnFollowShopRequest struct {
	UnFollowShopName string `json:"unfollow_shop"`
}
type ResponseSoldItemInfo struct {
	Name           string
	ItemID         uint
	OriginalPrice  float64
	CurrencySymbol string
	SalePrice      float64
	DiscoutPercent string
	ItemLink       string
	Available      bool
	SoldQuantity   int
}

type DailySoldStats struct {
	TotalSales   int     `json:"total_sales"`
	DailyRevenue float64 `json:"daily_revenue"`
	Items        []models.Item
}
type itemsCount struct {
	Available       int
	OutOfProduction int
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

type ShopOperations interface {
	GetShopByID(ID uint) (*models.Shop, error)
	CreateNewShop(ShopRequest *models.ShopRequest) error
	GetItemsByShopID(ID uint) ([]models.Item, error)
	CreateShopRequest(ShopRequest *models.ShopRequest) error
	GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error)
	GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
	UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
	CreateSoldStats(dailyShopSales []models.DailyShopSales) (map[string]DailySoldStats, error)
	EstablishAccountShopRelation(requestedShop *models.Shop, userID uuid.UUID) error
	SaveShopToDB(scrappedShop *models.Shop, ShopRequest *models.ShopRequest) error
	UpdateShopMenuToDB(Shop *models.Shop, ShopRequest *models.ShopRequest) error
	CreateOutOfProdMenu(Shop *models.Shop, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) error
	CheckAndUpdateOutOfProdMenu(AllMenus []models.MenuItem, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) (bool, error)
	GetItemsBySoldItems(SoldItems []models.SoldItems) ([]models.Item, error)
}

var queueMutex sync.Mutex

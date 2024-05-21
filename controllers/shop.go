package controllers

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
)

type Shop struct {
	DB         *gorm.DB
	Scraper    scrap.ScrapeUpdateProcess
	Operations ShopOperations
}

func NewShopController(implementSHOP Shop) *Shop {
	return &Shop{
		DB:         implementSHOP.DB,
		Scraper:    implementSHOP.Scraper,
		Operations: &implementSHOP,
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

type ShopOperations interface {
	GetShopByName(ShopName string) (shop *models.Shop, err error)
	CreateNewShop(ShopRequest *models.ShopRequest) error
	GetItemsByShopID(ID uint) ([]models.Item, error)
	GetAverageItemPrice(ShopID uint) (float64, error)
	CreateShopRequest(ShopRequest *models.ShopRequest) error
	GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error)
	GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
	UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
}

var queueMutex sync.Mutex

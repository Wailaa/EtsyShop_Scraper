package controllers

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
)

type Shop struct {
	DB      *gorm.DB
	Process ShopMethodExecutor
	Scraper scrap.ScrapeUpdateProcess
}

func NewShopController(implementSHOP Shop) *Shop {

	return &Shop{
		DB:      implementSHOP.DB,
		Process: implementSHOP.Process,
		Scraper: implementSHOP.Scraper,
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

type ShopUpdater interface {
	UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
}
type ShopMethodExecutor interface {
	ExecuteGetShopByName(dispatch ExecShopMethodProcess, ShopName string) (*models.Shop, error)
	ExecuteGetItemsByShopID(dispatch ExecShopMethodProcess, ID uint) ([]models.Item, error)
	ExecuteGetAverageItemPrice(dispatch ExecShopMethodProcess, ShopID uint) (float64, error)
	ExecuteCreateShopRequest(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest) error
	ExecuteCreateShop(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest)
	ExecuteUpdateSellingHistory(dispatch ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	ExecuteUpdateDiscontinuedItems(dispatch ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
	ExecuteGetTotalRevenue(dispatch ExecShopMethodProcess, ShopID uint, AverageItemPrice float64) (float64, error)
	ExecuteGetSoldItemsByShopID(dispatch ExecShopMethodProcess, ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	ExecuteGetSellingStatsByPeriod(dispatch ExecShopMethodProcess, ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
}

type ExecShopMethodProcess interface {
	GetShopByName(ShopName string) (shop *models.Shop, err error)
	CreateNewShop(ShopRequest *models.ShopRequest) error
	GetItemsByShopID(ID uint) ([]models.Item, error)
	GetAverageItemPrice(ShopID uint) (float64, error)
	CreateShopRequest(ShopRequest *models.ShopRequest) error
	GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error)
	GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
}

var queueMutex sync.Mutex

type ShopCreators struct {
	DB *gorm.DB
}

func NewShopCreators(DB *gorm.DB) *ShopCreators {
	return &ShopCreators{DB}
}

func (ps *ShopCreators) ExecuteCreateShop(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest) {
	dispatch.CreateNewShop(ShopRequest)
}

func (ps *ShopCreators) ExecuteUpdateSellingHistory(dispatch ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {
	err := dispatch.UpdateSellingHistory(Shop, Task, ShopRequest)
	return err
}

func (ps *ShopCreators) ExecuteGetAverageItemPrice(dispatch ExecShopMethodProcess, ShopID uint) (float64, error) {
	averagePrice, err := dispatch.GetAverageItemPrice(ShopID)
	return averagePrice, err
}

func (ps *ShopCreators) ExecuteUpdateDiscontinuedItems(dispatch ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	ScrappedSoldItems, err := dispatch.UpdateDiscontinuedItems(Shop, Task, ShopRequest)
	return ScrappedSoldItems, err
}

func (ps *ShopCreators) ExecuteGetTotalRevenue(dispatch ExecShopMethodProcess, ShopID uint, AverageItemPrice float64) (float64, error) {
	Average, err := dispatch.GetTotalRevenue(ShopID, AverageItemPrice)
	return Average, err
}

func (ps *ShopCreators) ExecuteGetSoldItemsByShopID(dispatch ExecShopMethodProcess, ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	SoldItems, err := dispatch.GetSoldItemsByShopID(ID)
	return SoldItems, err
}

func (ps *ShopCreators) ExecuteGetSellingStatsByPeriod(dispatch ExecShopMethodProcess, ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error) {
	SoldItems, err := dispatch.GetSellingStatsByPeriod(ShopID, timePeriod)
	return SoldItems, err
}

func (ps *ShopCreators) ExecuteCreateShopRequest(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest) error {
	err := dispatch.CreateShopRequest(ShopRequest)
	return err
}

func (ps *ShopCreators) ExecuteGetItemsByShopID(dispatch ExecShopMethodProcess, ID uint) ([]models.Item, error) {
	items, err := dispatch.GetItemsByShopID(ID)
	return items, err
}

func (ps *ShopCreators) ExecuteGetShopByName(dispatch ExecShopMethodProcess, ShopName string) (*models.Shop, error) {
	shop, err := dispatch.GetShopByName(ShopName)
	return shop, err
}

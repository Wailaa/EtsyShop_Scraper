package controllers

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"EtsyScraper/utils"
)

type Shop struct {
	DB      *gorm.DB
	Process ShopProcess
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
type ShopProcess interface {
	GetShopByName(ShopName string) (shop *models.Shop, err error)
	GetItemsByShopID(ID uint) (items []models.Item, err error)
	GetAverageItemPrice(ShopID uint) (float64, error)
	CreateShopRequest(ShopRequest *models.ShopRequest) error
	ExecuteCreateShop(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest)
	ExecuteUpdateSellingHistory(dispatch ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	ExecuteUpdateDiscontinuedItems(dispatch ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
	ExecuteGetTotalRevenue(dispatch ExecShopMethodProcess, ShopID uint, AverageItemPrice float64) (float64, error)
	ExecuteGetSoldItemsByShopID(dispatch ExecShopMethodProcess, ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	ExecuteGetSellingStatsByPeriod(dispatch ExecShopMethodProcess, ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
}

type ExecShopMethodProcess interface {
	CreateNewShop(ShopRequest *models.ShopRequest) error
	GetAverageItemPrice(ShopID uint) (float64, error)
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

func (pr *ShopCreators) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	if ShopRequest.AccountID == uuid.Nil {
		err := errors.New("no AccountID was passed")
		return utils.HandleError(err)
	}

	if err := pr.DB.Save(ShopRequest).Error; err != nil {
		return utils.HandleError(err)
	}

	return nil
}

func (pr *ShopCreators) GetShopByName(ShopName string) (shop *models.Shop, err error) {

	if err = pr.DB.Preload("Member").Preload("ShopMenu.Menu.Items").Preload("Reviews.ReviewsTopic").Where("name = ?", ShopName).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ,error")
	}
	return
}

func (ps *ShopCreators) GetItemsByShopID(ID uint) (items []models.Item, err error) {
	shop := &models.Shop{}
	if err := ps.DB.Preload("ShopMenu.Menu.Items").Where("id = ?", ID).First(shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found")
	}

	for _, menu := range shop.ShopMenu.Menu {
		items = append(items, menu.Items...)
	}
	return
}

func (s *ShopCreators) GetAverageItemPrice(ShopID uint) (float64, error) {
	var averagePrice float64

	if err := s.DB.Table("items").
		Joins("JOIN menu_items ON items.menu_item_id = menu_items.id").
		Joins("JOIN shop_menus ON menu_items.shop_menu_id = shop_menus.id").
		Joins("JOIN shops ON shop_menus.shop_id = shops.id").
		Where("shops.id = ? AND items.original_price > 0 ", ShopID).
		Select("AVG(items.original_price) as average_price").
		Row().Scan(&averagePrice); err != nil {

		return 0, utils.HandleError(err)
	}
	averagePrice = RoundToTwoDecimalDigits(averagePrice)

	return averagePrice, nil
}

package controllers

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
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
	GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error)
	GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
}

var queueMutex sync.Mutex

func (s *Shop) CreateNewShop(ShopRequest *models.ShopRequest) error {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	scrappedShop, err := s.Scraper.ScrapShop(ShopRequest.ShopName)
	if err != nil {
		message := fmt.Sprintf("failed to initiate Shop while handling ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	scrappedShop.CreatedByUserID = ShopRequest.AccountID

	if err = s.SaveShopToDB(scrappedShop, ShopRequest); err != nil {
		return utils.HandleError(err)
	}

	log.Println("starting Shop's menu scraping for ShopRequest.ID: ", ShopRequest.ID)

	scrapeMenu := s.Scraper.ScrapAllMenuItems(scrappedShop)

	if err = s.UpdateShopMenuToDB(scrapeMenu, ShopRequest); err != nil {
		return utils.HandleError(err)

	}

	Task := new(models.TaskSchedule)

	if scrapeMenu.HasSoldHistory && scrapeMenu.TotalSales > 0 {
		log.Println("Shop's selling history initiated for ShopRequest.ID: ", ShopRequest.ID)

		if err := s.Process.ExecuteUpdateSellingHistory(s, scrapeMenu, Task, ShopRequest); err != nil {
			ShopRequest.Status = "failed"
			s.Process.CreateShopRequest(ShopRequest)
			message := fmt.Sprintf("Shop's selling history failed for ShopRequest.ID: %v", ShopRequest.ID)
			return utils.HandleError(err, message)

		}
	} else {
		ShopRequest.Status = "done"
		s.Process.CreateShopRequest(ShopRequest)
	}
	return nil
}

func (s *Shop) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {

	ScrappedSoldItems, err := s.Process.ExecuteUpdateDiscontinuedItems(s, Shop, Task, ShopRequest)
	if err != nil {
		ShopRequest.Status = "failed"
		s.Process.CreateShopRequest(ShopRequest)

		message := fmt.Sprintf("Shop's selling history failed while initiating UpdateDiscontinuedItems for ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	if len(ScrappedSoldItems) == 0 {
		err := fmt.Errorf("empty scrapped Sold data")
		return utils.HandleError(err)
	}

	AllItems, err := s.Process.GetItemsByShopID(Shop.ID)
	if err != nil {
		return utils.HandleError(err)
	}

	ScrappedSoldItems, dailyRevenue := PopulateItemIDsFromListings(ScrappedSoldItems, AllItems)

	ScrappedSoldItems = ReverseSoldItems(ScrappedSoldItems)

	if err = s.SaveSoldItemsToDB(ScrappedSoldItems); err != nil {
		return utils.HandleError(err)
	}

	if Task.UpdateSoldItems > 0 {

		if err = s.UpdateDailySales(ScrappedSoldItems, Shop.ID, dailyRevenue); err != nil {
			return utils.HandleError(err)
		}
	}

	ShopRequest.Status = "done"
	log.Printf("Shop's selling history successfully saved %v items for ShopRequest.ID: %v \n", len(ScrappedSoldItems), ShopRequest.ID)
	s.Process.CreateShopRequest(ShopRequest)

	return nil
}

func (s *Shop) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {

	FilterSoldItems := map[uint]struct{}{}

	scrapSoldItems, NewTask := s.Scraper.ScrapSalesHistory(Shop.Name, Task)
	if !NewTask.IsScrapeFinished {
		go s.SoldItemsTask(Shop, NewTask, ShopRequest)
	}

	if len(scrapSoldItems) == 0 {
		return scrapSoldItems, nil
	}

	getAllItems, err := s.Process.GetItemsByShopID(Shop.ID)
	if err != nil {
		return nil, utils.HandleError(err)
	}
	SoldOutItems := FilterSoldOutItems(scrapSoldItems, getAllItems, FilterSoldItems)

	isOutOfProduction, err := s.CheckAndUpdateOutOfProdMenu(Shop.ShopMenu.Menu, SoldOutItems, ShopRequest)
	if err != nil {
		return nil, utils.HandleError(err)

	}

	if len(SoldOutItems) != 0 && !isOutOfProduction {
		if err := s.CreateOutOfProdMenu(Shop, SoldOutItems, ShopRequest); err != nil {
			return nil, utils.HandleError(err)
		}

	}

	return scrapSoldItems, nil
}

func (s *Shop) GetShopByID(ID uint) (shop *models.Shop, err error) {

	if err := s.DB.Preload("Member").Preload("ShopMenu.Menu").Preload("Reviews.ReviewsTopic").Where("id = ?", ID).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ")

	}

	shop.AverageItemsPrice, err = s.Process.GetAverageItemPrice(shop.ID)
	if err != nil {
		return nil, utils.HandleError(err, "error while calculating item avearage price")
	}

	if !shop.HasSoldHistory {
		shop.Revenue = shop.AverageItemsPrice * float64(shop.TotalSales)
		return
	}

	shop.Revenue, err = s.Process.ExecuteGetTotalRevenue(s, shop.ID, shop.AverageItemsPrice)
	if err != nil {
		return nil, utils.HandleError(err, "error while calculating shop's revenue")
	}

	return
}

func (s *Shop) GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error) {

	soldItems, err := s.Process.ExecuteGetSoldItemsByShopID(s, ShopID)
	if err != nil {
		return 0, utils.HandleError(err, "error while calculating revenue")
	}
	revenue := CalculateTotalRevenue(soldItems, AverageItemPrice)
	return revenue, nil
}

func (s *Shop) SoldItemsTask(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {

	randTimeSet := time.Duration(rand.Intn(79) + 10)
	durationUntilNextTask := time.Until(time.Now().Add(randTimeSet * time.Second))

	time.AfterFunc(durationUntilNextTask, func() {
		s.Process.ExecuteUpdateSellingHistory(s, Shop, Task, ShopRequest)
	})
	return nil
}

func (s *Shop) GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error) {

	dailyShopSales := []models.DailyShopSales{}

	if err := s.DB.Where("shop_id = ? AND created_at > ?", ShopID, timePeriod).Find(&dailyShopSales).Error; err != nil {
		return nil, utils.HandleError(err)
	}

	stats, err := s.CreateSoldStats(dailyShopSales)
	if err != nil {
		return nil, utils.HandleError(err)
	}
	return stats, nil
}

func (s *Shop) UpdateAccountShopRelation(requestedShop *models.Shop, UserID uuid.UUID) error {
	account := &models.Account{}

	if err := s.DB.Preload("ShopsFollowing").Where("id = ?", UserID).First(&account).Error; err != nil {
		return utils.HandleError(err)
	}

	if err := s.DB.Model(&account).Association("ShopsFollowing").Delete(requestedShop); err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *Shop) EstablishAccountShopRelation(requestedShop *models.Shop, userID uuid.UUID) error {
	Utils := &utils.Utils{}
	currentAccount, err := NewUserController(s.DB, Utils).GetAccountByID(userID)
	if err != nil {
		return utils.HandleError(err)
	}

	currentAccount.ShopsFollowing = append(currentAccount.ShopsFollowing, *requestedShop)
	if err := s.DB.Save(&currentAccount).Error; err != nil {
		return utils.HandleError(err)
	}

	return nil
}

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

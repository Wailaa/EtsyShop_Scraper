package controllers

import (
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"EtsyScraper/utils"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shop struct {
	DB      *gorm.DB
	Process ShopCreatorGetters
	Scraper scrap.ScrapeUpdateProcess
}
type ShopCreators struct {
	DB *gorm.DB
}

func NewShopController(implementSHOP Shop) *Shop {

	return &Shop{
		DB:      implementSHOP.DB,
		Process: implementSHOP.Process,
		Scraper: implementSHOP.Scraper,
	}
}
func NewShopCreators(DB *gorm.DB) *ShopCreators {
	return &ShopCreators{DB}
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
	SoldQauntity   int
}

type DailySoldStats struct {
	TotalSales int `json:"total_sales"`
	Items      []models.Item
}
type itemsCount struct {
	Available       int
	OutOfProduction int
}

func CreateSoldItemInfo(Item *models.Item) *ResponseSoldItemInfo {
	newSoldItem := &ResponseSoldItemInfo{
		Name:           Item.Name,
		ItemID:         Item.ID,
		OriginalPrice:  Item.OriginalPrice,
		CurrencySymbol: Item.CurrencySymbol,
		SalePrice:      Item.SalePrice,
		DiscoutPercent: Item.DiscoutPercent,
		ItemLink:       Item.ItemLink,
		Available:      Item.Available,
	}
	return newSoldItem
}

type ShopController interface {
	UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
}
type ShopCreatorGetters interface {
	GetShopByName(ShopName string) (shop *models.Shop, err error)
	GetItemsByShopID(ID uint) (items []models.Item, err error)
	GetAvarageItemPrice(ShopID uint) (float64, error)
	CreateShopRequest(ShopRequest *models.ShopRequest) error
	ExecuteCreateShop(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest)
	ExecuteUpdateSellingHistory(dispatch ShopController, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error
	ExecuteUpdateDiscontinuedItems(dispatch ShopController, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error)
	ExecuteGetTotalRevenue(dispatch ExecShopMethodProcess, ShopID uint, AvarageItemPrice float64) (float64, error)
	ExecuteGetSoldItemsByShopID(dispatch ExecShopMethodProcess, ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	ExecuteGetSellingStatsByPeriod(dispatch ExecShopMethodProcess, ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
}

type ExecShopMethodProcess interface {
	CreateNewShop(ShopRequest *models.ShopRequest) error
	GetTotalRevenue(ShopID uint, AvarageItemPrice float64) (float64, error)
	GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error)
	GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error)
}

var queueMutex sync.Mutex

func (s *Shop) CreateNewShopRequest(ctx *gin.Context) {

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
	ShopRequest := &models.ShopRequest{}
	var shop NewShopRequest

	if err := ctx.ShouldBindJSON(&shop); err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to get the Shop's name"})
		return
	}

	ShopRequest.AccountID = currentUserUUID
	ShopRequest.ShopName = shop.ShopName

	existedShop, err := s.Process.GetShopByName(shop.ShopName)
	if err != nil && err.Error() != "record not found" {
		log.Println(err)
		ShopRequest.Status = "failed"
		s.Process.CreateShopRequest(ShopRequest)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "internal error"})
		return
	} else if existedShop != nil {
		ShopRequest.Status = "denied"
		s.Process.CreateShopRequest(ShopRequest)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Shop already exists"})
		return
	}

	ShopRequest.Status = "Pending"
	s.Process.CreateShopRequest(ShopRequest)

	message := "shop request received successfully"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": message})
	go s.Process.ExecuteCreateShop(s, ShopRequest)

}
func (ps *ShopCreators) ExecuteCreateShop(dispatch ExecShopMethodProcess, ShopRequest *models.ShopRequest) {
	dispatch.CreateNewShop(ShopRequest)
}
func (ps *ShopCreators) ExecuteUpdateSellingHistory(dispatch ShopController, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {
	err := dispatch.UpdateSellingHistory(Shop, Task, ShopRequest)
	return err
}
func (ps *ShopCreators) ExecuteUpdateDiscontinuedItems(dispatch ShopController, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	ScrappedSoldItems, err := dispatch.UpdateDiscontinuedItems(Shop, Task, ShopRequest)
	return ScrappedSoldItems, err
}

func (ps *ShopCreators) ExecuteGetTotalRevenue(dispatch ExecShopMethodProcess, ShopID uint, AvarageItemPrice float64) (float64, error) {
	Avarage, err := dispatch.GetTotalRevenue(ShopID, AvarageItemPrice)
	return Avarage, err
}
func (ps *ShopCreators) ExecuteGetSoldItemsByShopID(dispatch ExecShopMethodProcess, ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	SoldItems, err := dispatch.GetSoldItemsByShopID(ID)
	return SoldItems, err
}
func (ps *ShopCreators) ExecuteGetSellingStatsByPeriod(dispatch ExecShopMethodProcess, ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error) {
	SoldItems, err := dispatch.GetSellingStatsByPeriod(ShopID, timePeriod)
	return SoldItems, err
}

func (s *Shop) CreateNewShop(ShopRequest *models.ShopRequest) error {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	scrappedShop, err := s.Scraper.ScrapShop(ShopRequest.ShopName)
	if err != nil {
		log.Println("failed to initiate Shop while handling ShopRequest.ID: ", ShopRequest.ID)
		return err
	}

	scrappedShop.CreatedByUserID = ShopRequest.AccountID

	err = s.SaveShopToDB(scrappedShop, ShopRequest)
	if err != nil {
		return err
	}

	log.Println("starting Shop's menu scraping for ShopRequest.ID: ", ShopRequest.ID)

	scrapeMenu := s.Scraper.ScrapAllMenuItems(scrappedShop)

	err = s.UpdateShopMenuToDB(scrapeMenu, ShopRequest)
	if err != nil {
		return err
	}

	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}

	if scrapeMenu.HasSoldHistory && scrapeMenu.TotalSales > 0 {
		log.Println("Shop's selling history initiated for ShopRequest.ID: ", ShopRequest.ID)

		if err := s.Process.ExecuteUpdateSellingHistory(s, scrapeMenu, Task, ShopRequest); err != nil {
			ShopRequest.Status = "failed"
			s.Process.CreateShopRequest(ShopRequest)
			log.Println("Shop's selling history failed for ShopRequest.ID: ", ShopRequest.ID)

			return err

		}
	} else {
		ShopRequest.Status = "done"
		s.Process.CreateShopRequest(ShopRequest)
	}
	return nil
}

func (s *Shop) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {
	var dailyRevenue float64
	ScrappedSoldItems, err := s.Process.ExecuteUpdateDiscontinuedItems(s, Shop, Task, ShopRequest)
	if err != nil {
		ShopRequest.Status = "failed"
		s.Process.CreateShopRequest(ShopRequest)
		log.Println("Shop's selling history failed while initiating UpdateDiscontinuedItems for ShopRequest.ID: ", ShopRequest.ID)

		return err
	}

	if reflect.DeepEqual(ScrappedSoldItems, []models.SoldItems{}) {
		return fmt.Errorf("empty scrapped Sold data")
	}

	AllItems, err := s.Process.GetItemsByShopID(Shop.ID)
	if err != nil {
		return err
	}

	for i, ScrappedSoldItem := range ScrappedSoldItems {
		for _, item := range AllItems {
			if ScrappedSoldItem.ListingID == item.ListingID {
				ScrappedSoldItems[i].ItemID = item.ID
				dailyRevenue += item.OriginalPrice
				break
			}
		}
	}

	for i, j := 0, len(ScrappedSoldItems)-1; i < j; i, j = i+1, j-1 {
		ScrappedSoldItems[i], ScrappedSoldItems[j] = ScrappedSoldItems[j], ScrappedSoldItems[i]
	}

	result := s.DB.Create(&ScrappedSoldItems)

	if result.Error != nil {
		log.Println("Shop's selling history failed while saving to database for ShopRequest.ID: ", ShopRequest.ID)
		return result.Error
	} else if Task.UpdateSoldItems > 0 {

		now := time.Now().UTC().Truncate(24 * time.Hour)

		UpdatedSoldItemIDs := []uint{}
		for _, UpdatedSoldItem := range ScrappedSoldItems {
			UpdatedSoldItemIDs = append(UpdatedSoldItemIDs, UpdatedSoldItem.ID)
		}

		jsonArray, err := json.Marshal(UpdatedSoldItemIDs)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			return err
		}
		dailyRevenue = math.Round(dailyRevenue*100) / 100
		s.DB.Model(&models.DailyShopSales{}).Where("created_at > ?", now).Where("shop_id = ?", Shop.ID).Updates(&models.DailyShopSales{SoldItems: jsonArray, DailyRevenue: dailyRevenue})

	}

	ShopRequest.Status = "done"
	log.Printf("Shop's selling history successfully saved %v items for ShopRequest.ID: %v \n", len(ScrappedSoldItems), ShopRequest.ID)
	s.Process.CreateShopRequest(ShopRequest)

	return nil
}

func (s *Shop) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {

	SoldOutItems := []models.Item{}
	FilterSoldItems := map[uint]struct{}{}

	scrapSoldItems, NewTask := s.Scraper.ScrapSalesHistory(Shop.Name, Task)
	if !NewTask.IsScrapeFinished {
		go s.SoldItemsTask(Shop, NewTask, ShopRequest)
	}

	if reflect.DeepEqual(scrapSoldItems, []models.SoldItems{}) {
		return scrapSoldItems, nil
	}

	getAllItems, err := s.Process.GetItemsByShopID(Shop.ID)
	if err != nil {
		log.Println("UpdateDiscontinuedItems failed for ShopRequest.ID: ", ShopRequest.ID)
		log.Println(err)
		return nil, err
	}

	for i, scrapedItem := range scrapSoldItems {
		for _, item := range getAllItems {
			if scrapedItem.ListingID == item.ListingID && scrapedItem.ItemID == 0 {
				scrapSoldItems[i].ItemID = item.ID
				break
			}

		}
		if scrapSoldItems[i].ItemID == 0 {
			if _, exists := FilterSoldItems[scrapedItem.ListingID]; !exists {
				FilterSoldItems[scrapedItem.ListingID] = struct{}{}
				SoldItem := models.CreateSoldOutItem(&scrapedItem)
				SoldOutItems = append(SoldOutItems, *SoldItem)
			}
		}

	}
	isOutOfProduction := false
	for index, menu := range Shop.ShopMenu.Menu {
		if menu.Category == "Out Of Production" {
			isOutOfProduction = true
			Shop.ShopMenu.Menu[index].Amount = Shop.ShopMenu.Menu[index].Amount + len(SoldOutItems)
			Shop.ShopMenu.Menu[index].Items = append(menu.Items, SoldOutItems...)

			s.DB.Save(Shop.ShopMenu.Menu[index])
			log.Println("Out Of Production successfully updated for ShopRequest.ID: ", ShopRequest.ID)
		}
	}

	if len(SoldOutItems) != 0 && !isOutOfProduction {
		Menu := models.MenuItem{
			ShopMenuID: Shop.ShopMenu.ID,
			Category:   "Out Of Production",
			SectionID:  "0",
			Amount:     len(SoldOutItems),
			Items:      SoldOutItems,
		}

		Shop.ShopMenu.Menu = append(Shop.ShopMenu.Menu, Menu)
		s.DB.Save(Shop)
		log.Println("Out Of Production successfully created for ShopRequest.ID: ", ShopRequest.ID)

	}

	return scrapSoldItems, nil
}

func (s *Shop) FollowShop(ctx *gin.Context) {

	var shopToFollow *FollowShopRequest
	if err := ctx.ShouldBindJSON(&shopToFollow); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

	requestedShop, err := s.Process.GetShopByName(shopToFollow.FollowShopName)
	if err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "shop not found"})
			return
		}
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "error while processing the request"})
		return
	}
	utils := &utils.Utils{}
	currentAccount, err := NewUserController(s.DB, utils).GetAccountByID(currentUserUUID)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	currentAccount.ShopsFollowing = append(currentAccount.ShopsFollowing, *requestedShop)
	if err := s.DB.Save(&currentAccount).Error; err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return

	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": "following shop"})

}

func (s *Shop) UnFollowShop(ctx *gin.Context) {

	var unFollowShop *UnFollowShopRequest
	if err := ctx.ShouldBindJSON(&unFollowShop); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)

	requestedShop, err := s.Process.GetShopByName(unFollowShop.UnFollowShopName)
	if err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "shop not found"})
			return
		}
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	account := &models.Account{}
	if err := s.DB.Preload("ShopsFollowing").Where("id = ?", currentUserUUID).First(&account).Error; err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if err := s.DB.Model(&account).Association("ShopsFollowing").Delete(requestedShop); err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	if err := s.DB.Save(&account).Error; err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": "Unfollowed shop"})

}

func (pr *ShopCreators) GetShopByName(ShopName string) (shop *models.Shop, err error) {

	if err = pr.DB.Preload("Member").Preload("ShopMenu.Menu.Items").Preload("Reviews.ReviewsTopic").Where("name = ?", ShopName).First(&shop).Error; err != nil {
		log.Println("no Shop was Found ,error :", err)
		return nil, err
	}
	return
}

func (s *Shop) GetShopByID(ID uint) (shop *models.Shop, err error) {

	if err := s.DB.Preload("Member").Preload("ShopMenu.Menu").Preload("Reviews.ReviewsTopic").Where("id = ?", ID).First(&shop).Error; err != nil {
		log.Println("no Shop was Found ")

		return nil, err
	}

	shop.AvarageItemsPrice, err = s.Process.GetAvarageItemPrice(shop.ID)
	if err != nil {
		log.Println("error while calculating item avarage price")
		return nil, err
	}

	shop.Revenue, err = s.Process.ExecuteGetTotalRevenue(s, shop.ID, shop.AvarageItemsPrice)
	if err != nil {
		log.Println("error while calculating shop's revenue")
		return nil, err
	}

	return
}

func (ps *ShopCreators) GetItemsByShopID(ID uint) (items []models.Item, err error) {
	shop := &models.Shop{}
	if err := ps.DB.Preload("ShopMenu.Menu.Items").Where("id = ?", ID).First(shop).Error; err != nil {
		log.Println("no Shop was Found")
		return nil, err
	}

	for _, menu := range shop.ShopMenu.Menu {
		items = append(items, menu.Items...)
	}
	return
}

func (s *Shop) GetItemsCountByShopID(ID uint) (itemsCount, error) {
	itemCount := itemsCount{}
	items, err := s.Process.GetItemsByShopID(ID)
	if err != nil {
		log.Println("error while calculating item avarage price")
		return itemCount, err
	}
	for _, item := range items {
		if item.Available {
			itemCount.Available++
		} else {
			itemCount.OutOfProduction++
		}
	}

	return itemCount, nil
}

func (s *Shop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	listingIDs := []uint{}
	Solditems := []models.SoldItems{}

	AllItems, err := s.Process.GetItemsByShopID(ID)
	if err != nil {
		log.Println("items where not found ")
		return nil, err
	}

	for _, item := range AllItems {
		listingIDs = append(listingIDs, item.ListingID)
	}

	result := s.DB.Where("listing_id IN ?", listingIDs).Find(&Solditems)
	if result.Error != nil {
		log.Println("items where not found ")
		return nil, result.Error
	}

	soldQauntity := map[uint]int{}
	for _, SoldItem := range Solditems {
		soldQauntity[SoldItem.ItemID]++
	}

	for key, value := range soldQauntity {
		for _, item := range AllItems {
			if key == item.ID {
				SoldItemInfo := CreateSoldItemInfo(&item)
				SoldItemInfo.SoldQauntity = value
				SoldItemInfos = append(SoldItemInfos, *SoldItemInfo)
			}
		}

	}

	return
}

func (s *ShopCreators) GetAvarageItemPrice(ShopID uint) (float64, error) {
	var averagePrice float64

	if err := s.DB.Table("items").
		Joins("JOIN menu_items ON items.menu_item_id = menu_items.id").
		Joins("JOIN shop_menus ON menu_items.shop_menu_id = shop_menus.id").
		Joins("JOIN shops ON shop_menus.shop_id = shops.id").
		Where("shops.id = ? AND items.original_price > 0 ", ShopID).
		Select("AVG(items.original_price) as average_price").
		Row().Scan(&averagePrice); err != nil {

		return 0, err
	}
	averagePrice = math.Round(averagePrice*100) / 100

	return averagePrice, nil
}

func (s *Shop) GetTotalRevenue(ShopID uint, AvarageItemPrice float64) (float64, error) {
	var revenue float64

	soldItems, err := s.Process.ExecuteGetSoldItemsByShopID(s, ShopID)
	if err != nil {
		log.Println("error while calculating revenue")
		return 0, err
	}
	for _, soldItem := range soldItems {
		if soldItem.Available {
			revenue += soldItem.OriginalPrice * float64(soldItem.SoldQauntity)
		} else {
			revenue += AvarageItemPrice * float64(soldItem.SoldQauntity)
		}
	}
	revenue = math.Round(revenue*100) / 100
	return revenue, nil
}

func (s *Shop) SoldItemsTask(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {

	randTimeSet := time.Duration(rand.Intn(89-10) + 10)
	durationUntilNextTask := time.Until(time.Now().Add(randTimeSet * time.Second))

	time.AfterFunc(durationUntilNextTask, func() {
		s.Process.ExecuteUpdateSellingHistory(s, Shop, Task, ShopRequest)
	})
	return nil
}

func (pr *ShopCreators) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	if ShopRequest.AccountID == uuid.Nil {
		return errors.New("no AccountID was passed")
	}

	result := pr.DB.Save(ShopRequest)
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}

	return nil
}

func (s *Shop) ProcessStatsRequest(ctx *gin.Context, ShopID uint, Period string) error {

	year, month, day := 0, 0, 0

	switch Period {
	case "lastSevenDays":
		day = -6
	case "lastThirtyDays":
		day = -29
	case "lastThreeMonths":
		month = -3
	case "lastSixMonths":
		month = -6
	case "lastYear":
		year = -1
	default:

		return fmt.Errorf("invalid period provided")

	}
	date := time.Now().AddDate(year, month, day)
	dateMidnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	LastSevenDays, err := s.Process.ExecuteGetSellingStatsByPeriod(s, ShopID, dateMidnight)
	if err != nil {
		log.Println("error while retreiving shop selling stats ,error :", err)
		return fmt.Errorf("error while retreiving shop selling stats ,error : %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "stats": LastSevenDays})
	return nil
}

func (s *Shop) GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error) {

	dailyShopSales := []models.DailyShopSales{}
	itemIDs := []uint{}
	item := models.Item{}

	stats := make(map[string]DailySoldStats)

	result := s.DB.Where("shop_id = ? AND created_at > ?", ShopID, timePeriod).Find(&dailyShopSales)

	if result.Error != nil {
		log.Println(result.Error)
		return nil, result.Error
	}

	for _, sales := range dailyShopSales {
		items := []models.Item{}
		dateCreated := sales.CreatedAt.Format("2006-01-02")

		if len(sales.SoldItems) == 0 {
			stats[dateCreated] = DailySoldStats{
				TotalSales: sales.TotalSales,
			}
			continue
		}

		if err := json.Unmarshal(sales.SoldItems, &itemIDs); err != nil {
			fmt.Println("Error parsing sold items:", err)
			return nil, err
		}
		for _, itemID := range itemIDs {
			result := s.DB.Raw("SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = (?)", itemID).Scan(&item)
			if result.Error != nil {
				return nil, result.Error
			}
			items = append(items, item)
		}

		stats[dateCreated] = DailySoldStats{
			TotalSales: sales.TotalSales,
			Items:      items,
		}

	}

	return stats, nil
}

func (s *Shop) SaveShopToDB(scrappedShop *models.Shop, ShopRequest *models.ShopRequest) error {

	err := s.DB.Create(scrappedShop).Error
	if err != nil {
		log.Println("failed to save Shop's data while handling ShopRequest.ID: ", ShopRequest.ID)
		ShopRequest.Status = "failed"
		s.Process.CreateShopRequest(ShopRequest)
		return err
	}

	log.Println("Shop's data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

func (s *Shop) UpdateShopMenuToDB(Shop *models.Shop, ShopRequest *models.ShopRequest) error {

	err := s.DB.Save(Shop).Error

	if err != nil {
		ShopRequest.Status = "failed"
		log.Println("failed to save Shop's menu into database for ShopRequest.ID: ", ShopRequest.ID)
		s.Process.CreateShopRequest(ShopRequest)
		return err
	}

	ShopRequest.Status = "done"
	s.Process.CreateShopRequest(ShopRequest)
	log.Println("Shop's menu data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

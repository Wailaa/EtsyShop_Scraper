package controllers

import (
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shop struct {
	DB *gorm.DB
}

func NewShopController(DB *gorm.DB) *Shop {
	return &Shop{DB}
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
	SoldQauntity   int
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
	}
	return newSoldItem
}

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

	IsShop, err := s.GetShopByName(shop.ShopName)
	if err != nil && err.Error() != "record not found" {
		log.Println(err)
		ShopRequest.Status = "failed"
		s.CreateShopRequest(ShopRequest)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "internal error"})
		return
	} else if IsShop != nil {
		ShopRequest.Status = "denied"
		s.CreateShopRequest(ShopRequest)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Shop already exists"})
		return
	}

	ShopRequest.Status = "Pending"
	s.CreateShopRequest(ShopRequest)

	message := "shop request received successfully"
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": message})
	go s.CreateNewShop(ShopRequest)
}

// func (s *Shop) CreateShopTaskQueue(shopReqTask models.ShopRequest) error {

// 	result := s.DB.Create(&shopReqTask)
// 	if result.Error != nil {
// 		log.Println("error while creating tas queue for new shop request")
// 		return result.Error
// 	}
// 	return nil
// }

func (s *Shop) CreateNewShop(ShopRequest *models.ShopRequest) error {

	scrappedShop, err := scrap.ScrapShop(ShopRequest.ShopName)
	if err != nil {
		log.Println("failed to initiate Shop while handling ShopRequest.ID: ", ShopRequest.ID)
		return err
	}

	scrappedShop.CreatedByUserID = ShopRequest.AccountID

	result := s.DB.Create(scrappedShop)
	if result.Error != nil {
		log.Println("failed to save Shop's data while handling ShopRequest.ID: ", ShopRequest.ID)
		ShopRequest.Status = "failed"
		return err
	}

	log.Println("Shop's data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)

	time.Sleep(10 * time.Second)

	log.Println("starting Shop's menu scraping for ShopRequest.ID: ", ShopRequest.ID)
	scrapeMenu := scrap.ScrapAllMenuItems(scrappedShop)

	result = s.DB.Save(scrapeMenu)
	if result.Error != nil {
		ShopRequest.Status = "failed"
		log.Println("failed to save Shop's menu into database for ShopRequest.ID: ", ShopRequest.ID)
		s.CreateShopRequest(ShopRequest)
		return err
	}

	ShopRequest.Status = "done"
	s.CreateShopRequest(ShopRequest)
	log.Println("Shop's menu data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)

	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}

	if scrapeMenu.HasSoldHistory && scrapeMenu.TotalSales > 0 {
		log.Println("Shop's selling history initiated for ShopRequest.ID: ", ShopRequest.ID)
		time.Sleep(10 * time.Second)

		if err := s.UpdateSellingHistory(scrapeMenu, Task, ShopRequest); err != nil {
			ShopRequest.Status = "failed"
			s.CreateShopRequest(ShopRequest)
			log.Println("Shop's selling history failed for ShopRequest.ID: ", ShopRequest.ID)

			return err

		}
	} else {
		ShopRequest.Status = "done"
		s.CreateShopRequest(ShopRequest)
	}
	return nil
}

func (s *Shop) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {

	ScrappedSoldItems, err := s.UpdateDiscontinuedItems(Shop, Task, ShopRequest)
	if err != nil {
		ShopRequest.Status = "failed"
		s.CreateShopRequest(ShopRequest)
		log.Println("Shop's selling history failed while initiating UpdateDiscontinuedItems for ShopRequest.ID: ", ShopRequest.ID)

		return err
	}

	if reflect.DeepEqual(ScrappedSoldItems, []models.SoldItems{}) {
		return fmt.Errorf("empty scrapped Sold data")
	}

	AllItems, _ := s.GetItemsByShopID(Shop.ID)

	for i, ScrappedSoldItem := range ScrappedSoldItems {
		for _, item := range AllItems {
			if ScrappedSoldItem.ListingID == item.ListingID {
				ScrappedSoldItems[i].ItemID = item.ID
				break
			}
		}
	}

	for i, j := 0, len(ScrappedSoldItems)-1; i < j; i, j = i+1, j-1 {
		ScrappedSoldItems[i], ScrappedSoldItems[j] = ScrappedSoldItems[j], ScrappedSoldItems[i]
	}

	result := s.DB.Create(ScrappedSoldItems)
	if result.Error != nil {
		log.Println("Shop's selling history failed while saving to database for ShopRequest.ID: ", ShopRequest.ID)
		return err
	}

	ShopRequest.Status = "done"
	log.Printf("Shop's selling history successfully saved %v items for ShopRequest.ID: %v \n", len(ScrappedSoldItems), ShopRequest.ID)
	s.CreateShopRequest(ShopRequest)

	return nil
}

func (s *Shop) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {

	SoldOutItems := []models.Item{}
	FilterSoldItems := map[uint]struct{}{}

	scrapSoldItems, NewTask := scrap.ScrapSalesHistory(Shop.Name, Task)
	if !NewTask.IsScrapeFinished {
		log.Println("Task :", Task)
		go s.SoldItemsTask(Shop, NewTask, ShopRequest)
	}

	if reflect.DeepEqual(scrapSoldItems, []models.SoldItems{}) {
		return scrapSoldItems, nil
	}

	getAllItems, err := s.GetItemsByShopID(Shop.ID)
	if err != nil {
		log.Println("UpdateDiscontinuedItems failed for ShopRequest.ID: ", ShopRequest.ID)
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

	requestedShop, err := s.GetShopByName(shopToFollow.FollowShopName)
	if err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "shop not found"})
			return
		}
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	currentAccount, err := s.GetAccountByID(currentUserUUID)
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

	requestedShop, err := s.GetShopByName(unFollowShop.UnFollowShopName)
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

func (s *Shop) GetShopByName(ShopName string) (shop *models.Shop, err error) {

	if err = s.DB.Preload("Member").Preload("ShopMenu.Menu.Items").Preload("Reviews.ReviewsTopic").Where("name = ?", ShopName).First(&shop).Error; err != nil {
		log.Println("no Shop was Found ,error :", err)
		return nil, err
	}
	return
}

func (s *Shop) GetShopByID(ID uint) (shop *models.Shop, err error) {

	if err := s.DB.Preload("Member").Preload("ShopMenu.Menu.Items").Preload("Reviews.ReviewsTopic").Where("id = ?", ID).First(&shop).Error; err != nil {
		log.Println("no Shop was Found ")

		return nil, err
	}
	return
}

func (s *Shop) GetItemsByShopID(ID uint) (items []models.Item, err error) {
	shop := &models.Shop{}
	if err := s.DB.Preload("ShopMenu.Menu.Items").Where("id = ?", ID).First(shop).Error; err != nil {
		log.Println("no Shop was Found")
		return nil, err
	}

	for _, menu := range shop.ShopMenu.Menu {
		items = append(items, menu.Items...)
	}
	return
}

func (s *Shop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	listingIDs := []uint{}
	Solditems := []models.SoldItems{}

	AllItems, err := s.GetItemsByShopID(ID)
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
		return nil, err
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

func (s *Shop) SoldItemsTask(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {
	log.Println("new task is created")

	randTimeSet := time.Duration(rand.Intn(89-10) + 10)
	durationUntilNextTask := time.Until(time.Now().Add(randTimeSet * time.Second))

	time.AfterFunc(durationUntilNextTask, func() {
		s.UpdateSellingHistory(Shop, Task, ShopRequest)
	})
	return nil
}

func (s *Shop) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	if ShopRequest.AccountID == uuid.Nil {
		return nil
	}

	result := s.DB.Save(ShopRequest)
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}

	return nil
}

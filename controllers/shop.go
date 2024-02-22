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

	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
	var shop NewShopRequest

	if err := ctx.ShouldBindJSON(&shop); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	IsShop, err := s.GetShopByName(shop.ShopName)
	if err != nil && err.Error() != "record not found" {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if IsShop != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Shop exists"})
		return
	}

	scrappedShop, err := scrap.ScrapShop(shop.ShopName)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	scrappedShop.CreatedByUserID = currentUserUUID
	time.Sleep(10 * time.Second)
	secondStage := scrap.ScrapAllMenuItems(scrappedShop)

	tx := s.DB.Begin()

	result := tx.Create(secondStage)
	if result.Error != nil {
		tx.Rollback()
		log.Println(err)
		return
	}

	tx.Commit()

	Task := &models.TaskSchedule{
		IsScrapped: false,
		FirstPage:  2,
		LastPage:   0,
	}

	if secondStage.HasSoldHistory && secondStage.TotalSales > 0 {
		time.Sleep(35 * time.Second)
		if err := s.UpdateSellingHistory(secondStage, Task); err != nil {
			log.Println(err)
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to create history"})
			return

		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": secondStage})

}

func (s *Shop) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule) error {

	ScrappedSoldItems, err := s.UpdateDiscontinuedItems(Shop, Task)
	if err != nil {
		log.Println(err)

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
		log.Println(err)
		return err
	}

	return nil
}

func (s *Shop) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule) ([]models.SoldItems, error) {

	SoldOutItems := []models.Item{}
	FilterSoldItems := map[uint]struct{}{}

	scrapSoldItems, NewTask := scrap.ScrapSalesHistory(Shop.Name, Task)
	if NewTask.FirstPage != 0 {
		NewTask.IsScrapped = false
		go s.SoldItemsTask(Shop, NewTask)
	}

	if reflect.DeepEqual(scrapSoldItems, []models.SoldItems{}) {
		return nil, fmt.Errorf("empty scrapped Sold data")
	}

	getAllItems, err := s.GetItemsByShopID(Shop.ID)
	if err != nil {
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
			Shop.ShopMenu.Menu[index].Amount += len(SoldOutItems)
			Shop.ShopMenu.Menu[index].Items = append(Shop.ShopMenu.Menu[index].Items, SoldOutItems...)

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

	}
	s.DB.Save(Shop)

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

func (s *Shop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []models.ResponseSoldItemInfo, err error) {
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

func (s *Shop) SoldItemsTask(Shop *models.Shop, Task *models.TaskSchedule) error {
	log.Println("new task is created")

	randTimeSet := time.Duration(rand.Intn(89-10) + 10)
	durationUntilNextTask := time.Until(time.Now().Add(randTimeSet * time.Second))

	time.AfterFunc(durationUntilNextTask, func() {
		s.UpdateSellingHistory(Shop, Task)
	})
	return nil
}

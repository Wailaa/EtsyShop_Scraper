package controllers

import (
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
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
func (s *Shop) CreateNewShop(ctx *gin.Context) {
	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
	var shop *models.CreateNewShopReuest
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

	secondStage := scrap.ScrapAllMenuItems(scrappedShop)

	tx := s.DB.Begin()

	result := tx.Create(secondStage)
	if result.Error != nil {
		tx.Rollback()
		log.Println(err)
		return
	}

	tx.Commit()

	if err := s.UpdateSellingHistory(secondStage.Name, secondStage.ID); err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "failed to create history"})
		return

	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": secondStage})

}

func (s *Shop) UpdateSellingHistory(ShopName string, ShopID uint) error {
	scrapSoldItems := scrap.ScrapSalesHistory(ShopName)

	getAllItems, err := s.GetItemsByShopID(ShopID)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, SoldUnit := range scrapSoldItems {
		for _, item := range getAllItems {
			if SoldUnit.ListingID == item.ListingID {
				SoldUnit.ItemID = item.ID
				if err := s.DB.Create(&SoldUnit).Error; err != nil {
					log.Println(err)
					return err
				}
			}
		}
	}
	return nil
}

func (s *Shop) FollowShop(ctx *gin.Context) {

	var shopToFollow *models.FollowShopRequest
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

	var unFollowShop *models.UnFollowShopRequest
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
		items = append(items, *menu.Items...)
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
				SoldItemInfo := models.CreateSoldItemInfo(&item)
				SoldItemInfo.SoldQauntity = value
				SoldItemInfos = append(SoldItemInfos, *SoldItemInfo)
			}
		}

	}

	return
}

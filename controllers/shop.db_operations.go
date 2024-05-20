package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func (s *Shop) SaveShopToDB(scrappedShop *models.Shop, ShopRequest *models.ShopRequest) error {

	if err := s.DB.Create(scrappedShop).Error; err != nil {
		ShopRequest.Status = "failed"
		s.Process.ExecuteCreateShopRequest(s, ShopRequest)
		message := fmt.Sprintf("failed to save Shop's data while handling ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	log.Println("Shop's data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

func (s *Shop) UpdateShopMenuToDB(Shop *models.Shop, ShopRequest *models.ShopRequest) error {

	if err := s.DB.Save(Shop).Error; err != nil {
		ShopRequest.Status = "failed"
		s.Process.CreateShopRequest(ShopRequest)
		message := fmt.Sprintf("failed to save Shop's menu into database for ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	ShopRequest.Status = "done"
	s.Process.CreateShopRequest(ShopRequest)
	log.Println("Shop's menu data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

func (s *Shop) SaveSoldItemsToDB(ScrappedSoldItems []models.SoldItems) error {
	err := s.DB.Create(&ScrappedSoldItems).Error

	if err != nil {
		return utils.HandleError(err, "Shop's selling history failed while saving to database")
	}
	return nil
}

func (s *Shop) UpdateDailySales(ScrappedSoldItems []models.SoldItems, ShopID uint, dailyRevenue float64) error {

	now := utils.TruncateDate(time.Now())

	dailyRevenue = RoundToTwoDecimalDigits(dailyRevenue)

	if err := s.DB.Model(&models.DailyShopSales{}).Where("created_at > ?", now).Where("shop_id = ?", ShopID).Updates(&models.DailyShopSales{DailyRevenue: dailyRevenue}).Error; err != nil {
		return utils.HandleError(err)
	}

	return nil
}

func (s *Shop) CheckAndUpdateOutOfProdMenu(AllMenus []models.MenuItem, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) (bool, error) {
	isOutOfProduction := false
	for index, menu := range AllMenus {
		if menu.Category == "Out Of Production" {
			isOutOfProduction = true
			AllMenus[index].Amount = AllMenus[index].Amount + len(SoldOutItems)
			AllMenus[index].Items = append(menu.Items, SoldOutItems...)

			if err := s.DB.Save(&AllMenus[index]).Error; err != nil {
				return false, utils.HandleError(err)
			}
			ShopRequest.Status = "OutOfProduction Successfully updated"
			s.Process.CreateShopRequest(ShopRequest)
			log.Println("Out Of Production successfully updated for ShopRequest.ID: ", ShopRequest.ID)
			break
		}
	}
	return isOutOfProduction, nil
}

func (s *Shop) CreateOutOfProdMenu(Shop *models.Shop, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) error {
	Menu := models.MenuItem{
		ShopMenuID: Shop.ShopMenu.ID,
		Category:   "Out Of Production",
		SectionID:  "0",
		Amount:     len(SoldOutItems),
		Items:      SoldOutItems,
	}

	Shop.ShopMenu.Menu = append(Shop.ShopMenu.Menu, Menu)
	if err := s.DB.Save(Shop).Error; err != nil {
		return utils.HandleError(err)
	}

	log.Println("Out Of Production successfully created for ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

func (s *Shop) GetShopByID(ID uint) (shop *models.Shop, err error) {

	if err := s.DB.Preload("Member").Preload("ShopMenu.Menu").Preload("Reviews.ReviewsTopic").Where("id = ?", ID).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ")

	}

	shop.AverageItemsPrice, err = s.Process.ExecuteGetAverageItemPrice(s, shop.ID)
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

func (s *Shop) GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error) {

	soldItems, err := s.Process.ExecuteGetSoldItemsByShopID(s, ShopID)
	if err != nil {
		return 0, utils.HandleError(err, "error while calculating revenue")
	}
	revenue := CalculateTotalRevenue(soldItems, AverageItemPrice)
	return revenue, nil
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

func (s *Shop) GetAverageItemPrice(ShopID uint) (float64, error) {
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

func (s *Shop) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	if ShopRequest.AccountID == uuid.Nil {
		err := errors.New("no AccountID was passed")
		return utils.HandleError(err)
	}

	if err := s.DB.Save(ShopRequest).Error; err != nil {
		return utils.HandleError(err)
	}

	return nil
}

package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"fmt"
	"log"
	"time"
)

func (s *Shop) SaveShopToDB(scrappedShop *models.Shop, ShopRequest *models.ShopRequest) error {

	if err := s.DB.Create(scrappedShop).Error; err != nil {
		ShopRequest.Status = "failed"
		s.Process.CreateShopRequest(ShopRequest)
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

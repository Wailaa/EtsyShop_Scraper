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

	if err := s.Shop.CreateShop(scrappedShop); err != nil {
		ShopRequest.Status = "failed"
		s.Operations.CreateShopRequest(ShopRequest)
		message := fmt.Sprintf("failed to save Shop's data while handling ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	log.Println("Shop's data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

func (s *Shop) UpdateShopMenuToDB(Shop *models.Shop, ShopRequest *models.ShopRequest) error {

	if err := s.Shop.SaveShop(Shop); err != nil {
		ShopRequest.Status = "failed"
		s.Operations.CreateShopRequest(ShopRequest)
		message := fmt.Sprintf("failed to save Shop's menu into database for ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	ShopRequest.Status = "done"
	s.Operations.CreateShopRequest(ShopRequest)
	log.Println("Shop's menu data saved successfully while handling ShopRequest.ID: ", ShopRequest.ID)
	return nil
}


func (s *Shop) UpdateDailySales(ScrappedSoldItems []models.SoldItems, ShopID uint, dailyRevenue float64) error {

	now := utils.TruncateDate(time.Now())

	dailyRevenue = utils.RoundToTwoDecimalDigits(dailyRevenue)

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
			s.Operations.CreateShopRequest(ShopRequest)
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
	if err := s.Shop.SaveShop(Shop); err != nil {
		return utils.HandleError(err)
	}

	log.Println("Out Of Production successfully created for ShopRequest.ID: ", ShopRequest.ID)
	return nil
}

func (s *Shop) GetShopByID(ID uint) (shop *models.Shop, err error) {

	if err := s.DB.Preload("Member").Preload("ShopMenu.Menu").Preload("Reviews.ReviewsTopic").Where("id = ?", ID).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ")

	}

	shop.AverageItemsPrice, err = s.Operations.GetAverageItemPrice(shop.ID)
	if err != nil {
		return nil, utils.HandleError(err, "error while calculating item avearage price")
	}

	if !shop.HasSoldHistory {
		shop.Revenue = shop.AverageItemsPrice * float64(shop.TotalSales)
		return
	}

	shop.Revenue, err = s.Operations.GetTotalRevenue(shop.ID, shop.AverageItemsPrice)
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

	stats, err := s.Operations.CreateSoldStats(dailyShopSales)
	if err != nil {
		return nil, utils.HandleError(err)
	}
	return stats, nil
}

func (s *Shop) GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error) {

	soldItems, err := s.Operations.GetSoldItemsByShopID(ShopID)
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

	currentAccount, err := s.User.GetAccountByID(userID)
	if err != nil {
		return utils.HandleError(err)
	}

	currentAccount.ShopsFollowing = append(currentAccount.ShopsFollowing, *requestedShop)
	if err := s.User.SaveAccount(currentAccount); err != nil {
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
	averagePrice = utils.RoundToTwoDecimalDigits(averagePrice)

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

func (s *Shop) GetItemsByShopID(ID uint) (items []models.Item, err error) {
	shop := &models.Shop{}
	if err := s.DB.Preload("ShopMenu.Menu.Items").Where("id = ?", ID).First(shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found")
	}

	for _, menu := range shop.ShopMenu.Menu {
		items = append(items, menu.Items...)
	}
	return
}

func (s *Shop) GetShopByName(ShopName string) (shop *models.Shop, err error) {

	if err = s.DB.Preload("Member").Preload("ShopMenu.Menu.Items").Preload("Reviews.ReviewsTopic").Where("name = ?", ShopName).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ,error")
	}
	return
}

func (s *Shop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	listingIDs := []uint{}
	Solditems := []models.SoldItems{}

	AllItems, err := s.Operations.GetItemsByShopID(ID)
	if err != nil {
		return nil, utils.HandleError(err, "items here not found ")
	}

	for _, item := range AllItems {
		listingIDs = append(listingIDs, item.ListingID)
	}

	if err := s.DB.Where("listing_id IN ?", listingIDs).Find(&Solditems).Error; err != nil {
		return nil, utils.HandleError(err, "items were not found ")
	}

	soldQuantity := map[uint]int{}
	for _, SoldItem := range Solditems {
		soldQuantity[SoldItem.ItemID]++
	}

	for key, value := range soldQuantity {
		for _, item := range AllItems {
			if key == item.ID {
				SoldItemInfo := CreateSoldItemInfo(&item)
				SoldItemInfo.SoldQuantity = value
				SoldItemInfos = append(SoldItemInfos, *SoldItemInfo)
			}
		}

	}

	return
}

func (s *Shop) GetItemsBySoldItems(SoldItems []models.SoldItems) ([]models.Item, error) {

	item := models.Item{}

	items := []models.Item{}

	for _, soldItem := range SoldItems {
		if err := s.DB.Raw("SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = (?)", soldItem.ID).Scan(&item).Error; err != nil {
			return nil, utils.HandleError(err, "error parsing sold items")
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Shop) GetSoldItemsInRange(fromDate time.Time, ShopID uint) ([]models.SoldItems, error) {
	soldItems := []models.SoldItems{}

	tillDate := fromDate.Add(24 * time.Hour)

	if err := s.DB.Table("shops").
		Select("sold_items.*").
		Joins("JOIN shop_menus ON shops.id = shop_menus.shop_id").
		Joins("JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id").
		Joins("JOIN items ON menu_items.id = items.menu_item_id").
		Joins("JOIN sold_items ON items.id = sold_items.item_id").
		Where("shops.id = ? AND sold_items.created_at BETWEEN ? AND ?", ShopID, fromDate, tillDate).
		Find(&soldItems).Error; err != nil {
		return nil, utils.HandleError(err)
	}
	return soldItems, nil
}

func (s *Shop) GetItemsCountByShopID(ID uint) (itemsCount, error) {
	itemCount := itemsCount{}

	items, err := s.Operations.GetItemsByShopID(ID)
	if err != nil {
		return itemCount, utils.HandleError(err, "error while calculating item average price")
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

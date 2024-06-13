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

func (s *Shop) CheckAndUpdateOutOfProdMenu(AllMenus []models.MenuItem, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) (bool, error) {
	isOutOfProduction := false
	for index, menu := range AllMenus {
		if menu.Category == "Out Of Production" {
			isOutOfProduction = true
			AllMenus[index].Amount = AllMenus[index].Amount + len(SoldOutItems)
			AllMenus[index].Items = append(menu.Items, SoldOutItems...)

			if err := s.Shop.SaveMenu(AllMenus[index]); err != nil {
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

func (s *Shop) GetShopByID(ID uint) (*models.Shop, error) {

	shop, err := s.Shop.FetchShopByID(ID)
	if err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ")

	}

	shop.AverageItemsPrice, err = s.Shop.GetAverageItemPrice(shop.ID)
	if err != nil {
		return nil, utils.HandleError(err, "error while calculating item avearage price")
	}

	if !shop.HasSoldHistory {
		shop.Revenue = shop.AverageItemsPrice * float64(shop.TotalSales)
		return shop, nil
	}

	shop.Revenue, err = s.Operations.GetTotalRevenue(shop.ID, shop.AverageItemsPrice)
	if err != nil {
		return nil, utils.HandleError(err, "error while calculating shop's revenue")
	}

	return shop, nil
}

func (s *Shop) GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]DailySoldStats, error) {

	dailyShopSales, err := s.Shop.FetchStatsByPeriod(ShopID, timePeriod)

	if err != nil {
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

func (s *Shop) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	if ShopRequest.AccountID == uuid.Nil {
		err := errors.New("no AccountID was passed")
		return utils.HandleError(err)
	}

	if err := s.Shop.SaveShopRequestToDB(ShopRequest); err != nil {
		return utils.HandleError(err)
	}

	return nil
}

func (s *Shop) GetItemsByShopID(ID uint) (items []models.Item, err error) {
	shop, err := s.Shop.GetShopWithItemsByShopID(ID)
	if err != nil {
		return nil, utils.HandleError(err, "no Shop was Found")
	}

	for _, menu := range shop.ShopMenu.Menu {
		items = append(items, menu.Items...)
	}
	return
}

func (s *Shop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	listingIDs := []uint{}

	AllItems, err := s.Operations.GetItemsByShopID(ID)
	if err != nil {
		return nil, utils.HandleError(err, "items here not found ")
	}

	for _, item := range AllItems {
		listingIDs = append(listingIDs, item.ListingID)
	}

	Solditems, err := s.Shop.FetchSoldItemsByListingID(listingIDs)
	if err != nil {
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

	items := []models.Item{}

	for _, soldItem := range SoldItems {
		item, err := s.Shop.FetchItemsBySoldItems(soldItem.ID)
		if err != nil {
			return nil, utils.HandleError(err, "error parsing sold items")
		}
		items = append(items, item)
	}

	return items, nil
}



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

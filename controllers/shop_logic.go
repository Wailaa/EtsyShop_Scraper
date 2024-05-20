package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"fmt"
	"log"
	"math/rand"
	"time"
)

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
			s.Process.ExecuteCreateShopRequest(s, ShopRequest)
			message := fmt.Sprintf("Shop's selling history failed for ShopRequest.ID: %v", ShopRequest.ID)
			return utils.HandleError(err, message)

		}
	} else {
		ShopRequest.Status = "done"
		s.Process.ExecuteCreateShopRequest(s, ShopRequest)
	}
	return nil
}

func (s *Shop) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {

	ScrappedSoldItems, err := s.Process.ExecuteUpdateDiscontinuedItems(s, Shop, Task, ShopRequest)
	if err != nil {
		ShopRequest.Status = "failed"
		s.Process.ExecuteCreateShopRequest(s, ShopRequest)

		message := fmt.Sprintf("Shop's selling history failed while initiating UpdateDiscontinuedItems for ShopRequest.ID: %v", ShopRequest.ID)
		return utils.HandleError(err, message)
	}

	if len(ScrappedSoldItems) == 0 {
		err := fmt.Errorf("empty scrapped Sold data")
		return utils.HandleError(err)
	}

	AllItems, err := s.Operations.GetItemsByShopID(Shop.ID)
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
	s.Process.ExecuteCreateShopRequest(s, ShopRequest)

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

	getAllItems, err := s.Process.ExecuteGetItemsByShopID(s, Shop.ID)
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

func (s *Shop) SoldItemsTask(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {

	randTimeSet := time.Duration(rand.Intn(79) + 10)
	durationUntilNextTask := time.Until(time.Now().Add(randTimeSet * time.Second))

	time.AfterFunc(durationUntilNextTask, func() {
		s.Process.ExecuteUpdateSellingHistory(s, Shop, Task, ShopRequest)
	})
	return nil
}

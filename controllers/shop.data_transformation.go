package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"log"
	"math"
	"time"
)

func (s *Shop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []ResponseSoldItemInfo, err error) {
	listingIDs := []uint{}
	Solditems := []models.SoldItems{}

	AllItems, err := s.Process.GetItemsByShopID(ID)
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

func (s *Shop) CreateSoldStats(dailyShopSales []models.DailyShopSales) (map[string]DailySoldStats, error) {
	stats := make(map[string]DailySoldStats)

	for _, sales := range dailyShopSales {

		day := utils.TruncateDate(sales.CreatedAt)

		soldItems, err := s.GetSoldItemsInRange(day, sales.ShopID)
		if err != nil {
			log.Println(err)
			return nil, utils.HandleError(err)
		}

		dateCreated := sales.CreatedAt.Format("2006-01-02")
		if len(soldItems) == 0 {
			stats[dateCreated] = DailySoldStats{
				TotalSales:   sales.TotalSales,
				DailyRevenue: sales.DailyRevenue,
			}
			continue
		}
		items, err := s.GetItemsBySoldItems(soldItems)
		if err != nil {
			return nil, utils.HandleError(err)
		}

		stats[dateCreated] = DailySoldStats{
			TotalSales:   sales.TotalSales,
			DailyRevenue: sales.DailyRevenue,
			Items:        items,
		}

	}

	return stats, nil
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

	items, err := s.Process.GetItemsByShopID(ID)
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

func CalculateTotalRevenue(soldItems []ResponseSoldItemInfo, AverageItemPrice float64) float64 {
	var revenue float64
	var ItemPrice float64

	for _, soldItem := range soldItems {
		if soldItem.OriginalPrice > 0 {
			ItemPrice = soldItem.OriginalPrice
		} else {
			ItemPrice = AverageItemPrice
		}
		revenue += ItemPrice * float64(soldItem.SoldQuantity)
	}
	revenue = RoundToTwoDecimalDigits(revenue)
	return revenue
}
func RoundToTwoDecimalDigits(value float64) float64 {
	return math.Round(value*100) / 100
}

func FilterSoldOutItems(scrapSoldItems []models.SoldItems, existingItems []models.Item, FilterSoldItems map[uint]struct{}) []models.Item {
	SoldOutItems := []models.Item{}

	for i, scrapedItem := range scrapSoldItems {
		for _, item := range existingItems {
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
	return SoldOutItems
}

func PopulateItemIDsFromListings(ScrappedSoldItems []models.SoldItems, AllItems []models.Item) ([]models.SoldItems, float64) {
	var dailyRevenue float64

	for i, ScrappedSoldItem := range ScrappedSoldItems {
		for _, item := range AllItems {
			if ScrappedSoldItem.ListingID == item.ListingID {
				ScrappedSoldItems[i].ItemID = item.ID
				dailyRevenue += item.OriginalPrice
				break
			}
		}
	}
	return ScrappedSoldItems, dailyRevenue
}

func ReverseSoldItems(ScrappedSoldItems []models.SoldItems) []models.SoldItems {
	for i, j := 0, len(ScrappedSoldItems)-1; i < j; i, j = i+1, j-1 {
		ScrappedSoldItems[i], ScrappedSoldItems[j] = ScrappedSoldItems[j], ScrappedSoldItems[i]
	}
	return ScrappedSoldItems
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

package controllers

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
)

func (s *Shop) CreateSoldStats(dailyShopSales []models.DailyShopSales) (map[string]DailySoldStats, error) {
	stats := make(map[string]DailySoldStats)

	for _, sales := range dailyShopSales {

		day := utils.TruncateDate(sales.CreatedAt)

		soldItems, err := s.Shop.GetSoldItemsInRange(day, sales.ShopID)
		if err != nil {
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
		items, err := s.Operations.GetItemsBySoldItems(soldItems)
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
	revenue = utils.RoundToTwoDecimalDigits(revenue)
	return revenue
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

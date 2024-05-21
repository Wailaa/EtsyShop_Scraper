package scrap

import (
	"log"

	"github.com/gocolly/colly/v2"

	"EtsyScraper/collector"
	"EtsyScraper/models"
	"EtsyScraper/utils"
)

type ScrapeUpdateProcess interface {
	CheckForUpdates(Shop string, needUpdateItems bool) (*models.Shop, error)
	ScrapAllMenuItems(shop *models.Shop) *models.Shop
	ScrapShop(shopName string) (*models.Shop, error)
	ScrapSalesHistory(ShopName string, Task *models.TaskSchedule) ([]models.SoldItems, *models.TaskSchedule)
}
type Scraper struct {
}

func (sc *Scraper) CheckForUpdates(Shop string, needUpdateItems bool) (*models.Shop, error) {
	UpdatedShop := &models.Shop{}

	shopLink := Config.ScrapShopURL

	c := collector.NewCollyCollector().C
	c.AllowURLRevisit = true

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 404 {
			r.Request.Abort()
			log.Println("shop was not found. error 404 was returned")
		} else {
			failedURL := "https://" + r.Request.URL.Host + r.Request.URL.RequestURI()

			MaxSeconds := 89
			utils.SetSleep(MaxSeconds)

			c.Visit(failedURL)
		}
	})
	UpdatedShop.Name = Shop

	if err := scrapShopTotalSales(c, UpdatedShop); err != nil {
		return nil, utils.HandleError(err)

	}

	if err := scrapShopAdmirers(c, UpdatedShop); err != nil {
		return nil, utils.HandleError(err)

	}

	if err := scrapShopvacation(c, UpdatedShop); err != nil {
		return nil, utils.HandleError(err)

	}
	if needUpdateItems {
		if err := scrapShopMenu(c, UpdatedShop); err != nil {
			return nil, utils.HandleError(err)
		}
	}

	c.Visit(shopLink + Shop)
	c.Wait()

	return UpdatedShop, nil
}

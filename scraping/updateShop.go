package scrap

import (
	"EtsyScraper/collector"
	"EtsyScraper/models"
	"log"
	"math/rand"
	"time"

	"github.com/gocolly/colly/v2"
)

type ScrapeUpdateProcess interface {
	CheckForUpdates(Shop string, needUpdateItems bool) (*models.Shop, error)
	ScrapAllMenuItems(shop *models.Shop) *models.Shop
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

			randTimeSet := time.Duration(rand.Intn(89-10) + 10)
			time.Sleep(randTimeSet * time.Second)

			c.Visit(failedURL)
		}
	})
	UpdatedShop.Name = Shop

	if err := scrapShopTotalSales(c, UpdatedShop); err != nil {
		return nil, err
	}

	if err := scrapShopAdmirers(c, UpdatedShop); err != nil {
		return nil, err
	}

	if err := scrapShopvacation(c, UpdatedShop); err != nil {
		return nil, err
	}
	if needUpdateItems {
		if err := scrapShopMenu(c, UpdatedShop); err != nil {
			return nil, err
		}
	}

	c.Visit(shopLink + Shop)
	c.Wait()

	return UpdatedShop, nil
}

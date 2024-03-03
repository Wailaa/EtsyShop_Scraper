package scrap

import (
	"EtsyScraper/collector"
	"EtsyScraper/models"
	"log"
	"math/rand"
	"time"

	"github.com/gocolly/colly/v2"
)

func CheckForUpdates(Shop string) (*models.Shop, error) {
	UpdatedShop := &models.Shop{}

	shopLink := config.ScrapShopURL

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

	// if err := scrapShopMenu(c, newShop); err != nil {
	// 	return nil, err
	// }

	c.Visit(shopLink + Shop)
	c.Wait()

	return UpdatedShop, nil
}

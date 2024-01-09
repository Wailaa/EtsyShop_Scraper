package scrap

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

var pagination = []string{}

func ScrapSalesHistory(ShopName string) []models.SoldItems {

	c := colly.NewCollector(
		colly.ParseHTTPErrorResponse(),
		colly.MaxDepth(5),
	)

	c.SetProxy(config.ProxyHostURL)

	userAgent := utils.CreateUserAgent()

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})

	c.UserAgent = userAgent

	c.Limit(&colly.LimitRule{
		Delay:       5 * time.Second,
		RandomDelay: 5 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL: ", r.Request.URL, " failed with response: ", r, "\nError: ", err)
	})

	Items := scrapSoldItems(c)
	scrapSoldItemPages(c, ShopName)

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("done scrapping sales history")
		pageToScrap := ""
		if len(pagination) != 0 {
			pageToScrap = pagination[0]
			pagination = pagination[1:]
		}
		c.Visit(pageToScrap)
	})

	c.Visit(Shoplink + ShopName + "/sold")
	c.Wait()

	return *Items
}

func scrapSoldItems(c *colly.Collector) *[]models.SoldItems {
	TotalItemSold := &[]models.SoldItems{}

	c.OnHTML("div#content", func(e *colly.HTMLElement) {
		itemsSold := models.SoldItems{}

		e.ForEach("div[data-shop-id]", func(i int, h *colly.HTMLElement) {

			ListingID := h.Attr("data-listing-id")
			ListingIDToUint64, err := strconv.ParseUint(ListingID, 10, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}
			ListingIDToUint := uint(ListingIDToUint64)
			itemsSold.ListingID = ListingIDToUint

			itemsSold.DataShopID = h.Attr("data-shop-id")

			divID := "h3#listing-title-" + ListingID
			itemsSold.Name = h.ChildText(divID)

			itemsSold.ItemLink = h.ChildAttr("a.listing-link", "href")

			*TotalItemSold = append(*TotalItemSold, itemsSold)

		})

	})

	return TotalItemSold
}

func scrapSoldItemPages(c *colly.Collector, ShopName string) {
	var onHTMLExecuted bool
	var onHTMLMutex sync.Mutex

	c.OnHTML("div#content", func(h *colly.HTMLElement) {
		isPagination := false
		onHTMLMutex.Lock()
		defer onHTMLMutex.Unlock()

		if !onHTMLExecuted {
			onHTMLExecuted = true

			SoldPages := []string{}

			h.ForEach("li", func(i int, k *colly.HTMLElement) {
				page := k.ChildAttr("a", "data-page")
				SoldPages = append(SoldPages, page)
				isPagination = true
			})
			if isPagination {
				lastpage := SoldPages[len(SoldPages)-2]
				lastPageInt, err := strconv.Atoi(lastpage)

				if err != nil {
					log.Println(err.Error())
					return
				}

				i := 2
				for i <= lastPageInt {
					link := Shoplink + ShopName + "/sold?ref=pagination&page="
					Param := fmt.Sprint(i)
					link += Param

					pagination = append(pagination, link)
					i++
				}
			}
		}
	})

}

package scrap

import (
	"EtsyScraper/models"
	"crypto/tls"
	"net/http"

	"EtsyScraper/utils"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/imroc/req/v3"
)

var pagination = []string{}

func ScrapSalesHistory(ShopName string, Task *models.TaskSchedule) ([]models.SoldItems, *models.TaskSchedule) {

	Chrome := req.DefaultClient().ImpersonateChrome()

	c := colly.NewCollector(colly.AllowURLRevisit())

	c.UserAgent = utils.GetRandomUserAgent()

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	})

	c.SetProxy(config.ProxyHostURL)

	c.SetClient(&http.Client{
		Transport: Chrome.Transport,
	})

	extensions.Referer(c)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       3 * time.Second,
		RandomDelay: 3 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("-----------------------------")
		fmt.Println("Visiting", r.URL)
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9,de;q=0.7")
		r.Headers.Set("Accept", "test/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.7")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		for key, value := range *r.Headers {
			fmt.Printf("%s: %s\n", key, value)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("-----------------------------")
		fmt.Println("Responce on Scraping a Sold Items")
		fmt.Println(r.StatusCode)
		if r.StatusCode != 200 {
			for key, value := range *r.Headers {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL: ", r.Request.URL, "\nProxy handled the request", r.Request.ProxyURL, "\nfailed with response: ", r.Body, "\nError: ", err)

		for key, value := range *r.Headers {
			fmt.Printf("%s: %s\n", key, value)
		}
		c.WithTransport(&http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		})

		c.SetProxy(config.ProxyHostURL)

		c.UserAgent = utils.GetRandomUserAgent()

		c.SetClient(&http.Client{
			Transport: Chrome.Transport,
		})

		failedURL := "https://" + r.Request.URL.Host + r.Request.URL.RequestURI()

		time.Sleep(50 * time.Second)

		r.Request.Visit(failedURL)
	})

	Items := scrapSoldItems(c)
	Task = scrapSoldItemPages(c, ShopName, Task)

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("done scrapping sales history")
		pageToScrap := ""
		if len(pagination) != 0 {
			pageToScrap = pagination[0]
			pagination = pagination[1:]
		}
		c.Visit(pageToScrap)
	})

	if Task.FirstPage > 2 {
		pageString := fmt.Sprint(Task.FirstPage)
		c.Visit(Shoplink + ShopName + "/sold?ref=pagination&page=" + pageString)
	} else {
		c.Visit(Shoplink + ShopName + "/sold")
	}
	c.Wait()

	return *Items, Task
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

func scrapSoldItemPages(c *colly.Collector, ShopName string, Task *models.TaskSchedule) *models.TaskSchedule {
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

				loopStart := Task.FirstPage
				if Task.FirstPage != 2 {
					loopStart = Task.FirstPage + 1
				}
				loopEnds := Task.LastPage

				if lastPageInt > Task.LastPage {
					Task.FirstPage = Task.LastPage + 1
					if lastPageInt-Task.LastPage >= config.MaxPageLimit {
						Task.LastPage += config.MaxPageLimit
					} else {
						Task.LastPage = lastPageInt
					}
				} else {
					loopEnds = lastPageInt
					Task.FirstPage = 0
					Task.LastPage = 0
				}

				for loopStart <= loopEnds {
					link := Shoplink + ShopName + "/sold?ref=pagination&page="
					Param := fmt.Sprint(loopStart)
					link += Param

					pagination = append(pagination, link)
					loopStart++
				}

			}
		}

	})
	fmt.Println("modified Task", Task)
	return Task
}

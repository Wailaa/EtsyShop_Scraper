package scrap

import (
	"EtsyScraper/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type PagesToScrap struct {
	pagesCount int
	scrapURLs  []string
}

var pages = &PagesToScrap{}

var ListingIdCount = map[uint]int{}

func ScrapAllMenuItems(shop *models.Shop) *models.Shop {

	newModifiedMenuItem := []models.MenuItem{}
	AllItemCategory := models.MenuItem{}
	UnCategorizedItems := []models.Item{}

	for _, Menu := range shop.ShopMenu.Menu {

		if Menu.Category != "On sale" {

			ModifiedMenu := scrapMenuItems(&Menu)
			if Menu.Category == "All" {
				AllItemCategory = *ModifiedMenu
			} else {
				newModifiedMenuItem = append(newModifiedMenuItem, *ModifiedMenu)
			}
		}
	}

	if len(shop.ShopMenu.Menu) > 1 {
		for ListingID, Amount := range ListingIdCount {
			if Amount == 1 {
				for _, item := range *AllItemCategory.Items {
					if item.ListingID == ListingID {
						UnCategorizedItems = append(UnCategorizedItems, item)
					}
				}
			}
		}
		if len(UnCategorizedItems) > 0 {
			UnCategorizedMenu := models.MenuItem{
				ShopMenuID: AllItemCategory.ShopMenuID,
				Category:   "UnCategorized",
				SectionID:  AllItemCategory.SectionID,
				Link:       AllItemCategory.Link,
				Amount:     len(UnCategorizedItems),
				Items:      &UnCategorizedItems,
			}

			newModifiedMenuItem = append(newModifiedMenuItem, UnCategorizedMenu)
		}

		AllItemCategory.Items = &[]models.Item{}
	}
	newModifiedMenuItem = append(newModifiedMenuItem, AllItemCategory)
	shop.ShopMenu.Menu = newModifiedMenuItem
	return shop
}

func scrapMenuItems(Menu *models.MenuItem) *models.MenuItem {

	c := colly.NewCollector()

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.SetProxy(config.ProxyHostURL)

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})

	c.Limit(&colly.LimitRule{
		Delay:       5 * time.Second,
		RandomDelay: 5 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)

		for key, value := range *r.Headers {
			fmt.Printf("%s: %s\n", key, value)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("-----------------------------")
		fmt.Println("Responce on Scraping a Shop Items")
		fmt.Println(r.StatusCode)
		if r.StatusCode != 200 {
			for key, value := range *r.Headers {
				fmt.Printf("%s: %s\n", key, value)
			}
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL: ", r.Request.URL, " failed with response: ", r, "\nError: ", err)

	})
	items := scrapShopItems(c, Menu)
	Menu.Items = items

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Done scraping")
		pageToScrap := ""
		if len(pages.scrapURLs) != 0 {
			pageToScrap = pages.scrapURLs[0]
			pages.scrapURLs = pages.scrapURLs[1:]

			c.Visit(pageToScrap)

		}

	})

	scrapNextItemPage(c, Menu)

	c.Visit(Menu.Link + "&sort_order=price_desc")
	c.Wait()

	return Menu
}

func scrapNextItemPage(c *colly.Collector, shopMenu *models.MenuItem) {

	pages.scrapURLs = []string{}
	lastpage := ""

	c.OnHTML(`div[data-item-pagination]`, func(h *colly.HTMLElement) {
		h.ForEachWithBreak("nav", func(i int, g *colly.HTMLElement) bool {
			justaslice := []string{}
			if i == 1 {
				g.ForEach("li", func(i int, k *colly.HTMLElement) {
					page := k.ChildAttr("a", "data-page")

					justaslice = append(justaslice, page)

				})
				lastpage = justaslice[len(justaslice)-2]
				lastPageInt, _ := strconv.Atoi(lastpage)
				pages.pagesCount = lastPageInt
				return false
			}
			return true
		})
		splitLink := strings.Split(shopMenu.Link, "?")

		link := splitLink[0]

		i := 2
		for i <= pages.pagesCount {
			Param := fmt.Sprint("?ref=items-pagination&page=", i, "&section_id=", shopMenu.SectionID, "&sort_order=price_desc")
			link += Param
			pages.scrapURLs = append(pages.scrapURLs, link)

			link = splitLink[0]

			i++
		}

	})

}

func scrapShopItems(c *colly.Collector, shopMenu *models.MenuItem) *[]models.Item {
	testingItems := &[]models.Item{}

	c.OnHTML(`div[data-appears-component-name="shop_home_listing_grid"]`, func(e *colly.HTMLElement) {

		e.ForEach("div.js-merch-stash-check-listing", func(i int, h *colly.HTMLElement) {

			newItem := models.Item{}

			ListingID := h.Attr("data-listing-id")
			ListingIDToUint64, err := strconv.ParseUint(ListingID, 10, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}
			newItem.ListingID = uint(ListingIDToUint64)

			ListingIdCount[newItem.ListingID]++

			newItem.DataShopID = h.Attr("data-shop-id")
			newItem.MenuItemID = shopMenu.ID

			divID := "h3#listing-title-" + ListingID
			newItem.Name = h.ChildText(divID)

			OriginalPrice := h.ChildText("span.currency-value")
			SalesPrice := "-1"
			h.ForEachWithBreak("p.search-collage-promotion-price", func(i int, g *colly.HTMLElement) bool {
				SalesPrice = h.DOM.Find("span.currency-value").Eq(0).Text()
				OriginalPrice = g.ChildText("span.currency-value")

				return false
			})

			SalesPriceToFloat, err := strconv.ParseFloat(SalesPrice, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}
			OriginalPricetoFloat, err := strconv.ParseFloat(OriginalPrice, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}

			newItem.OriginalPrice = OriginalPricetoFloat

			newItem.SalePrice = SalesPriceToFloat

			newItem.CurrencySymbol = h.DOM.Find("span.currency-symbol").Eq(0).Text()

			getDiscoutPrice := h.DOM.Find("p.search-collage-promotion-price").Find("span").Last().Text()
			getDiscoutPrice = strings.TrimSpace(getDiscoutPrice)
			newItem.DiscoutPercent = getDiscoutPrice

			newItem.ItemLink = h.ChildAttr("a.listing-link", "href")
			newItem.Available = true

			*testingItems = append(*testingItems, newItem)
		})

	})

	return testingItems
}

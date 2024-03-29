package scrap

import (
	"EtsyScraper/collector"
	"EtsyScraper/models"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

var SectionIdPages = map[string]struct{}{}
var ListingIdCount = map[uint]int{}

func ScrapAllMenuItems(shop *models.Shop) *models.Shop {
	HasSalesCategory := false
	AllItemCategoryIndex := 0
	UnCategorizedItems := []models.Item{}

	c := collector.NewCollyCollector().C
	c.AllowURLRevisit = true

	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	backUpQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	c.OnError(func(r *colly.Response, err error) {
		failedURL := r.Request.URL.String()
		log.Println("failed url is :", failedURL)

		randTimeSet := time.Duration(rand.Intn(89-10) + 10)
		time.Sleep(randTimeSet * time.Second)

		backUpQueue.AddURL(failedURL)
		log.Println("Url is added to queue :", failedURL)

	})

	for index, Menu := range shop.ShopMenu.Menu {

		if !CheckCategoryName(Menu.Category) {
			OriginalQueue.AddURL(Menu.Link + "&sort_order=price_desc")
		}
		if Menu.Category == "On sale" {
			HasSalesCategory = true
		}
		if Menu.Category == "All" {
			AllItemCategoryIndex = index
		}
	}

	scrapShopItems(c, shop)
	scrapNextItemPage(c, OriginalQueue)

	OriginalQueue.Run(c)
	c.Wait()

	backUpQueue.Run(c)
	c.Wait()

	if (len(shop.ShopMenu.Menu) > 1 && !HasSalesCategory) || (len(shop.ShopMenu.Menu) > 2 && HasSalesCategory) {
		for ListingID, Amount := range ListingIdCount {
			if Amount == 1 {
				for _, item := range shop.ShopMenu.Menu[AllItemCategoryIndex].Items {
					if item.ListingID == ListingID {
						UnCategorizedItems = append(UnCategorizedItems, item)
					}
				}
			}
		}
		if len(UnCategorizedItems) > 0 {
			UnCategorizedMenu := models.MenuItem{
				ShopMenuID: shop.ShopMenu.Menu[AllItemCategoryIndex].ShopMenuID,
				Category:   "UnCategorized",
				SectionID:  shop.ShopMenu.Menu[AllItemCategoryIndex].SectionID,
				Link:       shop.ShopMenu.Menu[AllItemCategoryIndex].Link,
				Amount:     len(UnCategorizedItems),
				Items:      UnCategorizedItems,
			}

			shop.ShopMenu.Menu = append(shop.ShopMenu.Menu, UnCategorizedMenu)
		}

		shop.ShopMenu.Menu[AllItemCategoryIndex].Items = []models.Item{}
	}
	SectionIdPages = make(map[string]struct{})
	return shop
}

func scrapNextItemPage(c *colly.Collector, q *queue.Queue) {

	c.OnHTML(`div[data-item-pagination]`, func(h *colly.HTMLElement) {
		CurrentQueueURL := "https://" + h.Request.URL.Host + h.Request.URL.RequestURI()
		link := strings.Split(CurrentQueueURL, "?")[0]
		lastpage := ""
		pagesCount := 0

		h.ForEachWithBreak("nav", func(i int, g *colly.HTMLElement) bool {
			justaslice := []string{}
			if i == 1 {
				g.ForEach("li", func(i int, k *colly.HTMLElement) {
					page := k.ChildAttr("a", "data-page")

					justaslice = append(justaslice, page)

				})
				lastpage = justaslice[len(justaslice)-2]
				lastPageInt, _ := strconv.Atoi(lastpage)
				pagesCount = lastPageInt
				return false
			}
			return true
		})
		SectionID := GetSectionID(CurrentQueueURL)

		if _, ok := SectionIdPages[SectionID]; !ok {
			for i := 2; i <= pagesCount; i++ {
				SectionIdPages[SectionID] = struct{}{}

				QueueURL := fmt.Sprint(link, "?ref=items-pagination&page=", i, "&section_id=", SectionID, "&sort_order=price_desc")

				q.AddURL(QueueURL)

			}
		}

	})

}

func scrapShopItems(c *colly.Collector, shop *models.Shop) *models.Shop {

	c.OnHTML(`div[data-appears-component-name="shop_home_listing_grid"]`, func(e *colly.HTMLElement) {

		newItem := models.Item{}
		newItemsSlice := []models.Item{}

		MenuIndex := 0
		CurrentQueueURL := "https://" + e.Request.URL.Host + e.Request.URL.RequestURI()
		Section_ID := GetSectionID(CurrentQueueURL)

		for index, menu := range shop.ShopMenu.Menu {
			if Section_ID == menu.SectionID {
				MenuIndex = index
			}
		}

		e.ForEach("div.js-merch-stash-check-listing", func(i int, h *colly.HTMLElement) {

			ListingID := h.Attr("data-listing-id")
			ListingIDToUint64, err := strconv.ParseUint(ListingID, 10, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}
			newItem.ListingID = uint(ListingIDToUint64)

			ListingIdCount[newItem.ListingID]++

			newItem.DataShopID = h.Attr("data-shop-id")
			newItem.MenuItemID = shop.ShopMenu.Menu[MenuIndex].ID

			divID := "h3#listing-title-" + ListingID
			newItem.Name = h.ChildText(divID)

			OriginalPrice := h.ChildText("span.currency-value")
			OriginalPrice = strings.Replace(OriginalPrice, ",", "", -1)
			SalesPrice := "-1"
			h.ForEachWithBreak("p.search-collage-promotion-price", func(i int, g *colly.HTMLElement) bool {
				SalesPrice = h.DOM.Find("span.currency-value").Eq(0).Text()
				SalesPrice = strings.Replace(SalesPrice, ",", "", -1)

				OriginalPrice = g.ChildText("span.currency-value")
				OriginalPrice = strings.Replace(OriginalPrice, ",", "", -1)

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

			newItemsSlice = append(newItemsSlice, newItem)

		})
		shop.ShopMenu.Menu[MenuIndex].Items = append(shop.ShopMenu.Menu[MenuIndex].Items, newItemsSlice...)
	})

	return shop
}
func GetSectionID(link string) (SectionID string) {
	linkSplit := strings.Split(link, "?")
	if len(linkSplit) > 1 {
		linkSplit = strings.Split(linkSplit[1], "&")
		for _, param := range linkSplit {
			if strings.Contains(param, "section_id") {
				SectionID = strings.Split(param, "=")[1]
				return SectionID
			}
		}
	}
	return ""
}

func CheckCategoryName(Category string) bool {
	MenuCategoryNames := []string{"Out Of Production", "UnCategorized", "On sale"}
	for _, MenuCategoryName := range MenuCategoryNames {
		if Category == MenuCategoryName {
			return true
		}
	}
	return false

}

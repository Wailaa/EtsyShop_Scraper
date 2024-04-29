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

func (sc *Scraper) ScrapAllMenuItems(shop *models.Shop) *models.Shop {

	HasSalesCategory := false
	AllItemCategoryIndex := 0

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

	HandleUnCategorized(shop, HasSalesCategory, AllItemCategoryIndex)

	SectionIdPages = make(map[string]struct{})
	ListingIdCount = make(map[uint]int)
	return shop
}

func scrapNextItemPage(c *colly.Collector, q *queue.Queue) {

	c.OnHTML(`div[data-item-pagination]`, func(h *colly.HTMLElement) {
		CurrentQueueURL := h.Request.URL.Scheme + "://" + h.Request.URL.Host + h.Request.URL.RequestURI()
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

		CurrentQueueURL := e.Request.URL.Scheme + "://" + e.Request.URL.Host + e.Request.URL.RequestURI()
		Section_ID := GetSectionID(CurrentQueueURL)
		MenuIndex := GetMenuIndex(shop, Section_ID)

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

			OriginalPrice, SalesPrice := ExtractPrices(h)

			newItem.OriginalPrice = OriginalPrice

			newItem.SalePrice = SalesPrice

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

func HandleUnCategorized(shop *models.Shop, HasSalesCategory bool, AllItemCategoryIndex int) *models.Shop {
	UnCategorizedItems := []models.Item{}

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

	return shop
}

func GetMenuIndex(shop *models.Shop, SectionID string) int {
	MenuIndex := 0
	for index, menu := range shop.ShopMenu.Menu {
		if SectionID == menu.SectionID {
			MenuIndex = index
			break
		}
	}
	return MenuIndex
}

func StringToFloat(price string) (float64, error) {
	result, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return float64(0), err
	}
	return result, nil
}

func ReplaceSign(sentence, oldSign, newSign string) string {
	result := strings.Replace(sentence, oldSign, newSign, -1)
	return result
}

func ExtractPrices(h *colly.HTMLElement) (float64, float64) {
	OriginalPrice := h.ChildText("span.currency-value")
	OriginalPrice = ReplaceSign(OriginalPrice, ",", "")
	SalesPrice := "-1"
	h.ForEachWithBreak("p.search-collage-promotion-price", func(i int, g *colly.HTMLElement) bool {
		SalesPrice = h.DOM.Find("span.currency-value").Eq(0).Text()
		SalesPrice = ReplaceSign(SalesPrice, ",", "")

		OriginalPrice = g.ChildText("span.currency-value")
		OriginalPrice = ReplaceSign(OriginalPrice, ",", "")

		return false
	})

	SalesPriceToFloat, err := StringToFloat(SalesPrice)
	if err != nil {
		log.Println(err.Error())
		return float64(0), float64(0)
	}

	OriginalPricetoFloat, err := StringToFloat(OriginalPrice)
	if err != nil {
		log.Println(err.Error())
		return float64(0), float64(0)
	}

	return OriginalPricetoFloat, SalesPriceToFloat
}

func HandleItem(h *colly.HTMLElement, MenuID uint) models.Item {
	newItem := models.Item{}
	ListingID := h.Attr("data-listing-id")
	ListingIDToUint64, err := strconv.ParseUint(ListingID, 10, 64)
	if err != nil {
		log.Println(err.Error())
	}

	newItem.ListingID = uint(ListingIDToUint64)

	ListingIdCount[newItem.ListingID]++

	newItem.DataShopID = h.Attr("data-shop-id")
	newItem.MenuItemID = MenuID

	divID := "h3#listing-title-" + ListingID
	newItem.Name = h.ChildText(divID)

	OriginalPrice, SalesPrice := ExtractPrices(h)

	newItem.OriginalPrice = OriginalPrice

	newItem.SalePrice = SalesPrice

	newItem.CurrencySymbol = h.DOM.Find("span.currency-symbol").Eq(0).Text()

	getDiscoutPrice := h.DOM.Find("p.search-collage-promotion-price").Find("span").Last().Text()
	getDiscoutPrice = strings.TrimSpace(getDiscoutPrice)
	newItem.DiscoutPercent = getDiscoutPrice

	newItem.ItemLink = h.ChildAttr("a.listing-link", "href")
	newItem.Available = true

	return newItem
}

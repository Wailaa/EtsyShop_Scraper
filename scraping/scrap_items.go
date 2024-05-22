package scrap

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"

	"EtsyScraper/collector"
	"EtsyScraper/models"
	"EtsyScraper/utils"
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

		MaxSeconds := 89
		utils.SetSleep(MaxSeconds)

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

		AddToQueue(SectionID, pagesCount, link, q)

	})

}

func scrapShopItems(c *colly.Collector, shop *models.Shop) *models.Shop {

	c.OnHTML(`div[data-appears-component-name="shop_home_listing_grid"]`, func(e *colly.HTMLElement) {

		newItemsSlice := []models.Item{}

		CurrentQueueURL := e.Request.URL.Scheme + "://" + e.Request.URL.Host + e.Request.URL.RequestURI()
		SectionID := GetSectionID(CurrentQueueURL)
		MenuIndex := GetMenuIndex(shop, SectionID)

		e.ForEach("div.js-merch-stash-check-listing", func(i int, h *colly.HTMLElement) {

			newItem := HandleItem(h, shop.ShopMenu.Menu[MenuIndex].ID)
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

	if ShouldProcessItems(shop, HasSalesCategory) {
		UnCategorizedItems := FilterUncategorizedItems(shop, AllItemCategoryIndex, ListingIdCount)

		if len(UnCategorizedItems) > 0 {
			CreateUncategorizedMenu(shop, AllItemCategoryIndex, UnCategorizedItems)
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

func ExtractPrices(h *colly.HTMLElement) (float64, float64) {
	OriginalPrice := h.ChildText("span.currency-value")
	OriginalPrice = utils.ReplaceSign(OriginalPrice, ",", "")
	SalesPrice := "-1"
	h.ForEachWithBreak("p.search-collage-promotion-price", func(i int, g *colly.HTMLElement) bool {
		SalesPrice = h.DOM.Find("span.currency-value").Eq(0).Text()
		SalesPrice = utils.ReplaceSign(SalesPrice, ",", "")

		OriginalPrice = g.ChildText("span.currency-value")
		OriginalPrice = utils.ReplaceSign(OriginalPrice, ",", "")

		return false
	})

	SalesPriceToFloat, err := utils.StringToFloat(SalesPrice)
	if err != nil {
		utils.HandleError(nil, err.Error())
		return float64(0), float64(0)
	}

	OriginalPricetoFloat, err := utils.StringToFloat(OriginalPrice)
	if err != nil {
		utils.HandleError(nil, err.Error())
		return float64(0), float64(0)
	}

	return OriginalPricetoFloat, SalesPriceToFloat
}

func HandleItem(h *colly.HTMLElement, MenuID uint) models.Item {
	newItem := models.Item{}
	ListingID := h.Attr("data-listing-id")
	ListingIDToUint64, err := utils.StringToUint(ListingID)
	if err != nil {
		utils.HandleError(nil, err.Error())
	}

	newItem.ListingID = ListingIDToUint64

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

func AddToQueue(SectionID string, pagesCount int, link string, q *queue.Queue) {
	if _, ok := SectionIdPages[SectionID]; !ok {
		for i := 2; i <= pagesCount; i++ {
			SectionIdPages[SectionID] = struct{}{}

			QueueURL := fmt.Sprint(link, "?ref=items-pagination&page=", i, "&section_id=", SectionID, "&sort_order=price_desc")

			q.AddURL(QueueURL)

		}
	}
}

func ShouldProcessItems(shop *models.Shop, hasSalesCategory bool) bool {
	return (len(shop.ShopMenu.Menu) > 1 && !hasSalesCategory) || (len(shop.ShopMenu.Menu) > 2 && hasSalesCategory)
}

func FilterUncategorizedItems(shop *models.Shop, allItemCategoryIndex int, listingIdCount map[uint]int) []models.Item {
	uncategorizedItems := []models.Item{}

	itemsOfCategoryAll := shop.ShopMenu.Menu[allItemCategoryIndex].Items
	for listingID, amount := range listingIdCount {
		if amount == 1 {
			for _, item := range itemsOfCategoryAll {
				if item.ListingID == listingID {
					uncategorizedItems = append(uncategorizedItems, item)
				}
			}
		}
	}

	return uncategorizedItems
}

func CreateUncategorizedMenu(shop *models.Shop, AllItemCategoryIndex int, UnCategorizedItems []models.Item) *models.Shop {
	Menu := shop.ShopMenu.Menu[AllItemCategoryIndex]
	UnCategorizedMenu := models.MenuItem{
		ShopMenuID: Menu.ShopMenuID,
		Category:   "UnCategorized",
		SectionID:  shop.ShopMenu.Menu[AllItemCategoryIndex].SectionID,
		Link:       shop.ShopMenu.Menu[AllItemCategoryIndex].Link,
		Amount:     len(UnCategorizedItems),
		Items:      UnCategorizedItems,
	}

	shop.ShopMenu.Menu = append(shop.ShopMenu.Menu, UnCategorizedMenu)
	return shop
}

func ResetMenuAllToZeroItems(shop *models.Shop, AllItemCategoryIndex int) *models.Shop {
	shop.ShopMenu.Menu[AllItemCategoryIndex].Items = []models.Item{}
	return shop
}

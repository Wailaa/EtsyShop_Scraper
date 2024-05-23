package scrap

import (
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"

	"EtsyScraper/collector"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
)

var Config = initializer.LoadProjConfig(".")
var Shoplink = Config.ScrapShopURL

var MissingInfo string = "INFORMATION_NOT_AVAILABLE"

func (sc *Scraper) ScrapShop(shopName string) (*models.Shop, error) {

	NewShop := &models.Shop{}

	NewShopCollector := collector.NewCollyCollector().C

	NewShopCollector.AllowURLRevisit = true

	NewShopCollector.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 404 {
			r.Request.Abort()
			log.Println("shop was not found. error 404 was returned")
		} else {
			failedURL := "https://" + r.Request.URL.Host + r.Request.URL.RequestURI()

			MaxSeconds := 89
			utils.SetSleep(MaxSeconds)

			NewShopCollector.Visit(failedURL)
		}
	})

	if err := scrapShopDetails(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	if err := scrapShopvacation(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}
	if err := scrapShopTotalSales(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	if err := scrapShopMenu(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	if err := scrapShopAdmirers(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	if err := scrapShopReviews(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	if err := scrapShopLastUpdate(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	if err := scrapShopJoinedSince(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}
	if err := scrapShopMembers(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}
	if err := scrapShopSocialMediaAcc(NewShopCollector, NewShop); err != nil {
		return nil, utils.HandleError(err)
	}

	NewShopCollector.Visit(Shoplink + shopName)
	NewShopCollector.Wait()

	return NewShop, nil
}

func scrapShopDetails(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML("div.shop-home-header-info", func(e *colly.HTMLElement) {

		shop.Name = e.ChildText("div.shop-name-and-title-container h1")
		shop.Description = e.ChildText("div.shop-name-and-title-container h2")
		if shop.Description == "" {
			shop.Description = MissingInfo
		}

		shop.Location = e.ChildText("span.shop-location")
		if shop.Location == "" {
			shop.Location = MissingInfo
		}

	})
	return nil
}

func scrapShopvacation(c *colly.Collector, shop *models.Shop) error {
	shop.OnVacation = false
	c.OnHTML(`div[data-region="vacation-notification-bar"]`, func(e *colly.HTMLElement) {

		shop.OnVacation = true

	})
	return nil
}

func scrapShopTotalSales(c *colly.Collector, shop *models.Shop) error {
	IsElementFound := false
	c.OnHTML(`div[data-appears-component-name="shop_home_listings_section"]`, func(e *colly.HTMLElement) {
		IsElementFound = true

		TotalSales := e.ChildText("div.wt-mt-lg-5 div:first-child")
		TotalSales = strings.Split(TotalSales, " ")[0]
		TotalSales = utils.ReplaceSign(TotalSales, ",", "")
		TotalSalesToInt, _ := strconv.Atoi(TotalSales)

		shop.TotalSales = TotalSalesToInt

		Href := e.ChildAttr("div.wt-mt-lg-5 a", "href")
		if utils.StringContains(Href, "sold") {
			shop.HasSoldHistory = true
		}
	})
	if !IsElementFound {
		shop.TotalSales = 0
	}
	return nil
}

func scrapShopMenu(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML(`div[data-appears-component-name="shop_home_listings_section"]`, func(e *colly.HTMLElement) {
		Menu := []models.MenuItem{}
		e.ForEach("li[data-wt-tab]", func(i int, h *colly.HTMLElement) {

			var key, value string

			attValue := h.ChildText("span[data-shop-pretranslations-translation]")

			if attValue != "" {
				key = h.ChildText("span[data-shop-pretranslations-translation]")
				value = h.ChildText("span.wt-mr-md-2")
			} else {
				key = h.ChildText("span:first-child")
				value = h.ChildText("span:last-child")
			}

			valueToInt, _ := strconv.Atoi(value)

			dataSectionId := h.Attr("data-section-id")
			dataSectionIdlink := Shoplink + shop.Name + "?&section_id=" + dataSectionId

			if i == 0 {
				shop.ShopMenu.TotalItemsAmount = valueToInt
			}
			Menu = append(Menu, models.MenuItem{
				Category:  key,
				Link:      dataSectionIdlink,
				Amount:    valueToInt,
				SectionID: dataSectionId,
			})

			shop.ShopMenu.Menu = Menu
		})
	})
	return nil
}

func scrapShopAdmirers(c *colly.Collector, shop *models.Shop) error {
	IsElementFound := false
	c.OnHTML("div.wt-mt-lg-5", func(e *colly.HTMLElement) {

		IsElementFound = true
		Admirers := e.ChildText("div:nth-child(2)")
		Admirers = strings.Split(Admirers, " ")[0]
		AdmirersToInt, _ := strconv.Atoi(Admirers)

		shop.Admirers = AdmirersToInt

	})
	if !IsElementFound {
		shop.Admirers = 0
	}
	return nil
}

func scrapShopReviews(c *colly.Collector, shop *models.Shop) error {

	c.OnHTML("div.reviews-total", func(e *colly.HTMLElement) {

		ratings := e.ChildAttr("input", "value")
		ratingsToFloat, _ := utils.StringToFloat(ratings)

		totalReviews := e.ChildText("div:last-child")
		totalReviews = totalReviews[1 : len(totalReviews)-1]

		totalReviewsToInt, _ := strconv.Atoi(totalReviews)

		shop.Reviews = models.Reviews{
			ReviewsCount: totalReviewsToInt,
			ShopRating:   ratingsToFloat,
		}
	})

	c.OnHTML(`div[data-appears-component-name="keyword_filters_reviews_page"]`, func(e *colly.HTMLElement) {

		ShopReviewTopic := []models.ReviewsTopic{}

		e.ForEach("button", func(i int, h *colly.HTMLElement) {
			keys := h.Attr("data-keyword-filter")
			value := h.ChildText("span")

			valueToInt, _ := strconv.Atoi(value)

			ShopReviewTopic = append(ShopReviewTopic, models.ReviewsTopic{
				Keyword:      keys,
				KeywordCount: valueToInt,
			})

		})
		shop.Reviews.ReviewsTopic = ShopReviewTopic
	})

	return nil
}

func scrapShopLastUpdate(c *colly.Collector, shop *models.Shop) error {
	IsElementFound := false

	c.OnHTML("span[data-more-last-updated]", func(e *colly.HTMLElement) {
		IsElementFound = true

		shop.LastUpdateTime = e.Text
	})

	if !IsElementFound {
		shop.LastUpdateTime = MissingInfo
	}

	return nil
}

func scrapShopJoinedSince(c *colly.Collector, shop *models.Shop) error {
	IsElementFound := false
	c.OnHTML("#about .shop-home-wider-sections", func(e *colly.HTMLElement) {
		IsElementFound = true
		shop.JoinedSince = e.DOM.Find("span").Eq(1).Text()

	})
	if !IsElementFound {
		shop.JoinedSince = MissingInfo
	}
	return nil
}

func scrapShopMembers(c *colly.Collector, shop *models.Shop) error {

	c.OnHTML("div#shop-members", func(e *colly.HTMLElement) {
		Members := []models.ShopMember{}
		e.ForEach(`li[data-region="shop-member"]`, func(i int, h *colly.HTMLElement) {

			name := h.ChildText(`h6[data-region="member-name"]`)
			role := h.ChildText(`p[data-region="member-role"]`)

			Members = append(Members, models.ShopMember{Name: name, Role: role})
		})
		shop.Member = Members
	})

	return nil
}

func scrapShopSocialMediaAcc(c *colly.Collector, shop *models.Shop) error {

	c.OnHTML("#about div.wt-mb-xs-6", func(e *colly.HTMLElement) {
		links := e.ChildAttrs("a", "href")
		for _, link := range links {
			shop.SocialMediaLinks = append(shop.SocialMediaLinks, models.SocialMediaLinks{Link: link})
		}
	})

	return nil
}

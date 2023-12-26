package scrap

import (
	"EtsyScraper/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

var link string

func ScrapShop(shopName string) (*models.Shop, error) {

	NewShop := &models.Shop{}
	link = "https://www.etsy.com/de-en/shop/"

	c := colly.NewCollector()

	if err := scrapShopDetails(c, NewShop); err != nil {
		return nil, err
	}

	if err := scrapShopTotalSales(c, NewShop); err != nil {
		return nil, err
	}

	if err := scrapShopMenu(c, NewShop); err != nil {
		return nil, err
	}

	if err := scrapShopAdmirers(c, NewShop); err != nil {
		return nil, err
	}

	if err := scrapShopReviews(c, NewShop); err != nil {
		return nil, err
	}

	if err := scrapShopLastUpdate(c, NewShop); err != nil {
		return nil, err
	}

	if err := scrapShopJoinedSince(c, NewShop); err != nil {
		return nil, err
	}
	if err := scrapShopMembers(c, NewShop); err != nil {
		return nil, err
	}
	if err := scrapShopSocialMediaAcc(c, NewShop); err != nil {
		return nil, err
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Got this error:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	c.Visit(link + shopName)
	c.Wait()

	return NewShop, nil
}

func scrapShopDetails(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML("div.shop-home-header-info", func(e *colly.HTMLElement) {

		shop.Name = e.ChildText("div.shop-name-and-title-container h1")
		shop.Description = e.ChildText("div.shop-name-and-title-container p")
		shop.Location = e.ChildText("span.shop-location")

	})
	return nil
}
func scrapShopTotalSales(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML(`div[data-appears-component-name="shop_home_about_section"]`, func(e *colly.HTMLElement) {

		TotalSales := e.ChildText("div.wt-mr-xs-6 span")
		TotalSales = strings.Replace(TotalSales, ",", "", -1)
		TotalSalesToInt, _ := strconv.Atoi(TotalSales)

		shop.TotalSales = TotalSalesToInt
	})
	return nil
}

func scrapShopMenu(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML(`div[data-appears-component-name="shop_home_listings_section"]`, func(e *colly.HTMLElement) {
		Menu := []models.MenuItem{}
		e.ForEach("li[data-wt-tab]", func(i int, h *colly.HTMLElement) {

			key := h.ChildText("span:nth-child(1)")
			value := h.ChildText("span:nth-child(2)")
			valueToInt, _ := strconv.Atoi(value)
			dataSectionId := h.Attr("data-section-id")
			dataSectionId_link := link + shop.Name + "?ref=shop_sugg_market&section_id=" + dataSectionId

			if i == 0 {
				shop.ShopMenu.TotalItemsAmmount = valueToInt
			} else {
				Menu = append(Menu, models.MenuItem{
					Category:  key,
					Link:      dataSectionId_link,
					Amount:    valueToInt,
					SectionID: dataSectionId,
				})
			}
			shop.ShopMenu.Menu = Menu
		})
	})
	return nil
}

func scrapShopAdmirers(c *colly.Collector, shop *models.Shop) error {

	c.OnHTML("div.wt-mt-lg-5", func(e *colly.HTMLElement) {

		Admirers := e.ChildText("div:nth-child(2)")
		Admirers = strings.Split(Admirers, " ")[0]
		AdmirersToInt, _ := strconv.Atoi(Admirers)

		shop.Admirers = AdmirersToInt

	})
	return nil
}

func scrapShopReviews(c *colly.Collector, shop *models.Shop) error {

	c.OnHTML("div.reviews-total", func(e *colly.HTMLElement) {

		ratings := e.ChildAttr("input", "value")
		ratingsToFloat, _ := strconv.ParseFloat(ratings, 64)

		totalReviews := e.ChildText("div:last-child")
		totalReviews = totalReviews[1 : len(totalReviews)-1]
		totalReviewsToInt, _ := strconv.Atoi(totalReviews)

		shop.Reviews = models.Reviews{
			ReviewsCount: totalReviewsToInt,
			ShopRating:   ratingsToFloat,
		}
	})

	c.OnHTML(`div[data-appears-component-name="keyword_filters_reviews_page"]`, func(e *colly.HTMLElement) {
		shop.Reviews.ReviewsTopic = make(map[string]int)
		e.ForEach("button", func(i int, h *colly.HTMLElement) {
			keys := h.Attr("data-keyword-filter")
			value := h.ChildText("span")

			valueToInt, _ := strconv.Atoi(value)

			shop.Reviews.ReviewsTopic[keys] = valueToInt

		})
	})
	return nil
}

func scrapShopLastUpdate(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML("span[data-more-last-updated]", func(e *colly.HTMLElement) {
		shop.LastUpdateTime = e.Text
	})
	return nil
}

func scrapShopJoinedSince(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML("#about .shop-home-wider-sections", func(e *colly.HTMLElement) {
		shop.JoinedSince = e.DOM.Find("span").Eq(1).Text()

	})
	return nil
}

func scrapShopMembers(c *colly.Collector, shop *models.Shop) error {

	c.OnHTML("div#shop-members", func(e *colly.HTMLElement) {
		shop.Member.Members = make(map[int]*models.Member)
		e.ForEach(`li[data-region="shop-member"]`, func(i int, h *colly.HTMLElement) {

			name := h.ChildText(`h6[data-region="member-name"]`)
			role := h.ChildText(`p[data-region="member-role"]`)

			shop.Member.Members[i+1] = &models.Member{Name: name, Role: role}
			shop.Member.Amount = len(shop.Member.Members)
		})
	})
	return nil
}

func scrapShopSocialMediaAcc(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML("#about div.wt-mb-xs-6", func(e *colly.HTMLElement) {
		shop.SocialMediaLinks = e.ChildAttrs("a", "href")
	})
	return nil
}

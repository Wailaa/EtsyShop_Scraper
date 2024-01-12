package scrap

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/imroc/req/v3"
)

var config = initializer.LoadProjConfig(".")
var Shoplink = config.ScrapShopURL

var MissingInfo string = "INFORMATION_NOT_AVAILABLE"

func ScrapShop(shopName string) (*models.Shop, error) {

	NewShop := &models.Shop{}

	Chrome := req.DefaultClient().ImpersonateChrome()

	c := colly.NewCollector(colly.AllowURLRevisit())

	c.UserAgent = utils.GetRandomUserAgent()

	extensions.Referer(c)

	c.Limit(&colly.LimitRule{
		Delay:       5 * time.Second,
		RandomDelay: 5 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {

		c.SetProxy(config.ProxyHostURL)
		c.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})

		c.UserAgent = utils.GetRandomUserAgent()

		c.SetClient(&http.Client{
			Transport: Chrome.Transport,
		})
		fmt.Println("-----------------------------")
		fmt.Println("Visiting", r.URL)
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept", "test/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		for key, value := range *r.Headers {
			fmt.Printf("%s: %s\n", key, value)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("-----------------------------")
		fmt.Println("Responce on Scraping a Shop")
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

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

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

	c.Visit(Shoplink + shopName)
	c.Wait()

	return NewShop, nil
}

func scrapShopDetails(c *colly.Collector, shop *models.Shop) error {
	c.OnHTML("div.shop-home-header-info", func(e *colly.HTMLElement) {

		shop.Name = e.ChildText("div.shop-name-and-title-container h1")
		shop.Description = e.ChildText("div.shop-name-and-title-container p")
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
func scrapShopTotalSales(c *colly.Collector, shop *models.Shop) error {
	IsElementFound := false
	c.OnHTML(`div[data-appears-component-name="shop_home_listings_section"]`, func(e *colly.HTMLElement) {
		IsElementFound = true

		TotalSales := e.ChildText("div.wt-mt-lg-5 div:first-child")
		TotalSales = strings.Split(TotalSales, " ")[0]
		TotalSales = strings.Replace(TotalSales, ",", "", -1)
		TotalSalesToInt, _ := strconv.Atoi(TotalSales)

		shop.TotalSales = TotalSalesToInt

		Href := e.ChildAttr("div.wt-mt-lg-5 a", "href")
		if strings.Contains(Href, "sold") {
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
			dataSectionId_link := Shoplink + shop.Name + "?&section_id=" + dataSectionId

			if i == 0 {
				shop.ShopMenu.TotalItemsAmmount = valueToInt
			}
			Menu = append(Menu, models.MenuItem{
				Category:  key,
				Link:      dataSectionId_link,
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
	IsElementFound := false
	c.OnHTML("#about div.wt-mb-xs-6", func(e *colly.HTMLElement) {
		IsElementFound = true
		shop.SocialMediaLinks = e.ChildAttrs("a", "href")
	})
	if !IsElementFound {
		shop.SocialMediaLinks = []string{MissingInfo}
	}
	return nil
}

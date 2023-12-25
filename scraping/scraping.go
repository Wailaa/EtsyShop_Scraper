package main

import (
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

var link string

type Shop struct {
	Name             string   `json:"shop_name"`
	Description      string   `json:"shop_description"`
	Location         string   `json:"location"`
	TotalSales       int      `json:"shop_total_sales"`
	JoinedSince      string   `json:"joined_since"`
	ShopMenu         ShopMenu `json:"shop_menu"`
	LastUpdateTime   string   `json:"last_update_time"`
	Admirers         int      `json:"admirers"`
	Reviews          Reviews  `json:"shop_reviews"`
	SocialMediaLinks []string `json:"social_media_links"`
	Member           Members  `json:"shop_member"`
}

type MenuItem struct {
	Category  string `json:"category_name"`
	SectionID string `json:"selection_id"`
	Link      string `json:"link"`
	Amount    int    `json:"item_amount"`
}

type ShopMenu struct {
	Menu map[int]MenuItem `json:"shop_item_id"`
}
type Reviews struct {
	ShopRating   float64        `json:"shop_rate"`
	ReviewsCount int            `json:"reviews_count"`
	ReviewsTopic map[string]int `json:"reviews_mentions"`
}

type Members struct {
	Amount  int             `json:"amount"`
	Members map[int]*Member `json:"members"`
}
type Member struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func ScrapShop(shopName string) (*Shop, error) {

	NewShop := &Shop{}
	link = "https://www.etsy.com/de-en/shop/"
	c := colly.NewCollector()

	c.Visit(link + shopName)
	c.Wait()

	if err := scrapShopDetails(c, NewShop); err != nil {
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

	return NewShop, nil
}

func scrapShopDetails(c *colly.Collector, shop *Shop) error {
	c.OnHTML("div.shop-home-header-info", func(e *colly.HTMLElement) {

		shop.Name = e.ChildText("div.shop-name-and-title-container h1")
		shop.Description = e.ChildText("div.shop-name-and-title-container p")
		shop.Location = e.ChildText("span.shop-location")
		TotalSales := e.ChildText("span.wt-text-caption a")

		TotalSales = strings.Split(TotalSales, " ")[0]
		TotalSales = strings.Replace(TotalSales, ",", "", -1)
		TotalSalesToInt, _ := strconv.Atoi(TotalSales)

		shop.TotalSales = TotalSalesToInt
	})
	return nil
}

func scrapShopMenu(c *colly.Collector, shop *Shop) error {
	c.OnHTML(`div[data-appears-component-name="shop_home_listings_section"]`, func(e *colly.HTMLElement) {
		shop.ShopMenu = ShopMenu{make(map[int]MenuItem)}
		e.ForEach("li[data-wt-tab]", func(i int, h *colly.HTMLElement) {

			key := h.ChildText("span:nth-child(1)")
			value := h.ChildText("span:nth-child(2)")
			valueToInt, _ := strconv.Atoi(value)
			dataSectionId := h.Attr("data-section-id")
			dataSectionId_link := link + shop.Name + "?ref=shop_sugg_market&section_id=" + dataSectionId

			shop.ShopMenu.Menu[i+1] = MenuItem{
				Category:  key,
				Link:      dataSectionId_link,
				Amount:    valueToInt,
				SectionID: dataSectionId,
			}

		})
	})
	return nil
}

func scrapShopAdmirers(c *colly.Collector, shop *Shop) error {

	c.OnHTML("div.wt-mt-lg-5", func(e *colly.HTMLElement) {

		Admirers := e.ChildText("div:nth-child(2)")
		Admirers = strings.Split(Admirers, " ")[0]
		AdmirersToInt, _ := strconv.Atoi(Admirers)

		shop.Admirers = AdmirersToInt

	})
	return nil
}

func scrapShopReviews(c *colly.Collector, shop *Shop) error {
	shop.Reviews.ReviewsTopic = make(map[string]int)
	c.OnHTML("div.reviews-total", func(e *colly.HTMLElement) {

		ratings := e.ChildAttr("input", "value")
		ratingsToFloat, _ := strconv.ParseFloat(ratings, 64)

		totalReviews := e.ChildText("div:last-child")
		totalReviews = totalReviews[1 : len(totalReviews)-1]
		totalReviewsToInt, _ := strconv.Atoi(totalReviews)

		shop.Reviews.ReviewsCount = totalReviewsToInt
		shop.Reviews.ShopRating = ratingsToFloat
	})

	c.OnHTML(`div[data-appears-component-name="keyword_filters_reviews_page"] button`, func(e *colly.HTMLElement) {

		keys := e.Attr("data-keyword-filter")
		value := e.ChildText("span")

		valueToInt, _ := strconv.Atoi(value)

		shop.Reviews.ReviewsTopic[keys] = valueToInt
	})

	return nil
}

func scrapShopLastUpdate(c *colly.Collector, shop *Shop) error {
	c.OnHTML("span[data-more-last-updated]", func(e *colly.HTMLElement) {
		shop.LastUpdateTime = e.Text
	})
	return nil
}

func scrapShopJoinedSince(c *colly.Collector, shop *Shop) error {
	c.OnHTML("#about .shop-home-wider-sections", func(e *colly.HTMLElement) {
		shop.JoinedSince = e.DOM.Find("span").Eq(1).Text()

	})
	return nil
}

func scrapShopMembers(c *colly.Collector, shop *Shop) error {
	shop.Member.Members = make(map[int]*Member)

	c.OnHTML("div#shop-members", func(e *colly.HTMLElement) {
		e.ForEach(`li[data-region="shop-member"]`, func(i int, h *colly.HTMLElement) {
			name := h.ChildText(`h6[data-region="member-name"]`)
			role := h.ChildText(`p[data-region="member-role"]`)
			newMember := &Member{Name: name, Role: role}
			shop.Member.Members[i+1] = newMember
			shop.Member.Amount = len(shop.Member.Members)
		})
	})
	return nil
}

func scrapShopSocialMediaAcc(c *colly.Collector, shop *Shop) error {
	c.OnHTML("#about div.wt-mb-xs-6", func(e *colly.HTMLElement) {
		shop.SocialMediaLinks = e.ChildAttrs("a", "href")
	})
	return nil
}

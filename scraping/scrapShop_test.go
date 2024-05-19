package scrap

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"EtsyScraper/collector"
	"EtsyScraper/models"
	"EtsyScraper/utils"
)

func TestScrapShop_AllCallBacks(t *testing.T) {
	shop := &models.Shop{}

	ShopMenuItems := `[{"category_name":"All","link":"MissArtisanShop?\u0026section_id=0","item_amount":15}]`
	ShopReviews := `{"shop_rate":4.9286,"reviews_count":217,"reviews_mentions":[{"keyword":"quality","keyword_count":47},{"keyword":"shipping","keyword_count":29},{"keyword":"customer_service","keyword_count":42}]}`
	ShopMembers := `[{"name":"example","role":"Shopkeeper, Владелец"},{"name":"example","role":"Maker, Shipper"}]`
	scrapShopSocialLink := `[{"link":"https://www.facebook.com/MissArtisan/"},{"link":"https://www.miss-artisan.com"}]`

	collector.RateLimiting = 0 * time.Second
	c := collector.NewCollyCollector().C

	tr := &http.Transport{}
	tr.RegisterProtocol("file", http.NewFileTransport(http.Dir("../.")))
	c.WithTransport(tr)

	scrapShopDetails(c, shop)
	scrapShopvacation(c, shop)
	scrapShopTotalSales(c, shop)
	scrapShopMenu(c, shop)
	scrapShopAdmirers(c, shop)
	scrapShopReviews(c, shop)
	scrapShopLastUpdate(c, shop)
	scrapShopJoinedSince(c, shop)
	scrapShopMembers(c, shop)
	scrapShopSocialMediaAcc(c, shop)

	u, err := url.Parse("file://./setupTests/testing.html")
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}

	err = c.Request("GET", u.String(), bytes.NewReader(nil), nil, nil)
	if err != nil {
		log.Fatalf("Failed to request local file: %v", err)
	}

	c.Wait()

	ShopMenuJSON, err := utils.MarshalJSONData(shop.ShopMenu.Menu)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	ShopReviewJSON, err := utils.MarshalJSONData(shop.Reviews)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	ShopMembersJson, err := utils.MarshalJSONData(shop.Member)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	ShopSocialMediaLinkJson, err := utils.MarshalJSONData(shop.SocialMediaLinks)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	ShopSocialMediaLinkAsString := string(ShopSocialMediaLinkJson)

	ShopMenuAsString := string(ShopMenuJSON)

	ShopMembersString := string(ShopMembersJson)

	ShopReviewAsString := string(ShopReviewJSON)

	assert.Equal(t, "MissArtisanShop", shop.Name)
	assert.Equal(t, "Own Something Beautiful Made With Love", shop.Description)
	assert.Equal(t, "London, United Kingdom", shop.Location)
	assert.Equal(t, false, shop.OnVacation)
	assert.Equal(t, 694, shop.Admirers)
	assert.Equal(t, 2072, shop.TotalSales)
	assert.Equal(t, ShopMenuItems, ShopMenuAsString)
	assert.Equal(t, ShopReviews, ShopReviewAsString)
	assert.Equal(t, "Nov 7, 2023", shop.LastUpdateTime)
	assert.Equal(t, "2018", shop.JoinedSince)
	assert.Equal(t, ShopMembers, ShopMembersString)
	assert.Equal(t, scrapShopSocialLink, ShopSocialMediaLinkAsString)

}

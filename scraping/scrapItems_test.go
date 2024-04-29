package scrap

import (
	"EtsyScraper/collector"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"log"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/stretchr/testify/assert"
)

func TestScrapALLMenuItems_Success(t *testing.T) {
	UpdateScraper := &Scraper{}
	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	collector.RateLimiting = 0 * time.Second

	isUnCategorized := false
	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	Shop := models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Link:      mockURL + "/ExampleShop?&section_id=0",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Link:      mockURL + "/ExampleShop?&section_id=1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Link:      mockURL + "/ExampleShop?&section_id=46696458",
					Amount:    45,
					Items:     []models.Item{{}, {}, {}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Link:      mockURL + "/ExampleShop?&section_id=46704593",
					Amount:    44,
					Items:     []models.Item{{}, {}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Link:      mockURL + "/ExampleShop?&section_id=46704591",
					Amount:    46,
					Items:     []models.Item{{}, {}, {}, {}},
				},
			},
		},
	}

	UpdateScraper.ScrapAllMenuItems(&Shop)

	for _, menu := range Shop.ShopMenu.Menu {
		if menu.Category == "UnCategorized" {
			isUnCategorized = true
		}
		if menu.Category != "On sale" && menu.Category != "All" {
			assert.Equal(t, menu.Amount, len(menu.Items))
		} else {
			assert.Equal(t, 0, len(menu.Items))
		}
	}

	assert.False(t, isUnCategorized)

}

func TestScrapALLMenuItems_UnCategorized(t *testing.T) {
	collector.RateLimiting = 0 * time.Second
	UpdateScraper := &Scraper{}
	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	ListingIdCount = map[uint]int{
		1: 1,
		2: 1,
		3: 1,
	}

	isUnCategorized := false
	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	Shop := models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Link:      mockURL + "/ExampleShop?&section_id=0",
					Amount:    0,
					Items:     []models.Item{{ListingID: 1}, {ListingID: 2}, {ListingID: 3}},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Link:      mockURL + "/ExampleShop?&section_id=1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Link:      mockURL + "/ExampleShop?&section_id=46696458",
					Amount:    24,
					Items:     []models.Item{{}, {}, {}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Link:      mockURL + "/ExampleShop?&section_id=46704593",
					Amount:    23,
					Items:     []models.Item{{}, {}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Link:      mockURL + "/ExampleShop?&section_id=46704591",
					Amount:    25,
					Items:     []models.Item{{}, {}, {}, {}},
				},
			},
		},
	}

	UpdateScraper.ScrapAllMenuItems(&Shop)

	UnCategorizedIndex := 0
	for index, menu := range Shop.ShopMenu.Menu {
		if menu.Category == "UnCategorized" {
			isUnCategorized = true
			UnCategorizedIndex = index
			break
		}

	}

	assert.True(t, isUnCategorized)
	assert.Equal(t, 3, len(Shop.ShopMenu.Menu[UnCategorizedIndex].Items))

}

func TestScrapNextItemPage_Success(t *testing.T) {
	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	SectionIdPages = map[string]struct{}{}

	collector.RateLimiting = 0 * time.Second
	c := collector.NewCollyCollector().C

	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingItems.html")
	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	scrapNextItemPage(c, OriginalQueue)

	c.Visit(mockURL)
	c.Wait()

	QueueSize, err := OriginalQueue.Size()
	if err != nil {
		t.Fatalf("error getting the size of the queue , the error : %v", err)
	}

	assert.Equal(t, 1, QueueSize)
	assert.Equal(t, 1, len(SectionIdPages))

}

func TestScrapShopItems(t *testing.T) {

	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	collector.RateLimiting = 0 * time.Second

	c := collector.NewCollyCollector().C

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	Shop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Link:      mockURL + "/ExampleShop?&section_id=0",
					Amount:    0,
					Items:     []models.Item{{ListingID: 1}, {ListingID: 2}, {ListingID: 3}},
				},
			},
		},
	}

	scrapShopItems(c, Shop)

	c.Visit(mockURL)
	c.Wait()

	itemCount := len(Shop.ShopMenu.Menu[0].Items)

	assert.Equal(t, 24, itemCount)

}

func TestGetSectionID_Success(t *testing.T) {

	link := "http://example.com/ExampleShop?section_id=46704591"
	SectionID := GetSectionID(link)
	assert.Equal(t, "46704591", SectionID)

}

func TestGetSectionID_MultipleParams(t *testing.T) {

	link := "http://example.com/ExampleShop?ref=items-pagination&page=2&section_id=46704591&sort_order=price_desc"
	SectionID := GetSectionID(link)
	assert.Equal(t, "46704591", SectionID)

}

func TestGetSectionID_EmptyString(t *testing.T) {

	link := "http://example.com/ExampleShop?ref=items-pagination&page=2"
	SectionID := GetSectionID(link)
	assert.Equal(t, "", SectionID)

}

func TestCheckCategoryName(t *testing.T) {
	result := CheckCategoryName("Out Of Production")
	assert.True(t, result)

	result = CheckCategoryName("UnCategorized")
	assert.True(t, result)

	result = CheckCategoryName("On sale")
	assert.True(t, result)

	result = CheckCategoryName("Chairs")
	assert.False(t, result)

}

func TestHandleUnCategorized_CreateUnCategorized(t *testing.T) {
	AllItemCategoryIndex := 0
	UpdatedShop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}, {ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}, {ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}, {ListingID: 10, DataShopID: "101", OriginalPrice: 10}, {ListingID: 11, DataShopID: "101", OriginalPrice: 10}, {ListingID: 12, DataShopID: "101", OriginalPrice: 10}, {ListingID: 13, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Amount:    45,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{{ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{{ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}, {ListingID: 10, DataShopID: "101", OriginalPrice: 10}},
				},
			},
		},
	}
	for i := uint(1); i <= 13; i++ {
		ListingIdCount[i]++
		if i < 11 {
			ListingIdCount[i]++
		}

	}
	log.Println(ListingIdCount)
	for index, menu := range UpdatedShop.ShopMenu.Menu {
		ID := uint(index + 1)
		UpdatedShop.ShopMenu.Menu[index].ID = ID

		if menu.Category == "All" {
			AllItemCategoryIndex = index
		}

	}
	HasSalesCategory := false
	IsUnCategorized := false
	UpdatedShop = HandleUnCategorized(UpdatedShop, HasSalesCategory, AllItemCategoryIndex)

	for _, Menu := range UpdatedShop.ShopMenu.Menu {
		if Menu.Category == "UnCategorized" {
			IsUnCategorized = true
			break
		}
	}

	assert.True(t, IsUnCategorized, "UnCategorized category is created")
}

func TestGetMenuIndex(t *testing.T) {
	UpdatedShop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}, {ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}, {ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}, {ListingID: 10, DataShopID: "101", OriginalPrice: 10}, {ListingID: 11, DataShopID: "101", OriginalPrice: 10}, {ListingID: 12, DataShopID: "101", OriginalPrice: 10}, {ListingID: 13, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Amount:    45,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{{ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{{ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}, {ListingID: 10, DataShopID: "101", OriginalPrice: 10}},
				},
			},
		},
	}

	tests := []struct {
		SectioID  string
		MenuIndex int
	}{
		{
			SectioID:  "0",
			MenuIndex: 0,
		},
		{
			SectioID:  "1",
			MenuIndex: 1,
		},
		{
			SectioID:  "46696458",
			MenuIndex: 2,
		},
		{
			SectioID:  "46704593",
			MenuIndex: 3,
		},
		{
			SectioID:  "46704591",
			MenuIndex: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.SectioID, func(t *testing.T) {
			actual := GetMenuIndex(UpdatedShop, tc.SectioID)
			if actual != tc.MenuIndex {
				t.Errorf("Expected GetMenuIndex to be %v, but got %v", tc.MenuIndex, actual)
			}
		})
	}

}

func TestStringToFloat(t *testing.T) {
	tests := []struct {
		Price  string
		result float64
		err    error
	}{
		{
			Price:  "19.7",
			result: 19.7,
		},
		{
			Price:  "1.8",
			result: 1.8,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Price, func(t *testing.T) {
			actual, _ := StringToFloat(tc.Price)
			if actual != tc.result {
				t.Errorf("Expected StringToFloat to be %v, but got %v", tc.result, actual)
			}
		})
	}
}

func TestReplaceSign(t *testing.T) {
	tests := []struct {
		Price    string
		oldSign  string
		newSign  string
		expected string
	}{
		{
			Price:    "1,232$",
			oldSign:  ",",
			newSign:  "",
			expected: "1232$",
		},
		{
			Price:    "1,232$",
			oldSign:  ",",
			newSign:  "",
			expected: "1232$",
		},
		{
			Price:    "1232$",
			oldSign:  ".",
			newSign:  "",
			expected: "1232$",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Price, func(t *testing.T) {
			actual := ReplaceSign(tc.Price, tc.oldSign, tc.newSign)
			if actual != tc.expected {
				t.Errorf("Expected StringToFloat to be %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestExtractPrices(t *testing.T) {

	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	collector.RateLimiting = 0 * time.Second

	c := collector.NewCollyCollector().C

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	tests := []struct {
		Expected []float64
	}{

		{

			Expected: []float64{21.77, 19.59},
		},
		{

			Expected: []float64{251.28, 238.72},
		},
		{

			Expected: []float64{426.93, 405.59},
		},
		{

			Expected: []float64{880.7, 836.67},
		},
		{

			Expected: []float64{308.61, 293.18},
		},
		{

			Expected: []float64{307.39, 292.02},
		},
	}

	c.OnHTML(`div[data-appears-component-name="shop_home_listing_grid"]`, func(e *colly.HTMLElement) {
		e.ForEachWithBreak("div.js-merch-stash-check-listing", func(i int, h *colly.HTMLElement) bool {

			OriginalPrice, SalesPrice := ExtractPrices(h)
			t.Run("", func(t *testing.T) {
				if OriginalPrice != tests[i].Expected[0] || SalesPrice != tests[i].Expected[1] {
					t.Errorf("Expected OriginalPrice to be %v, but got %v", tests[i].Expected[0], OriginalPrice)
				}
			})

			return i != 5
		})

	})

	c.Visit(mockURL)
	c.Wait()

}

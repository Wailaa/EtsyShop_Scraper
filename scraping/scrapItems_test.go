package scrap

import (
	"EtsyScraper/collector"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"log"
	"reflect"
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
func TestHandleItem(t *testing.T) {

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
		Expected models.Item
	}{
		{
			Expected: models.Item{Name: "INDUSTRIAL COAT HOOK- steampunk wall art", OriginalPrice: 21.77, CurrencySymbol: "€", SalePrice: 19.59, DiscoutPercent: "(10% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1616116159/industrial-coat-hook-steampunk-wall-art?click_key=0db4c2898f0e84d918cac1c3b0e13c3a09cfe4d4%3A1616116159\u0026click_sum=53d95add\u0026ref=shop_home_active_2\u0026pro=1", ListingID: 1616116159, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "Steampunk shelving unit - Retro Industrial wall art", OriginalPrice: 251.28, CurrencySymbol: "€", SalePrice: 238.72, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1573116439/steampunk-shelving-unit-retro-industrial?click_key=8fae5cf715d357c866d8a2805451331280299b16%3A1573116439\u0026click_sum=295cd595\u0026ref=shop_home_active_3\u0026pro=1", ListingID: 1573116439, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "The Manic Steampunk Side Table", OriginalPrice: 426.93, CurrencySymbol: "€", SalePrice: 405.59, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1572514925/the-manic-steampunk-side-table?click_key=de8520515537e83cb4803fbe876e59f3b20ea7ea%3A1572514925\u0026click_sum=f4f366f8\u0026ref=shop_home_active_4\u0026pro=1", ListingID: 1572514925, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "The industrial Shelving Unit - Steampunk wall art", OriginalPrice: 880.7, CurrencySymbol: "€", SalePrice: 836.67, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1539348644/the-industrial-shelving-unit-steampunk?click_key=efb54d919d2ca63a829189e57565939743a22ff4%3A1539348644\u0026click_sum=4b2c33b8\u0026ref=shop_home_active_5\u0026pro=1", ListingID: 1539348644, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "Brunel Steampunk shelving unit - Retro Industrial wall art", OriginalPrice: 308.61, CurrencySymbol: "€", SalePrice: 293.18, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1463373323/brunel-steampunk-shelving-unit-retro?click_key=5c76bd7420a06ad07de08392226c0b9c343f4b61%3A1463373323\u0026click_sum=91132e67\u0026ref=shop_home_active_6\u0026pro=1", ListingID: 1463373323, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "The Stephenson Steampunk Shelving Unit- - Retro Industrial wall art", OriginalPrice: 307.39, CurrencySymbol: "€", SalePrice: 292.02, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1468110403/the-stephenson-steampunk-shelving-unit?click_key=57bd4a72bfc90861d969a5697562fbf7e4a5f84f%3A1468110403\u0026click_sum=62e359b4\u0026ref=shop_home_active_7\u0026pro=1", ListingID: 1468110403, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "locke Steampunk shelving unit - Retro Industrial wall art Plant stand", OriginalPrice: 185.41, CurrencySymbol: "€", SalePrice: 176.14, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1527884665/locke-steampunk-shelving-unit-retro?click_key=4a26efa99764e4ed46c55ba9505421d8295c8387%3A1527884665\u0026click_sum=1b18c02f\u0026ref=shop_home_active_8\u0026pro=1", ListingID: 1527884665, PriceHistory: nil},
		},
		{
			Expected: models.Item{Name: "The Reynolds industrial Shelving Unit - Steampunk wall art", OriginalPrice: 602.59, CurrencySymbol: "€", SalePrice: 572.46, DiscoutPercent: "(5% off)", Available: true, ItemLink: "https://www.etsy.com/de-en/listing/1479436896/the-reynolds-industrial-shelving-unit?click_key=4fad9bdab5eb88c0a9f6fb4a95ea21fbc6a540f9%3A1479436896\u0026click_sum=b1da47c5\u0026ref=shop_home_active_9\u0026pro=1", ListingID: 1479436896, PriceHistory: nil},
		},
	}

	MenuID := uint(1)
	c.OnHTML(`div[data-appears-component-name="shop_home_listing_grid"]`, func(e *colly.HTMLElement) {
		e.ForEachWithBreak("div.js-merch-stash-check-listing", func(i int, h *colly.HTMLElement) bool {

			Items := HandleItem(h, MenuID)

			t.Run(tests[i].Expected.Name, func(t *testing.T) {
				if reflect.DeepEqual(Items, tests[i].Expected) {
					t.Errorf("Expected OriginalPrice to be %v, but got %v", tests[i].Expected, Items)
				}
			})
			return i != 7

		})

	})

	c.Visit(mockURL)
	c.Wait()

}
func TestAddToQueue(t *testing.T) {

	tests := []struct {
		name        string
		Section_ID  string
		pageCount   int
		link        string
		queueLength int
	}{
		{
			name:        "menu has 5 pages",
			Section_ID:  "0",
			pageCount:   5,
			link:        "www.JustEcample.com",
			queueLength: 4,
		},
		{
			name:        "has no pages",
			Section_ID:  "1",
			pageCount:   0,
			link:        "www.JustEcample.com",
			queueLength: 0,
		},
		{
			name:        "SectionID already consumed",
			Section_ID:  "10",
			pageCount:   5,
			link:        "www.JustEcample.com",
			queueLength: 0,
		},
	}

	SectionIdPages["10"] = struct{}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q, _ := queue.New(
				1,
				&queue.InMemoryQueueStorage{MaxSize: 10000},
			)

			AddToQueue(tc.Section_ID, tc.pageCount, tc.link, q)
			ActualQueueSize, _ := q.Size()
			if ActualQueueSize != tc.queueLength {
				t.Errorf("Expected StringToFloat to be %v, but got %v", tc.queueLength, ActualQueueSize)
			}
		})
	}

}

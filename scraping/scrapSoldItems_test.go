package scrap

import (
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/stretchr/testify/assert"

	"EtsyScraper/collector"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
)

func TestScrapesSoldItems_Success(t *testing.T) {
	scraper := &Scraper{}
	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	task := &models.TaskSchedule{
		CurrentPage:          0,
		LastPage:             0,
		IsPaginationScrapped: false,
		IsScrapeFinished:     false,
		UpdateSoldItems:      0,
	}

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingSoldItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	items, task := scraper.ScrapSalesHistory(mockURL, task)

	assert.Equal(t, 24, len(items))
	assert.Equal(t, 2, task.CurrentPage)
	assert.Equal(t, 87, task.LastPage)
	assert.True(t, task.IsPaginationScrapped)
	assert.False(t, task.IsScrapeFinished)
	assert.True(t, len(items) > 0)

}

func TestScrapSoldItems_Success(t *testing.T) {
	c := collector.NewCollyCollector().C
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       0 * time.Second,
		RandomDelay: 0 * time.Second,
	})

	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingSoldItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	items := scrapSoldItems(c)

	c.Visit(mockURL)
	c.Wait()

	assert.Equal(t, 24, len(*items))
}

func TestScrapSoldItemPages_Success(t *testing.T) {
	ShopName := "Example"

	c := collector.NewCollyCollector().C
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       0 * time.Second,
		RandomDelay: 0 * time.Second,
	})
	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	Task := &models.TaskSchedule{}

	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingSoldItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	Task = scrapSoldItemPages(c, ShopName, Task, OriginalQueue)

	c.Visit(mockURL)
	c.Wait()

	assert.Equal(t, 2, Task.CurrentPage)
	assert.Equal(t, 87, Task.LastPage)
	assert.True(t, Task.IsPaginationScrapped)
	assert.False(t, Task.IsScrapeFinished)

}

func TestScrapSoldItemPages_IsPaginationScrapped(t *testing.T) {
	ShopName := "Example"

	c := collector.NewCollyCollector().C
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       0 * time.Second,
		RandomDelay: 0 * time.Second,
	})
	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	Task := &models.TaskSchedule{
		CurrentPage:          0,
		LastPage:             0,
		IsPaginationScrapped: true,
		IsScrapeFinished:     false,
		UpdateSoldItems:      0,
	}

	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
	}
	Config = mockConfig

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingSoldItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	Task = scrapSoldItemPages(c, ShopName, Task, OriginalQueue)

	c.Visit(mockURL)
	c.Wait()

	assert.Equal(t, 0, Task.CurrentPage)
	assert.Equal(t, 0, Task.LastPage)
	assert.True(t, Task.IsPaginationScrapped)
	assert.False(t, Task.IsScrapeFinished)
}

func TestAddURLtoQueue_CurrentPage(t *testing.T) {
	ShopName := "Example"
	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
		MaxPageLimit:  9,
	}
	Config = mockConfig
	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	Task := &models.TaskSchedule{
		CurrentPage:          7,
		LastPage:             87,
		IsPaginationScrapped: true,
		IsScrapeFinished:     false,
		UpdateSoldItems:      0,
	}
	Task = AddURLtoQueue(ShopName, Task, OriginalQueue)

	assert.Equal(t, 16, Task.CurrentPage)

}

func TestAddURLtoQueue_IsScrapeFinished(t *testing.T) {
	ShopName := "Example"
	mockConfig := initializer.Config{
		ProxyHostURL1: "",
		ProxyHostURL2: "",
		ProxyHostURL3: "",
		MaxPageLimit:  9,
	}
	Config = mockConfig
	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	Task := &models.TaskSchedule{
		CurrentPage:          80,
		LastPage:             87,
		IsPaginationScrapped: true,
		IsScrapeFinished:     false,
		UpdateSoldItems:      0,
	}
	Task = AddURLtoQueue(ShopName, Task, OriginalQueue)

	assert.True(t, Task.IsScrapeFinished)

}

func TestExtractPageNumber(t *testing.T) {
	URL := "http://example.com?ref=pagination&page=10"

	Task := &models.TaskSchedule{
		CurrentPage:          4,
		LastPage:             87,
		IsPaginationScrapped: true,
		IsScrapeFinished:     true,
		UpdateSoldItems:      0,
	}
	Task = ExtractPageNumber(URL, Task)

	assert.Equal(t, 10, Task.CurrentPage)
	assert.False(t, Task.IsScrapeFinished)

}

func TestExtractPageNumber_NoPage(t *testing.T) {
	URL := "http://example.com"

	Task := &models.TaskSchedule{
		CurrentPage:          4,
		LastPage:             87,
		IsPaginationScrapped: true,
		IsScrapeFinished:     true,
		UpdateSoldItems:      0,
	}
	Task = ExtractPageNumber(URL, Task)

	assert.Equal(t, 0, Task.CurrentPage)
	assert.Equal(t, 0, Task.LastPage)
	assert.False(t, Task.IsPaginationScrapped)

}

package scrap

import (
	"EtsyScraper/collector"
	"EtsyScraper/models"
	"strings"
	"sync"

	"fmt"
	"log"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

func ScrapSalesHistory(ShopName string, Task *models.TaskSchedule) ([]models.SoldItems, *models.TaskSchedule) {
	TerminateCollector := false
	var TerminateCollectorMutex sync.Mutex

	c := collector.NewCollyCollector().C

	c.AllowURLRevisit = true

	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	// backUpQueue, _ := queue.New(
	// 	1,
	// 	&queue.InMemoryQueueStorage{MaxSize: 10000},
	// )

	c.OnRequest(func(r *colly.Request) {
		if TerminateCollector {
			r.Abort()
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		failedURL := r.Request.URL.String()
		log.Println("failed url is :", failedURL)

		// randTimeSet := time.Duration(rand.Intn(89-10) + 10)
		// time.Sleep(randTimeSet * time.Second)

		Task = ExtractFailedPage(failedURL, Task)
		TerminateCollector = true
		// backUpQueue.AddURL(failedURL)

	})

	Items := scrapSoldItems(c)

	Task = scrapSoldItemPages(c, ShopName, Task, OriginalQueue)

	if Task.FirstPage > 2 {
		pageString := fmt.Sprint(Task.FirstPage)
		OriginalQueue.AddURL(Shoplink + ShopName + "/sold?ref=pagination&page=" + pageString)
	} else {
		OriginalQueue.AddURL(Shoplink + ShopName + "/sold")
	}

	for !TerminateCollector && !OriginalQueue.IsEmpty() {
		TerminateCollectorMutex.Lock()
		defer TerminateCollectorMutex.Unlock()
		OriginalQueue.Run(c)
		c.Wait()
	}
	// backUpQueue.Run(c)
	// c.Wait()

	return *Items, Task
}

func scrapSoldItems(c *colly.Collector) *[]models.SoldItems {
	TotalItemSold := &[]models.SoldItems{}

	c.OnHTML("div#content", func(e *colly.HTMLElement) {
		itemsSold := models.SoldItems{}

		e.ForEach("div[data-shop-id]", func(i int, h *colly.HTMLElement) {

			ListingID := h.Attr("data-listing-id")
			ListingIDToUint64, err := strconv.ParseUint(ListingID, 10, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}
			ListingIDToUint := uint(ListingIDToUint64)
			itemsSold.ListingID = ListingIDToUint

			itemsSold.DataShopID = h.Attr("data-shop-id")

			divID := "h3#listing-title-" + ListingID
			itemsSold.Name = h.ChildText(divID)

			itemsSold.ItemLink = h.ChildAttr("a.listing-link", "href")

			*TotalItemSold = append(*TotalItemSold, itemsSold)

		})

	})

	return TotalItemSold
}

func scrapSoldItemPages(c *colly.Collector, ShopName string, Task *models.TaskSchedule, OriginalQueue *queue.Queue) *models.TaskSchedule {

	c.OnHTML("div#content", func(h *colly.HTMLElement) {

		SoldPages := []string{}
		if !Task.IsScrapped && Task.FirstPage != Task.LastPage {
			if Task.LastPage == 0 {
				h.ForEach("li", func(i int, k *colly.HTMLElement) {
					page := k.ChildAttr("a", "data-page")
					SoldPages = append(SoldPages, page)
				})

				lastpage := SoldPages[len(SoldPages)-2]
				Task.LastPage, _ = strconv.Atoi(lastpage)
			}
			Task.IsScrapped = true

			loopStart := Task.FirstPage
			if Task.FirstPage != 2 {
				loopStart = Task.FirstPage + 1
			}

			loopEnds := 0
			if Task.LastPage > Task.FirstPage+config.MaxPageLimit {
				loopEnds = Task.FirstPage + config.MaxPageLimit
				Task.FirstPage = loopEnds + 1
			} else {
				loopEnds = Task.LastPage
				Task.FirstPage = 0
				Task.LastPage = 0
			}

			for pageNum := loopStart; pageNum <= loopEnds; pageNum++ {
				link := fmt.Sprintf("%s%s/sold?ref=pagination&page=%d", Shoplink, ShopName, pageNum)
				OriginalQueue.AddURL(link)

			}

		}

	})

	return Task
}

func ExtractFailedPage(url string, Task *models.TaskSchedule) *models.TaskSchedule {
	splitURL := strings.Split(url, "ref=pagination&page=")
	if len(splitURL) == 1 {
		return Task
	}
	page, err := strconv.Atoi(splitURL[1])
	if err != nil {
		log.Println("error while extracting page number", err)
		return nil
	}
	Task.FirstPage = page

	return Task
}

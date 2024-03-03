package scrap

import (
	"EtsyScraper/collector"
	"EtsyScraper/models"
	"strings"

	"fmt"
	"log"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

func ScrapSalesHistory(ShopName string, Task *models.TaskSchedule) ([]models.SoldItems, *models.TaskSchedule) {
	TerminateCollector := false
	Items := &[]models.SoldItems{}
	c := collector.NewCollyCollector().C

	c.AllowURLRevisit = true

	OriginalQueue, _ := queue.New(
		1,
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	c.OnRequest(func(r *colly.Request) {
		if len(*Items) >= Task.UpdateSoldItems && Task.UpdateSoldItems != 0 {
			log.Println("Task.NewSoldItems :", Task.UpdateSoldItems)
			*Items = (*Items)[:Task.UpdateSoldItems]
			TerminateCollector = true
			Task.IsScrapeFinished = true
		}
		if TerminateCollector {
			r.Abort()
			log.Println("Request is aborted")
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		failedURL := r.Request.URL.String()
		log.Println("failed url is :", failedURL)

		Task = ExtractPageNumber(failedURL, Task)
		TerminateCollector = true

	})

	Items = scrapSoldItems(c)

	Task = scrapSoldItemPages(c, ShopName, Task, OriginalQueue)

	if Task.CurrentPage != 0 {
		AddURLtoQueue(ShopName, Task, OriginalQueue)

	} else {
		OriginalQueue.AddURL(Shoplink + ShopName + "/sold")

	}

	OriginalQueue.Run(c)

	c.Wait()

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
		IsPagination := false
		if !Task.IsPaginationScrapped {

			h.ForEach("li", func(i int, k *colly.HTMLElement) {

				page := k.ChildAttr("a", "data-page")
				SoldPages = append(SoldPages, page)
				IsPagination = true
			})

			if IsPagination {

				lastpage := SoldPages[len(SoldPages)-2]
				Task.CurrentPage = 2
				Task.LastPage, _ = strconv.Atoi(lastpage)
				Task.IsPaginationScrapped = true
				AddURLtoQueue(ShopName, Task, OriginalQueue)
			} else {
				Task.IsScrapeFinished = true
			}

		}

	})

	return Task
}

func AddURLtoQueue(ShopName string, Task *models.TaskSchedule, OriginalQueue *queue.Queue) *models.TaskSchedule {

	loopStart := Task.CurrentPage

	loopEnds := Task.CurrentPage + config.MaxPageLimit

	for pageNum := loopStart; pageNum < loopEnds; pageNum++ {
		link := fmt.Sprintf("%s%s/sold?ref=pagination&page=%d", Shoplink, ShopName, pageNum)
		OriginalQueue.AddURL(link)
		Task.CurrentPage = pageNum + 1
		if pageNum == Task.LastPage {
			Task.IsScrapeFinished = true
			break
		}
	}

	return Task
}

func ExtractPageNumber(url string, Task *models.TaskSchedule) *models.TaskSchedule {
	splitURL := strings.Split(url, "ref=pagination&page=")
	if len(splitURL) == 1 {
		return Task
	}
	page, err := strconv.Atoi(splitURL[1])
	if err != nil {
		log.Println("error while extracting page number", err)
		return nil
	}
	Task.CurrentPage = page
	Task.IsScrapeFinished = false

	return Task
}

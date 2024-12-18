package collector

import (
	"log"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"

	"EtsyScraper/utils"
)

type DefaultCollector struct {
	C *colly.Collector
}

var RateLimiting = 5 * time.Second

func NewCollyCollector() *DefaultCollector {
	utils := &utils.Utils{}
	Chrome := req.DefaultClient().ImpersonateChrome()
	getProxy := utils.PickProxyProvider()

	c := colly.NewCollector()

	c.SetClient(&http.Client{
		Transport: Chrome.Transport,
	})

	if getProxy.Url != "" {
		c.WithTransport(&http.Transport{
			DisableKeepAlives: true,
		})

		c.SetProxy(getProxy.Url)
	}

	c.UserAgent = utils.GetRandomUserAgent()

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       RateLimiting,
		RandomDelay: RateLimiting,
	})

	c.OnRequest(func(r *colly.Request) {

		log.Println("-----------------------------")
		log.Println("Visiting", r.URL)
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept", "test/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		for key, value := range *r.Headers {
			log.Printf("%s: %s\n", key, value)
		}
	})

	c.OnResponse(func(r *colly.Response) {

		log.Println("-----------------------------")
		log.Println(r.StatusCode)

		log.Println("ProxyProvider :", getProxy.Provider)

		if r.StatusCode != 200 {
			for key, value := range *r.Headers {
				log.Printf("%s: %s\n", key, value)
			}
		}

	})

	c.OnError(func(r *colly.Response, err error) {

		log.Println("ProxyProvider :", getProxy.Provider)

		getProxy = utils.PickProxyProvider()

		if r.StatusCode == 404 {
			r.Request.Abort()
			log.Println("shop was not found. error 404 was returned")
		} else {
			log.Println("Request URL: ", r.Request.URL, " failed with response: ", r, "\nError: ", err)

			if getProxy.Url != "" {
				c.SetProxy(getProxy.Url)
			}

			c.UserAgent = utils.GetRandomUserAgent()

			c.SetCookies(r.Request.URL.String(), nil)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		log.Println("-----------------------------")
		log.Println("Done scraping")
	})

	return &DefaultCollector{
		C: c,
	}
}

package collector

import (
	initializer "EtsyScraper/init"
	"EtsyScraper/utils"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/imroc/req/v3"
)

var config = initializer.LoadProjConfig(".")

type DefaultCollector struct {
	C *colly.Collector
}

func NewCollyCollector() *DefaultCollector {
	Chrome := req.DefaultClient().ImpersonateChrome()

	c := colly.NewCollector()

	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	})

	c.SetProxy(config.ProxyHostURL)

	c.UserAgent = utils.GetRandomUserAgent()

	c.SetClient(&http.Client{
		Transport: Chrome.Transport,
	})

	extensions.Referer(c)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       5 * time.Second,
		RandomDelay: 5 * time.Second,
	})
	c.OnRequest(func(r *colly.Request) {

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
		fmt.Println(r.StatusCode)
		if r.StatusCode != 200 {
			for key, value := range *r.Headers {
				fmt.Printf("%s: %s\n", key, value)
			}
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL: ", r.Request.URL, " failed with response: ", r, "\nError: ", err)
		for key, value := range *r.Headers {
			fmt.Printf("%s: %s\n", key, value)
		}

		r.Headers.Del("Cookie")
		c.WithTransport(&http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		})

		c.SetProxy(config.ProxyHostURL)

		c.UserAgent = utils.GetRandomUserAgent()

		c.SetClient(&http.Client{
			Transport: Chrome.Transport,
		})

	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("-----------------------------")
		fmt.Println("Done scraping")
	})

	return &DefaultCollector{
		C: c,
	}
}

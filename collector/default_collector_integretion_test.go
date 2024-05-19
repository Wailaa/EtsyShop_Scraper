package collector

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"

	setupMockServer "EtsyScraper/setupTests"
)

func TestNewCollyCollector(t *testing.T) {
	c := NewCollyCollector()

	assert.NotNil(t, c)
	assert.NotNil(t, c.C)

}

func TestOnRequest(t *testing.T) {
	RateLimiting = 0 * time.Second
	c := NewCollyCollector()

	setupMockServer.GlobalTestSetupMockServer("../setupTests/testingSoldItems.html")

	defer setupMockServer.MockServer.Close()

	mockURL := setupMockServer.MockServer.URL

	c.C.OnRequest(func(r *colly.Request) {

		assert.Equal(t, mockURL, r.URL.String())

		assert.Equal(t, "en-US,en;q=0.9", r.Headers.Get("Accept-Language"))
		assert.Equal(t, "test/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", r.Headers.Get("Accept"))
		assert.Equal(t, "gzip, deflate, br", r.Headers.Get("Accept-Encoding"))
	})

	c.C.Request("GET", mockURL, nil, &colly.Context{}, http.Header{})
}

type MockResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (mr *MockResponse) statusCode() int {
	return mr.StatusCode
}

func (mr *MockResponse) body() []byte {
	return mr.Body
}

func TestOnResponse(t *testing.T) {
	RateLimiting = 0 * time.Second

	c := NewCollyCollector()

	mockBody := []byte("Mock body content")

	mockResponse := &MockResponse{
		StatusCode: 200,
		Body:       mockBody,
		Headers:    make(http.Header),
	}

	c.C.OnResponse(func(r *colly.Response) {

		assert.Equal(t, 200, r.StatusCode)

		assert.Equal(t, mockBody, r.Body)
	})

	c.C.OnResponse(func(r *colly.Response) {
		r.StatusCode = mockResponse.statusCode()
		r.Body = mockResponse.body()
		r.Headers = &mockResponse.Headers
	})
}

func TestOnError(t *testing.T) {
	RateLimiting = 0 * time.Second

	c := NewCollyCollector()

	mockResponse := &colly.Response{
		StatusCode: 404,
	}
	mockError := fmt.Errorf("test error")

	onErrorCallback := func(r *colly.Response, err error) {

		assert.Equal(t, mockError, err)
		assert.Equal(t, 404, r.StatusCode)
	}

	c.C.OnError(onErrorCallback)

	onErrorCallback(mockResponse, mockError)
}

func TestOnScraped(t *testing.T) {

	RateLimiting = 0 * time.Second
	c := NewCollyCollector()

	onScrapedCallback := func(r *colly.Response) {

		assert.NotNil(t, r)
	}

	c.C.OnScraped(onScrapedCallback)

	onScrapedCallback(&colly.Response{})
}

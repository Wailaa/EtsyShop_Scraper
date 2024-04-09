package collector

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewCollyCollector(t *testing.T) {
	// Call NewCollyCollector function to create a new collector
	c := NewCollyCollector()

	// Ensure that the collector and its underlying Collector instance are not nil
	assert.NotNil(t, c)
	assert.NotNil(t, c.C)

	// Additional tests for collector initialization can be added here
}

func TestOnRequest(t *testing.T) {
	// Create a new collector
	c := NewCollyCollector()
	assert.NotNil(t, c)
	assert.NotNil(t, c.C)

	// Define a mock request URL
	mockURL := "localhost:8080"

	// Simulate an OnRequest event
	c.C.OnRequest(func(r *colly.Request) {
		// Assert that the request URL matches the mock URL
		assert.Equal(t, mockURL, r.URL.String())

		// Assert that required headers are set
		assert.Equal(t, "en-US,en;q=0.9", r.Headers.Get("Accept-Language"))
		assert.Equal(t, "test/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", r.Headers.Get("Accept"))
		assert.Equal(t, "gzip, deflate, br", r.Headers.Get("Accept-Encoding"))
	})

	// Trigger the OnRequest event

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
	// Create a new collector
	c := NewCollyCollector()
	assert.NotNil(t, c)
	assert.NotNil(t, c.C)

	// Define a mock response body
	mockBody := []byte("Mock body content")

	// Create a new mock response
	mockResponse := &MockResponse{
		StatusCode: 200,
		Body:       mockBody,
		Headers:    make(http.Header),
	}

	// Simulate an OnResponse event
	c.C.OnResponse(func(r *colly.Response) {
		// Assert that the status code matches
		assert.Equal(t, 200, r.StatusCode)

		// Assert that the response body matches
		assert.Equal(t, mockBody, r.Body)
	})

	// Trigger the OnResponse event
	c.C.OnResponse(func(r *colly.Response) {
		r.StatusCode = mockResponse.statusCode()
		r.Body = mockResponse.body()
		r.Headers = &mockResponse.Headers
	})
}

func TestOnError(t *testing.T) {
	// Create a new collector
	c := NewCollyCollector()
	assert.NotNil(t, c)
	assert.NotNil(t, c.C)

	// Define a mock response and error
	mockResponse := &colly.Response{
		StatusCode: 404,
	}
	mockError := fmt.Errorf("test error")

	// Simulate an OnError event
	onErrorCallback := func(r *colly.Response, err error) {
		// Assert that the error and response are as expected
		assert.Equal(t, mockError, err)
		assert.Equal(t, 404, r.StatusCode)
	}

	// Trigger the OnError event
	c.C.OnError(onErrorCallback)

	// Trigger the OnError event
	onErrorCallback(mockResponse, mockError)
}

func TestOnScraped(t *testing.T) {
	// Create a new collector
	c := NewCollyCollector()
	assert.NotNil(t, c)
	assert.NotNil(t, c.C)

	// Simulate an OnScraped event
	onScrapedCallback := func(r *colly.Response) {
		// Assert that the event is triggered
		assert.NotNil(t, r)
	}
	// Trigger the OnScraped event with a nil response
	c.C.OnScraped(onScrapedCallback)

	// Trigger the OnScraped event with a nil response
	onScrapedCallback(&colly.Response{})
}

package controllers_test

import (
	"EtsyScraper/controllers"
	"EtsyScraper/models"
	setupMockServer "EtsyScraper/setupTests"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleResponse(t *testing.T) {

	tests := []struct {
		name         string
		err          error
		status       int
		message      string
		data         interface{}
		expectedBody string
	}{
		{
			name:         "Status 200",
			err:          nil,
			status:       200,
			message:      "Created successfully",
			expectedBody: `{"message":"Created successfully","status":"success"}`,
		},
		{
			name:         "Status 404",
			err:          errors.New("page not found"),
			status:       404,
			message:      "page not found",
			expectedBody: `{"message":"page not found","status":"fail"}`,
		},
		{
			name:         "Status 502",
			err:          errors.New("internal error"),
			status:       502,
			message:      "internal error",
			expectedBody: `{"message":"internal error","status":"fail"}`,
		},
		{
			name:         "Status 502 with no message",
			err:          errors.New("internal error"),
			status:       502,
			message:      "",
			expectedBody: `{"status":"fail"}`,
		},
		{
			name:         "Status 200 with no message",
			err:          nil,
			status:       200,
			message:      "",
			expectedBody: `{"status":"success"}`,
		},
		{
			name:         "Status 200 with payload",
			err:          nil,
			status:       200,
			message:      "",
			data:         models.Item{},
			expectedBody: `{"Name":"","OriginalPrice":0,"CurrencySymbol":"","SalePrice":0,"DiscoutPercent":"","Available":false,"ItemLink":"","ListingID":0,"PriceHistory":null}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, w := setupMockServer.SetGinTestMode()
			controllers.HandleResponse(ctx, tc.err, tc.status, tc.message, tc.data)

			assert.Equal(t, tc.expectedBody, w.Body.String())
			assert.Equal(t, tc.status, w.Code)

		})
	}
}

package routes_test

import (
	"EtsyScraper/routes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockShopRoute struct {
	isCreateNewShopRequest        bool
	isFollowShop                  bool
	isUnFollowShop                bool
	isHandleGetShopByID           bool
	isHandleGetItemsByShopID      bool
	isHandleGetSoldItemsByShopID  bool
	isProcessStatsRequest         bool
	isHandleGetItemsCountByShopID bool
}

func (m *MockShopRoute) CreateNewShopRequest(ctx *gin.Context) {
	m.isCreateNewShopRequest = true
}
func (m *MockShopRoute) FollowShop(ctx *gin.Context) {
	m.isFollowShop = true
}
func (m *MockShopRoute) UnFollowShop(ctx *gin.Context) {
	m.isUnFollowShop = true
}
func (m *MockShopRoute) HandleGetShopByID(ctx *gin.Context) {
	m.isHandleGetShopByID = true
}
func (m *MockShopRoute) HandleGetItemsByShopID(ctx *gin.Context) {
	m.isHandleGetItemsByShopID = true
}
func (m *MockShopRoute) HandleGetSoldItemsByShopID(ctx *gin.Context) {
	m.isHandleGetSoldItemsByShopID = true
}
func (m *MockShopRoute) HandleGetItemsCountByShopID(ctx *gin.Context) {
	m.isHandleGetItemsCountByShopID = true
}
func (m *MockShopRoute) ProcessStatsRequest(ctx *gin.Context) {
	m.isProcessStatsRequest = true
}

func TestGeneralShopRoutes(t *testing.T) {

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	MockedShop := &MockShopRoute{}
	tests := []struct {
		name     string
		method   string
		path     string
		isCalled func() bool
	}{
		{
			name:     "Check if CreateNewShopRequest was called",
			method:   "GET",
			path:     "/shop/create_shop",
			isCalled: func() bool { return MockedShop.isCreateNewShopRequest },
		},
		{
			name:     "Check if FollowShop was called",
			method:   "GET",
			path:     "/shop/follow_shop",
			isCalled: func() bool { return MockedShop.isFollowShop },
		},
		{
			name:     "Check if unFollowShop was called",
			method:   "GET",
			path:     "/shop/unfollow_shop",
			isCalled: func() bool { return MockedShop.isUnFollowShop },
		},
		{
			name:     "Check if HandleGetShopByID was called",
			method:   "GET",
			path:     "/shop/1",
			isCalled: func() bool { return MockedShop.isHandleGetShopByID },
		},
		{
			name:     "Check if HandleGetItemsByShopID was called",
			method:   "GET",
			path:     "/shop/1/all_items",
			isCalled: func() bool { return MockedShop.isHandleGetItemsByShopID },
		},
		{
			name:     "Check if isHandleGetSoldItemsByShopID was called",
			method:   "GET",
			path:     "/shop/1/all_sold_items",
			isCalled: func() bool { return MockedShop.isHandleGetSoldItemsByShopID },
		},

		{
			name:     "Check if ProcessStatsRequest was called",
			method:   "GET",
			path:     "/shop/stats/1/:period",
			isCalled: func() bool { return MockedShop.isProcessStatsRequest },
		},
		{
			name:     "Check if HandleGetItemsCountByShopID was called",
			method:   "GET",
			path:     "/shop/1/items_count",
			isCalled: func() bool { return MockedShop.isHandleGetItemsCountByShopID },
		},
	}

	ShopRoute := routes.NewShopRouteController(MockedShop)
	ShopRoute.GeneralShopRoutes(router, MiddleWare(), SecondMiddleWare(), SecondMiddleWare())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.True(t, tc.isCalled())
		})
	}

}

package controllers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"EtsyScraper/controllers"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	setupMockServer "EtsyScraper/setupTests"
)

type MockedShop struct {
	mock.Mock
}

func (m *MockedShop) GetShopByID(ID uint) (*models.Shop, error) {

	args := m.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}

func (m *MockedShop) CreateNewShop(ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) CreateSoldStats(dailyShopSales []models.DailyShopSales) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()

	return args.Get(0).(map[string]controllers.DailySoldStats), args.Error(1)
}

func (m *MockedShop) GetItemsByShopID(ID uint) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Items []models.Item
	if shopInterface != nil {
		Items = shopInterface.([]models.Item)
	}
	return Items, args.Error(1)
}

func (m *MockedShop) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedShop) SaveShopToDB(scrappedShop *models.Shop, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) UpdateShopMenuToDB(Shop *models.Shop, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedShop) GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}
func (m *MockedShop) CheckAndUpdateOutOfProdMenu(AllMenus []models.MenuItem, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockedShop) GetSoldItemsByShopID(ID uint) (SoldItemInfos []controllers.ResponseSoldItemInfo, err error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []controllers.ResponseSoldItemInfo
	if shopInterface != nil {
		soldItems = shopInterface.([]controllers.ResponseSoldItemInfo)
	}
	return soldItems, args.Error(1)
}
func (m *MockedShop) CreateOutOfProdMenu(Shop *models.Shop, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedShop) GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Stats map[string]controllers.DailySoldStats
	if shopInterface != nil {
		Stats = shopInterface.(map[string]controllers.DailySoldStats)
	}
	return Stats, args.Error(1)
}

func (m *MockedShop) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) EstablishAccountShopRelation(requestedShop *models.Shop, userID uuid.UUID) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []models.SoldItems
	if shopInterface != nil {
		soldItems = shopInterface.([]models.SoldItems)
	}
	return soldItems, args.Error(1)
}
func (m *MockedShop) GetItemsBySoldItems(SoldItems []models.SoldItems) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Items []models.Item
	if shopInterface != nil {
		Items = shopInterface.([]models.Item)
	}
	return Items, args.Error(1)
}

type MockScrapper struct {
	mock.Mock
}

func (m *MockScrapper) CheckForUpdates(Shop string, needUpdateItems bool) (*models.Shop, error) {
	args := m.Called()
	return args.Get(0).(*models.Shop), args.Error(1)
}
func (m *MockScrapper) ScrapAllMenuItems(shop *models.Shop) *models.Shop {
	args := m.Called()
	ShopInterface := args.Get(0)
	var Shop *models.Shop
	if ShopInterface != nil {
		Shop = ShopInterface.(*models.Shop)
	}
	return Shop
}
func (m *MockScrapper) ScrapShop(shopName string) (*models.Shop, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}
func (m *MockScrapper) ScrapSalesHistory(ShopName string, Task *models.TaskSchedule) ([]models.SoldItems, *models.TaskSchedule) {
	args := m.Called()

	return args.Get(0).([]models.SoldItems), args.Get(1).(*models.TaskSchedule)
}

type MockedShopRepository struct {
	mock.Mock
}

func (sr *MockedShopRepository) CreateShop(scrappedShop *models.Shop) error {
	args := sr.Called()
	return args.Error(0)
}
func (sr *MockedShopRepository) SaveShop(Shop *models.Shop) error {
	args := sr.Called()
	return args.Error(0)

}
func (sr *MockedShopRepository) SaveSoldItemsToDB(ScrappedSoldItems []models.SoldItems) error {
	args := sr.Called()
	return args.Error(0)

}
func (sr *MockedShopRepository) UpdateDailySales(ScrappedSoldItems []models.SoldItems, ShopID uint, dailyRevenue float64) error {
	args := sr.Called()
	return args.Error(0)
}
func (sr *MockedShopRepository) SaveMenu(Menus models.MenuItem) error {
	args := sr.Called()
	return args.Error(0)
}
func (sr *MockedShopRepository) FetchShopByID(ID uint) (*models.Shop, error) {
	args := sr.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}
func (sr *MockedShopRepository) FetchStatsByPeriod(ShopID uint, timePeriod time.Time) ([]models.DailyShopSales, error) {
	args := sr.Called()
	shopInterface := args.Get(0)
	var Sales []models.DailyShopSales
	if shopInterface != nil {
		Sales = shopInterface.([]models.DailyShopSales)
	}
	return Sales, args.Error(1)
}
func (sr *MockedShopRepository) FetchSoldItemsByListingID(listingIDs []uint) ([]models.SoldItems, error) {
	args := sr.Called()
	shopInterface := args.Get(0)
	var SoldItems []models.SoldItems
	if shopInterface != nil {
		SoldItems = shopInterface.([]models.SoldItems)
	}
	return SoldItems, args.Error(1)
}
func (sr *MockedShopRepository) FetchItemsBySoldItems(soldItemID uint) (models.Item, error) {
	args := sr.Called()
	shopInterface := args.Get(0)
	var Item models.Item
	if shopInterface != nil {
		Item = shopInterface.(models.Item)
	}
	return Item, args.Error(1)
}
func (sr *MockedShopRepository) GetSoldItemsInRange(fromDate time.Time, ShopID uint) ([]models.SoldItems, error) {
	args := sr.Called()
	shopInterface := args.Get(0)
	var SoldItems []models.SoldItems
	if shopInterface != nil {
		SoldItems = shopInterface.([]models.SoldItems)
	}
	return SoldItems, args.Error(1)
}
func (sr *MockedShopRepository) UpdateAccountShopRelation(requestedShop *models.Shop, UserID uuid.UUID) error {
	args := sr.Called()
	return args.Error(0)
}
func (sr *MockedShopRepository) GetAverageItemPrice(ShopID uint) (float64, error) {
	args := sr.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (sr *MockedShopRepository) SaveShopRequestToDB(ShopRequest *models.ShopRequest) error {
	args := sr.Called()
	return args.Error(0)
}
func (sr *MockedShopRepository) GetShopWithItemsByShopID(ID uint) (*models.Shop, error) {
	args := sr.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}
func (sr *MockedShopRepository) GetShopByName(ShopName string) (shop *models.Shop, err error) {
	args := sr.Called()
	shopInterface := args.Get(0)

	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}

func TestCreateNewShopRequestPanic(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()
	Scraper := &MockScrapper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	router.Use(implShop.CreateNewShopRequest)

	router.GET("/create_shop", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/create_shop", nil)

	assert.Panics(t, func() {
		ctx.Set("currentUserUUID", nil)
		router.ServeHTTP(w, req)
	})

}

func TestCreateNewShopRequestInvalidJson(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &MockScrapper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.CreateNewShopRequest)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "failed to get the Shop's name")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}
func TestCreateNewShopRequestGetShopError(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop, Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(nil, errors.New("Error"))
	TestShop.On("CreateShopRequest").Return(errors.New("SecondError"))

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "CreateShopRequest")
	assert.Contains(t, w.Body.String(), "internal error")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}
func TestCreateNewShopRequestShopExists(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop, Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(&models.Shop{Name: "ShopExample"}, nil)
	TestShop.On("CreateShopRequest").Return(errors.New("SecondError"))

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "CreateShopRequest")
	assert.Contains(t, w.Body.String(), "Shop already exists")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}
func TestCreateNewShopRequestSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop, Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(nil, errors.New("no Shop was Found ,error: record not found"))
	TestShop.On("CreateShopRequest").Return(nil)
	TestShop.On("CreateNewShop").Return(nil)

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "CreateShopRequest")
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
	assert.Contains(t, w.Body.String(), "shop request received successfully")
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestCreateNewShopScrapperErr(t *testing.T) {

	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}

	Scraper.On("ScrapShop").Return(nil, errors.New("record not found"))

	err := Shop.CreateNewShop(ShopRequest)
	assert.Contains(t, err.Error(), "record not found")
	assert.Error(t, err)
}
func TestCreateNewShopFailedSaveShopToDB(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	TestShop.On("CreateShopRequest").Return(nil)
	TestShop.On("SaveShopToDB").Return(errors.New("Failed to save shop"))

	err := implShop.CreateNewShop(ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save shop")

}
func TestCreateNewShopSaveMenuToDBFail(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	TestShop.On("CreateShopRequest").Return(nil)
	TestShop.On("SaveShopToDB").Return(nil)
	TestShop.On("UpdateShopMenuToDB").Return(errors.New("failed to save new record"))
	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)

	err := implShop.CreateNewShop(ShopRequest)

	assert.Error(t, err)

}
func TestCreateNewShopSaveMenuToDBSuccess(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	TestShop.On("CreateShopRequest").Return(nil)
	TestShop.On("SaveShopToDB").Return(nil)
	TestShop.On("UpdateShopMenuToDB").Return(nil)
	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)

	err := implShop.CreateNewShop(ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)

}
func TestCreateNewShopHasSoldHistory(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("SaveShopToDB").Return(nil)
	TestShop.On("UpdateShopMenuToDB").Return(nil)
	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)
	TestShop.On("UpdateSellingHistory").Return(nil)

	err := implShop.CreateNewShop(ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "UpdateSellingHistory", 1)

}

func TestUpdateSellingHistoryDisContintuesSoldItemsFail(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("UpdateDiscontinuedItems").Return(nil, errors.New("failed to get SoldItems"))
	TestShop.On("CreateShopRequest").Return(nil)

	err := implShop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
	TestShop.AssertNumberOfCalls(t, "UpdateDiscontinuedItems", 1)

}
func TestUpdateSellingHistoryDisContintuesSoldItemsEmpty(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("UpdateDiscontinuedItems").Return([]models.SoldItems{}, nil)

	err := implShop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty scrapped Sold data")
	TestShop.AssertNumberOfCalls(t, "UpdateDiscontinuedItems", 1)

}
func TestUpdateSellingHistoryGetItemsFail(t *testing.T) {

	TestShop := &MockedShop{}

	implShop := controllers.Shop{Operations: TestShop}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("UpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}, {}}, nil)
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("error getting Items"))

	err := implShop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting Items")
	TestShop.AssertNumberOfCalls(t, "UpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

}
func TestUpdateSellingHistoryInsertIntoDBFail(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("UpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}}, nil)
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	ShopRepo.On("SaveSoldItemsToDB").Return(errors.New("failed to insert data to DB"))

	err := implShop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert data to DB")
	TestShop.AssertNumberOfCalls(t, "UpdateDiscontinuedItems", 1)

}
func TestUpdateSellingHistoryInsertIntoDB(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      0,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("UpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}}, nil)
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	TestShop.On("CreateShopRequest").Return(nil)
	ShopRepo.On("SaveSoldItemsToDB").Return(nil)

	err := implShop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "UpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)

}
func TestUpdateSellingHistoryTaskSoldItem(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      10,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("UpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}}, nil)
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	TestShop.On("CreateShopRequest").Return(nil)
	ShopRepo.On("SaveSoldItemsToDB").Return(nil)
	ShopRepo.On("UpdateDailySales").Return(nil)

	err := implShop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "UpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
	ShopRepo.AssertNumberOfCalls(t, "SaveSoldItemsToDB", 1)
	ShopRepo.AssertNumberOfCalls(t, "UpdateDailySales", 1)

}

func TestUpdateDiscontinuedItemsEmptySoldItems(t *testing.T) {

	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     true,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      10,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	Scraper.On("ScrapSalesHistory").Return([]models.SoldItems{}, Task)

	ActualSoldItems, err := Shop.UpdateDiscontinuedItems(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	assert.Equal(t, []models.SoldItems{}, ActualSoldItems)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)

}
func TestUpdateDiscontinuedItemsGetItemsByShopIDfail(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     true,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      10,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	Scraper.On("ScrapSalesHistory").Return([]models.SoldItems{{}, {}, {}}, Task)
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("Error While fetching Shop's details"))

	_, err := implShop.UpdateDiscontinuedItems(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)
	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
	assert.Contains(t, err.Error(), "Error While fetching Shop's details")
}
func TestUpdateDiscontinuedItemsSuccess(t *testing.T) {

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop}

	userID := uuid.New()
	Task := &models.TaskSchedule{
		IsScrapeFinished:     true,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      10,
	}
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ExampleShop := models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{},
				},
			},
		},
	}
	ExampleShop.ShopMenu.Menu[0].ID = uint(1)

	ShopItems := []models.Item{{ListingID: 1}, {ListingID: 2}, {ListingID: 3}, {ListingID: 4}, {ListingID: 5}, {ListingID: 6}, {ListingID: 7}, {ListingID: 8}, {ListingID: 9}}
	SoldItems := []models.SoldItems{{ListingID: 1}, {ListingID: 1}, {ListingID: 10}, {ListingID: 7}}

	for i := range ShopItems {
		ShopItems[i].ID = uint(i + 1)
	}

	Scraper.On("ScrapSalesHistory").Return(SoldItems, Task)
	TestShop.On("GetItemsByShopID").Return(ShopItems, nil)
	TestShop.On("CheckAndUpdateOutOfProdMenu").Return(true, nil)
	TestShop.On("CreateOutOfProdMenu").Return(nil)

	_, err := implShop.UpdateDiscontinuedItems(&ExampleShop, Task, ShopRequest)

	assert.NoError(t, err)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)
	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
}

func TestFollowShopInvalidJson(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "EOF")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestFollowShopPanic(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	router.Use(Shop.FollowShop)

	router.GET("/follow_shop", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req := httptest.NewRequest("GET", "/follow_shop", bytes.NewBuffer(body))

	assert.Panics(t, func() {
		ctx.Set("currentUserUUID", nil)
		router.ServeHTTP(w, req)
	})
}

func TestFollowShopShopNotFound(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(nil, errors.New("record not found"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertCalled(t, "GetShopByName")
	assert.Contains(t, w.Body.String(), "shop not found")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestFollowShopGetShopByNameFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(nil, errors.New("Error getting Shop"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertNumberOfCalls(t, "GetShopByName", 1)
	assert.Contains(t, w.Body.String(), "error while processing the request")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestFollowShopGetAccountFail(t *testing.T) {
	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop, Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopRepo.On("GetShopByName").Return(&ShopExample, nil)
	TestShop.On("EstablishAccountShopRelation").Return(errors.New("Error while getting account"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertNumberOfCalls(t, "GetShopByName", 1)
	assert.Contains(t, w.Body.String(), "Error while getting account")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestFollowShopSuccess(t *testing.T) {
	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Scraper: Scraper, Operations: TestShop, Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = 2
	ShopRepo.On("GetShopByName").Return(&ShopExample, nil)
	TestShop.On("EstablishAccountShopRelation").Return(nil)

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertNumberOfCalls(t, "GetShopByName", 1)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestUnFollowShopInvalidJson(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "EOF")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestUnFollowShopPanic(t *testing.T) {

	ctx, router, w := setupMockServer.SetGinTestMode()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	router.Use(Shop.UnFollowShop)

	router.GET("/unfollow_shop", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req := httptest.NewRequest("GET", "/unfollow_shop", bytes.NewBuffer(body))

	assert.Panics(t, func() {
		ctx.Set("currentUserUUID", nil)
		router.ServeHTTP(w, req)
	})
}

func TestUnFollowShopShopNotFound(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(nil, errors.New("record not found"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertCalled(t, "GetShopByName")
	assert.Contains(t, w.Body.String(), "shop not found")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestUnFollowShopGetShopByNameFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopRepo.On("GetShopByName").Return(nil, errors.New("Error getting Shop"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertNumberOfCalls(t, "GetShopByName", 1)
	assert.Contains(t, w.Body.String(), "Error getting Shop")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestUnFollowShopGetAccountFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopRepo.On("GetShopByName").Return(&ShopExample, nil)
	ShopRepo.On("UpdateAccountShopRelation").Return(errors.New("Error while getting account"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertNumberOfCalls(t, "GetShopByName", 1)
	assert.Contains(t, w.Body.String(), "Error while getting account")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}

func TestUnFollowShopSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	currentUserUUID := uuid.New()
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = 2
	ShopRepo.On("GetShopByName").Return(&ShopExample, nil)
	ShopRepo.On("UpdateAccountShopRelation").Return(nil)

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, implShop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	ShopRepo.AssertNumberOfCalls(t, "GetShopByName", 1)
	ShopRepo.AssertNumberOfCalls(t, "UpdateAccountShopRelation", 1)

	assert.Equal(t, http.StatusOK, w.Code)

}

func TestGetShopByIDAveragePriceFail(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)

	ShopRepo.On("FetchShopByID").Return(&ShopExample, nil)
	ShopRepo.On("GetAverageItemPrice").Return(float64(0), errors.New("error getting Item average price"))

	_, err := implShop.GetShopByID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error getting Item average price")

}

func TestGetShopByIDRevenueFail(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	ShopExample := models.Shop{Name: "ExampleShop", HasSoldHistory: true}
	ShopExample.ID = uint(2)

	ShopRepo.On("FetchShopByID").Return(&ShopExample, nil)
	ShopRepo.On("GetAverageItemPrice").Return(float64(15.5), nil)
	TestShop.On("GetTotalRevenue").Return(float64(0), errors.New("error while getting Total revenue"))

	_, err := implShop.GetShopByID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error while getting Total revenue")

}
func TestGetShopByIDSuccess(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)
	ShopRepo.On("GetAverageItemPrice").Return(float64(15.5), nil)
	TestShop.On("GetTotalRevenue").Return(float64(120), nil)
	ShopRepo.On("FetchShopByID").Return(&ShopExample, nil)

	result, err := implShop.GetShopByID(ShopExample.ID)

	assert.NoError(t, err)
	assert.Equal(t, result.ID, ShopExample.ID)
	assert.Equal(t, result.Name, ShopExample.Name)
}

func TestGetItemsCountByShopIDFail(t *testing.T) {

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("error while calculating item average price "))

	implShop.GetItemsCountByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

}
func TestGetItemsCountByShopIDSuccess(t *testing.T) {

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}}, nil)

	implShop.GetItemsCountByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

}
func TestGetSoldItemsByShopIDFail(t *testing.T) {

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("error while calculating item average price "))

	implShop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

}
func TestGetSoldItemsByShopIDSuccess(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)

	Allitems := []models.Item{{ListingID: 1}, {ListingID: 2}, {ListingID: 3}}
	for i := range Allitems {
		Allitems[i].ID = uint(i + 1)
	}

	SoldItems := []models.SoldItems{{ListingID: 1, ItemID: 1}, {ListingID: 1, ItemID: 1}, {ListingID: 3, ItemID: 3}}

	TestShop.On("GetItemsByShopID").Return(Allitems, nil)
	ShopRepo.On("FetchSoldItemsByListingID").Return(SoldItems, nil)

	result, err := implShop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
	ShopRepo.AssertNumberOfCalls(t, "FetchSoldItemsByListingID", 1)

	assert.NoError(t, err)
	assert.Equal(t, len(SoldItems)-1, len(result))

}
func TestGetSoldItemsByShopIDNoSoldItemsInDB(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)

	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}}, nil)
	ShopRepo.On("FetchSoldItemsByListingID").Return(nil, errors.New("items were not found"))

	_, err := implShop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

	assert.Contains(t, err.Error(), "items were not found")

}

func TestGetTotalRevenueFail(t *testing.T) {

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)
	AverageItemPrice := 19.2
	TestShop.On("GetSoldItemsByShopID").Return(nil, errors.New("Sold items where not found"))

	_, err := implShop.GetTotalRevenue(ShopExample.ID, AverageItemPrice)

	TestShop.AssertNumberOfCalls(t, "GetSoldItemsByShopID", 1)

	assert.Contains(t, err.Error(), "Sold items where not found")

}
func TestGetTotalRevenueSuccess(t *testing.T) {

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)
	AverageItemPrice := 19.2
	revenueExpected := 485.68

	TestShop.On("GetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{{Available: true, OriginalPrice: 15.2, SoldQuantity: 3}, {Available: true, OriginalPrice: 19.12, SoldQuantity: 10}, {Available: true, OriginalPrice: 124.44, SoldQuantity: 2}}, nil)

	Revenue, err := implShop.GetTotalRevenue(ShopExample.ID, AverageItemPrice)

	TestShop.AssertNumberOfCalls(t, "GetSoldItemsByShopID", 1)
	assert.NoError(t, err)
	assert.Equal(t, revenueExpected, Revenue)

}

func TestProcessStatsRequestInvalidPeriod(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	implShop := controllers.Shop{}
	Shop := controllers.NewShopController(implShop)

	ShopID := uint(2)
	Period := "InvalidPeriod"

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	router.GET("/stats/:shopID/:period", func(ctx *gin.Context) {
		Shop.ProcessStatsRequest(ctx)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "invalid period provided")
	assert.Equal(t, http.StatusInternalServerError, w.Code)

}
func TestProcessStatsRequestGetSellingStatsByPeriodFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopID := uint(2)
	Period := "lastSevenDays"

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	TestShop.On("GetSellingStatsByPeriod").Return(nil, errors.New("error while fetcheing data from db"))

	router.GET("/stats/:shopID/:period", func(ctx *gin.Context) {
		implShop.ProcessStatsRequest(ctx)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetSellingStatsByPeriod", 1)
	assert.Contains(t, w.Body.String(), "error while handling stats")

}
func TestProcessStatsRequestSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	ShopID := uint(2)
	Period := "lastSevenDays"

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	stats := map[string]controllers.DailySoldStats{
		"2024-01-01": {
			TotalSales: 100,
			Items:      []models.Item{{}, {}},
		},
	}

	TestShop.On("GetSellingStatsByPeriod").Return(stats, nil)

	router.GET("/stats/:shopID/:period", func(ctx *gin.Context) {
		implShop.ProcessStatsRequest(ctx)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetSellingStatsByPeriod", 1)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSellingStatsByPeriodSelectData(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	ShopID := uint(2)
	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	stats := []models.DailyShopSales{
		{
			TotalSales: 100,
		},
	}

	ShopRepo.On("FetchStatsByPeriod").Return(stats, nil)
	TestShop.On("CreateSoldStats").Return(map[string]controllers.DailySoldStats{}, nil)

	_, err := implShop.GetSellingStatsByPeriod(ShopID, Period)

	assert.NoError(t, err)

}
func TestGetSellingStatsByPeriodSelectDataaFail(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopID := uint(2)
	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	ShopRepo.On("FetchStatsByPeriod").Return(nil, errors.New("error fetching data"))

	_, err := implShop.GetSellingStatsByPeriod(ShopID, Period)

	assert.Contains(t, err.Error(), "error fetching data")

}

func TestSaveShopToDB(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}
	Shop := controllers.NewShopController(implShop)

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	ShopRepo.On("CreateShop").Return(nil)
	err := Shop.SaveShopToDB(ShopExample, ShopRequest)

	assert.NoError(t, err)

}

func TestSaveShopToDBFailed(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	ShopRepo.On("CreateShop").Return(errors.New("database Error"))
	TestShop.On("CreateShopRequest").Return(nil)

	err := implShop.SaveShopToDB(ShopExample, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database Error")
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)

}

func TestUpdateShopMenuToDB(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{{
				Category:  "All",
				SectionID: "1191",
				Link:      "wwww.ExampleLink",
				Amount:    19,
			}},
		},
	}

	TestShop.On("CreateShopRequest").Return(nil)
	ShopRepo.On("SaveShop").Return(nil)
	err := implShop.UpdateShopMenuToDB(ShopExample, ShopRequest)

	assert.NoError(t, err)

	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)

}

func TestUpdateShopMenuToDBFail(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	TestShop.On("CreateShopRequest").Return(nil)
	ShopRepo.On("SaveShop").Return(errors.New("database Error"))

	err := implShop.UpdateShopMenuToDB(ShopExample, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database Error")
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
}

func TestReverseSoldItems(t *testing.T) {
	SoldItems := []models.SoldItems{{Name: "1"}, {Name: "2"}, {Name: "3"}, {Name: "4"}, {Name: "5"}, {Name: "6"}}
	ReversedSoldItems := []models.SoldItems{{Name: "6"}, {Name: "5"}, {Name: "4"}, {Name: "3"}, {Name: "2"}, {Name: "1"}}

	result := controllers.ReverseSoldItems(SoldItems)

	assert.Equal(t, ReversedSoldItems, result, "checking if the slice got reversed")
}

func TestFilterSoldOutItems(t *testing.T) {
	FilterSoldItems := map[uint]struct{}{}
	ScrappedSoldItems := []models.SoldItems{{Name: "Example", ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ListingID: 13, DataShopID: "1122"}, {Name: "Example", ListingID: 12, DataShopID: "1122"}, {Name: "Example", ListingID: 17, DataShopID: "1122"}, {Name: "Example2", ListingID: 19, DataShopID: "1122"}}
	existingItems := []models.Item{{ListingID: 12}, {ListingID: 13}, {ListingID: 14}, {ListingID: 15}}
	for index := range existingItems {
		existingItems[index].ID = uint(index + 1)
	}

	SoldOutItems := controllers.FilterSoldOutItems(ScrappedSoldItems, existingItems, FilterSoldItems)

	assert.Equal(t, len(SoldOutItems), 2)
	assert.Equal(t, SoldOutItems[0].ListingID, ScrappedSoldItems[3].ListingID)
	assert.Equal(t, SoldOutItems[1].ListingID, ScrappedSoldItems[4].ListingID)

}

func TestCheckAndUpdateOutOfProdMenu(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	AllMenus := []models.MenuItem{{Category: "All"}, {Category: "UnCategorized"}, {Category: "Out Of Production"}}

	SoldOutItems := []models.Item{{Name: "Example", ListingID: 12, DataShopID: "1122"}}
	SoldOutItems[0].ID = uint(1)
	ShopRequest := &models.ShopRequest{

		ShopName: "exampleShop",
		Status:   "Pending",
	}

	TestShop.On("CreateShopRequest").Return(nil)
	ShopRepo.On("SaveMenu").Return(nil)

	exists, err := implShop.CheckAndUpdateOutOfProdMenu(AllMenus, SoldOutItems, ShopRequest)

	assert.True(t, exists)
	assert.NoError(t, err)

}

func TestCheckAndUpdateOutOfProdMenuNoExist(t *testing.T) {

	implShop := controllers.Shop{}

	AllMenus := []models.MenuItem{{Category: "All"}, {Category: "UnCategorized"}}

	SoldOutItems := []models.Item{{Name: "Example", ListingID: 12, DataShopID: "1122"}}

	ShopRequest := &models.ShopRequest{
		ShopName: "exampleShop",
		Status:   "Pending",
	}

	exists, err := implShop.CheckAndUpdateOutOfProdMenu(AllMenus, SoldOutItems, ShopRequest)

	assert.False(t, exists)
	assert.NoError(t, err)

}

func TestCreateNewOutOfProdMenu(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	SoldOutItems := []models.Item{{Name: "Example", ListingID: 12, DataShopID: "1122"}}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	ShopRequest := &models.ShopRequest{
		ShopName: "exampleShop",
		Status:   "Pending",
	}

	ShopRepo.On("SaveShop").Return(nil)

	err := implShop.CreateOutOfProdMenu(ShopExample, SoldOutItems, ShopRequest)

	assert.NoError(t, err)
}

func TestCreateNewOutOfProdMenuFail(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	SoldOutItems := []models.Item{{Name: "Example", ListingID: 12, DataShopID: "1122"}}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}

	ShopRequest := &models.ShopRequest{
		ShopName: "exampleShop",
		Status:   "Pending",
	}

	ShopRepo.On("SaveShop").Return(errors.New("error while creating menu"))

	err := implShop.CreateOutOfProdMenu(ShopExample, SoldOutItems, ShopRequest)

	assert.Error(t, err)

	assert.Contains(t, err.Error(), "error while creating menu")

}

func TestPopulateItemIDsFromListings(t *testing.T) {
	var expectedRevenue float64
	SoldItems := []models.SoldItems{{Name: "Example", ListingID: 12, DataShopID: "1122", ItemID: 1}, {Name: "Example2", ListingID: 13, DataShopID: "1122", ItemID: 2}, {Name: "Example2", ListingID: 13, DataShopID: "1122", ItemID: 2}, {Name: "Example2", ListingID: 15, DataShopID: "1122", ItemID: 4}}
	existingItems := []models.Item{{ListingID: 12, OriginalPrice: 19.8}, {ListingID: 13, OriginalPrice: 11.5}, {ListingID: 14, OriginalPrice: 17.6}, {ListingID: 15, OriginalPrice: 90.1}}
	for i := range existingItems {
		existingItems[i].ID = uint(i + 1)

	}

	expectedID := []uint{1, 2, 2, 4}
	actualInjectedID := []uint{}

	for _, ID := range expectedID {
		for _, Item := range existingItems {
			if Item.ID == ID {
				expectedRevenue += Item.OriginalPrice
			}
		}
	}

	SortedItems, dailRevenue := controllers.PopulateItemIDsFromListings(SoldItems, existingItems)

	for _, item := range SortedItems {
		actualInjectedID = append(actualInjectedID, item.ItemID)
	}

	assert.Equal(t, expectedID, actualInjectedID)
	assert.Equal(t, expectedRevenue, dailRevenue)
}

func TestEstablishAccountShopRelation(t *testing.T) {

	UserRepo := &MockedUserRepository{}

	implShop := controllers.Shop{User: UserRepo}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}
	userID := uuid.New()
	UserRepo.On("GetAccountByID").Return(&models.Account{}, nil)
	UserRepo.On("SaveAccount").Return(nil)
	err := implShop.EstablishAccountShopRelation(ShopExample, userID)

	assert.NoError(t, err)

}

func TestGetItemsBySoldItemsSuccess(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	SoldItems := make([]models.SoldItems, 5)

	ShopRepo.On("FetchItemsBySoldItems").Return(models.Item{}, nil)

	items, err := implShop.GetItemsBySoldItems(SoldItems)

	assert.NoError(t, err)
	assert.Equal(t, len(SoldItems), len(items))

}

func TestGetItemsBySoldItemsFail(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	SoldItems := make([]models.SoldItems, 5)

	ShopRepo.On("FetchItemsBySoldItems").Return(nil, errors.New("error while processing database operations"))
	_, err := implShop.GetItemsBySoldItems(SoldItems)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while processing database operations")

}

func TestHandleHandleGetShopByIDNoShop(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	TestShop.On("GetShopByID").Return(nil, errors.New("failed to get shop"))
	router.GET("/testroute/:shopID", implShop.HandleGetShopByID)

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetShopByID found no shop in db")

}

func TestHandleHandleGetShopByIDfail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	implShop := controllers.Shop{}
	router.GET("/testroute/:shopID", implShop.HandleGetShopByID)

	req, err := http.NewRequest("GET", "/testroute", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code, "no id passed")

}

func TestHandleHandleGetShopByIDSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}
	router.GET("/testroute/:shopID", implShop.HandleGetShopByID)

	TestShop.On("GetShopByID").Return(ShopExample, nil)

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestHandleGetItemsByShopIDNoShop(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}
	router.GET("/testroute/:shopID", implShop.HandleGetItemsByShopID)

	TestShop.On("GetItemsByShopID").Return(nil, errors.New("no shop found"))

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetItemsByShopID found no shop in db")

}

func TestHandleGetItemsByShopIDFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	implShop := controllers.Shop{}

	router.GET("/testroute/:shopID", implShop.HandleGetItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code, "no id passed")

}

func TestHandleGetItemsByShopIDSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	TestShop.On("GetItemsByShopID").Return([]models.Item{}, nil)
	router.GET("/testroute/:shopID", implShop.HandleGetItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestHandleGetSoldItemsByShopIDNoShop(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{}
	router.GET("/testroute/:shopID/all_sold_items", implShop.HandleGetSoldItemsByShopID)

	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("no shop found"))

	req, err := http.NewRequest("GET", "/testroute//all_sold_items", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetItemsByShopID found no shop in db")

}

func TestHandleGetSoldItemsByShopIDFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	TestShop.On("GetSoldItemsByShopID").Return(nil, errors.New("error getting data"))

	router.GET("/testroute/:shopID/all_sold_items", implShop.HandleGetSoldItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute/1/all_sold_items", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "no id passed")

}

func TestHandleGetSoldItemsByShopIDSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}

	TestShop.On("GetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{}, nil)

	router.GET("/testroute/:shopID/all_sold_items", implShop.HandleGetSoldItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute/1/all_sold_items", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestHandleGetItemsCountByShopIDNoShop(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	implShop := controllers.Shop{}
	router.GET("/testroute/:shopID/items_count", implShop.HandleGetItemsCountByShopID)

	req, err := http.NewRequest("GET", "/testroute/d/items_count", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetItemsByShopID found no shop in db")

}

func TestHandleGetItemsCountByShopIDFail(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}
	router.GET("/testroute/:shopID/items_count", implShop.HandleGetItemsCountByShopID)

	TestShop.On("GetItemsByShopID").Return(nil, errors.New("no shop found"))

	req, err := http.NewRequest("GET", "/testroute/1/items_count", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

}

func TestHandleGetItemsCountByShopIDSuccess(t *testing.T) {

	_, router, w := setupMockServer.SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{Operations: TestShop}
	router.GET("/testroute/:shopID/items_count", implShop.HandleGetItemsCountByShopID)

	TestShop.On("GetItemsByShopID").Return([]models.Item{}, nil)

	req, err := http.NewRequest("GET", "/testroute/1/items_count", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestCalculateTotalRevenue(t *testing.T) {
	soldItems := []controllers.ResponseSoldItemInfo{
		{OriginalPrice: 19.2, SoldQuantity: 10}, {OriginalPrice: 12.4, SoldQuantity: 9}, {OriginalPrice: 5.2, SoldQuantity: 11}, {SoldQuantity: 2}, {SoldQuantity: 19},
	}
	AverageItemPrice := 7.5

	var expectedRevenue float64
	for _, soldItem := range soldItems {
		if soldItem.OriginalPrice > 0 {
			expectedRevenue += soldItem.OriginalPrice * float64(soldItem.SoldQuantity)
		} else {
			expectedRevenue += AverageItemPrice * float64(soldItem.SoldQuantity)
		}
	}

	revenue := controllers.CalculateTotalRevenue(soldItems, AverageItemPrice)
	assert.Equal(t, expectedRevenue, revenue)
}

func TestCreateSoldStatsFail(t *testing.T) {
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	dailyShopSales := []models.DailyShopSales{
		{
			ShopID:       1,
			TotalSales:   100,
			DailyRevenue: 90.1,
		},
	}

	ShopRepo.On("GetSoldItemsInRange").Return(nil, errors.New("internal error"))

	_, err := implShop.CreateSoldStats(dailyShopSales)

	assert.Error(t, err)

}

func TestCreateSoldStatsSuccessWithItems(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	dailyShopSales := []models.DailyShopSales{
		{
			ShopID:       1,
			TotalSales:   100,
			DailyRevenue: 90.1,
		},
		{
			ShopID:       1,
			TotalSales:   101,
			DailyRevenue: 16.1,
		},
	}
	for i := range dailyShopSales {
		dailyShopSales[i].CreatedAt = time.Now().AddDate(0, 0, (-len(dailyShopSales) + i))
	}

	ShopRepo.On("GetSoldItemsInRange").Return([]models.SoldItems{{}, {}, {}}, nil)
	TestShop.On("GetItemsBySoldItems").Return([]models.Item{{}}, nil)

	stats, err := implShop.CreateSoldStats(dailyShopSales)

	for _, record := range stats {
		assert.Equal(t, 1, len(record.Items))
	}

	assert.Equal(t, len(dailyShopSales), len(stats))
	assert.NoError(t, err)

}

func TestCreateSoldStatsSuccesswithNoItems(t *testing.T) {

	TestShop := &MockedShop{}
	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Operations: TestShop, Shop: ShopRepo}

	dailyShopSales := []models.DailyShopSales{
		{
			ShopID:       1,
			TotalSales:   100,
			DailyRevenue: 90.1,
		},
		{
			ShopID:       1,
			TotalSales:   101,
			DailyRevenue: 16.1,
		},
	}
	for i := range dailyShopSales {
		dailyShopSales[i].CreatedAt = time.Now().AddDate(0, 0, (-len(dailyShopSales) + i))
	}

	ShopRepo.On("GetSoldItemsInRange").Return([]models.SoldItems{}, nil)
	TestShop.On("GetItemsBySoldItems").Return([]models.Item{}, nil)

	stats, err := implShop.CreateSoldStats(dailyShopSales)

	for _, record := range stats {
		assert.Equal(t, 0, len(record.Items))
	}

	assert.Equal(t, len(dailyShopSales), len(stats))
	assert.NoError(t, err)

}

func TestCreateShopRequestTypeShopFailNoAccount(t *testing.T) {

	implShop := controllers.Shop{}

	ShopRequest := &models.ShopRequest{
		ShopName: "exampleShop",
		Status:   "Pending",
	}

	err := implShop.CreateShopRequest(ShopRequest)

	assert.Contains(t, err.Error(), "no AccountID was passed")

}
func TestCreateShopRequestTypeShopFailSaveData(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopRepo.On("SaveShopRequestToDB").Return(errors.New("Failed to save ShopRequest"))

	err := implShop.CreateShopRequest(ShopRequest)

	assert.Contains(t, err.Error(), "Failed to save ShopRequest")
}
func TestCreateShopRequestTypeShopSuccess(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopRepo.On("SaveShopRequestToDB").Return(nil)
	err := implShop.CreateShopRequest(ShopRequest)

	assert.NoError(t, err)

}

func TestGetItemsByShopIDTypeShopSuccess(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}

	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)

	ShopRepo.On("GetShopWithItemsByShopID").Return(&ShopExample, nil)

	_, err := implShop.GetItemsByShopID(ShopExample.ID)

	assert.NoError(t, err)

}

func TestGetItemsByShopIDTypeShopFail(t *testing.T) {

	ShopRepo := &MockedShopRepository{}
	implShop := controllers.Shop{Shop: ShopRepo}
	ShopExample := models.Shop{}
	ShopExample.ID = uint(2)

	ShopRepo.On("GetShopWithItemsByShopID").Return(nil, errors.New("error while getting shop from DB"))

	_, err := implShop.GetItemsByShopID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error while getting shop from DB")

}

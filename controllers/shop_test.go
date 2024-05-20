package controllers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"EtsyScraper/controllers"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	setupMockServer "EtsyScraper/setupTests"
	"EtsyScraper/utils"
)

type MockedShop struct {
	mock.Mock
}

func (m *MockedShop) GetShopByName(ShopName string) (*models.Shop, error) {

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

func (m *MockedShop) GetItemsByShopID(ID uint) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Items []models.Item
	if shopInterface != nil {
		Items = shopInterface.([]models.Item)
	}
	return Items, args.Error(1)
}

func (m *MockedShop) GetAverageItemPrice(ShopID uint) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockedShop) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedShop) GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
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

func (m *MockedShop) GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Stats map[string]controllers.DailySoldStats
	if shopInterface != nil {
		Stats = shopInterface.(map[string]controllers.DailySoldStats)
	}
	return Stats, args.Error(1)
}

func (m *MockedShop) ExecuteCreateShop(dispatch controllers.ShopOperations, ShopRequest *models.ShopRequest) {

}

func (m *MockedShop) ExecuteGetTotalRevenue(dispatch controllers.ShopOperations, ShopID uint, AverageItemPrice float64) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockedShop) ExecuteGetAverageItemPrice(dispatch controllers.ShopOperations, ShopID uint) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockedShop) ExecuteCreateShopRequest(dispatch controllers.ShopOperations, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedShop) ExecuteGetItemsByShopID(dispatch controllers.ShopOperations, ID uint) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Items []models.Item
	if shopInterface != nil {
		Items = shopInterface.([]models.Item)
	}
	return Items, args.Error(1)
}

func (m *MockedShop) ExecuteGetShopByName(dispatch controllers.ShopOperations, ShopName string) (*models.Shop, error) {

	args := m.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}

func (m *MockedShop) ExecuteGetSoldItemsByShopID(dispatch controllers.ShopOperations, ID uint) (SoldItemInfos []controllers.ResponseSoldItemInfo, err error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []controllers.ResponseSoldItemInfo
	if shopInterface != nil {
		soldItems = shopInterface.([]controllers.ResponseSoldItemInfo)
	}
	return soldItems, args.Error(1)
}
func (m *MockedShop) ExecuteUpdateSellingHistory(controllers.ShopUpdater, *models.Shop, *models.TaskSchedule, *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) ExecuteUpdateDiscontinuedItems(dispatch controllers.ShopUpdater, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []models.SoldItems
	if shopInterface != nil {
		soldItems = shopInterface.([]models.SoldItems)
	}
	return soldItems, args.Error(1)
}

func (m *MockedShop) ExecuteGetSellingStatsByPeriod(dispatch controllers.ShopOperations, ShopID uint, timePeriod time.Time) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Stats map[string]controllers.DailySoldStats
	if shopInterface != nil {
		Stats = shopInterface.(map[string]controllers.DailySoldStats)
	}
	return Stats, args.Error(1)
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

func TestCreateNewShopRequest_Panic(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	router.Use(Shop.CreateNewShopRequest)

	router.GET("/create_shop", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/create_shop", nil)

	assert.Panics(t, func() {
		ctx.Set("currentUserUUID", nil)
		router.ServeHTTP(w, req)
	})

}

func TestCreateNewShopRequest_InvalidJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte{}
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "failed to get the Shop's name")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}
func TestCreateNewShopRequest_GetShopError(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(nil, errors.New("Error"))
	TestShop.On("ExecuteCreateShopRequest").Return(errors.New("SecondError"))

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "ExecuteCreateShopRequest")
	assert.Contains(t, w.Body.String(), "internal error")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}
func TestCreateNewShopRequest_ShopExists(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(&models.Shop{Name: "ShopName"}, nil)
	TestShop.On("ExecuteCreateShopRequest").Return(errors.New("SecondError"))

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "ExecuteCreateShopRequest")
	assert.Contains(t, w.Body.String(), "Shop already exists")
	assert.Equal(t, w.Code, http.StatusBadRequest)

}
func TestCreateNewShopRequest_Success(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(nil, errors.New("no Shop was Found ,error: record not found"))
	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "ExecuteCreateShopRequest")
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.Contains(t, w.Body.String(), "shop request received successfully")
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestCreateNewShop_ScrapperErr(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
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
func TestCreateNewShop_FailedSaveShopToDB(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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
	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("Failed to save shop"))
	sqlMock.ExpectRollback()

	err := Shop.CreateNewShop(ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save shop")
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestCreateNewShop_SaveMenuToDB_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("failed to save new record"))
	sqlMock.ExpectRollback()

	err := Shop.CreateNewShop(ShopRequest)

	assert.Error(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestCreateNewShop_SaveMenuToDB_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	userID := uuid.New()
	ShopRequest := &models.ShopRequest{
		AccountID: userID,
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()
	sqlMock.ExpectBegin()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()

	err := Shop.CreateNewShop(ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 2)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestCreateNewShop_HasSoldHistory(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
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

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)
	TestShop.On("ExecuteUpdateSellingHistory").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 10, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()
	sqlMock.ExpectBegin()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 10, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()

	err := Shop.CreateNewShop(ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateSellingHistory", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateSellingHistory_DisContintuesSoldItemsFail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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

	TestShop.On("ExecuteUpdateDiscontinuedItems").Return(nil, errors.New("failed to get SoldItems"))
	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)

}
func TestUpdateSellingHistory_DisContintuesSoldItemsEmpty(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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

	TestShop.On("ExecuteUpdateDiscontinuedItems").Return([]models.SoldItems{}, nil)

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty scrapped Sold data")
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)

}
func TestUpdateSellingHistory_GetItemsFail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}

	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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

	TestShop.On("ExecuteUpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}, {}}, nil)
	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("error getting Items"))

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting Items")
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)

}
func TestUpdateSellingHistory_InsertIntoDBFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}

	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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

	TestShop.On("ExecuteUpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}}, nil)
	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("failed to insert data to DB"))
	sqlMock.ExpectRollback()

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert data to DB")
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestUpdateSellingHistory_InsertIntoDB(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}

	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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

	TestShop.On("ExecuteUpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}}, nil)
	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestUpdateSellingHistory_TaskSoldItem(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}

	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

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

	TestShop.On("ExecuteUpdateDiscontinuedItems").Return([]models.SoldItems{{}, {}}, nil)
	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`updated_at"=$1 WHERE created_at > $2 AND shop_id = $3 AND "daily_shop_sales"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 14))
	sqlMock.ExpectCommit()

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateDiscontinuedItems_EmptySoldItems(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
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
func TestUpdateDiscontinuedItems_GetItemsByShopID_fail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
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

	Scraper.On("ScrapSalesHistory").Return([]models.SoldItems{{}, {}, {}}, Task)
	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("Error While fetching Shop's details"))

	_, err := Shop.UpdateDiscontinuedItems(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)
	assert.Contains(t, err.Error(), "Error While fetching Shop's details")
}
func TestUpdateDiscontinuedItems_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper, Process: TestShop}
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
	TestShop.On("ExecuteGetItemsByShopID").Return(ShopItems, nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_menus" ("created_at","updated_at","deleted_at","shop_id","total_items_amount") VALUES ($1,$2,$3,$4,$5) ON CONFLICT ("id") DO UPDATE SET "shop_id"="excluded"."shop_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9),($10,$11,$12,$13,$14,$15,$16,$17,DEFAULT) ON CONFLICT ("id") DO UPDATE SET "shop_menu_id"="excluded"."shop_menu_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT ("id") DO UPDATE SET "menu_item_id"="excluded"."menu_item_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectCommit()

	_, err := Shop.UpdateDiscontinuedItems(&ExampleShop, Task, ShopRequest)

	assert.NoError(t, err)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)
	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestFollowShop_InvalidJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
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

func TestFollowShop_Panic(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
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

func TestFollowShop_ShopNotFound(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(nil, errors.New("record not found"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "ExecuteGetShopByName")
	assert.Contains(t, w.Body.String(), "shop not found")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestFollowShop_GetShopByNameFail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(nil, errors.New("Error getting Shop"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetShopByName", 1)
	assert.Contains(t, w.Body.String(), "error while processing the request")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestFollowShop_GetAccountFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{}
	TestShop.On("ExecuteGetShopByName").Return(&ShopExample, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnError(errors.New("Error while getting account"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetShopByName", 1)
	assert.Contains(t, w.Body.String(), "Error while getting account")
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestFollowShop_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	time := time.Now()
	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{}
	ShopExample.ID = 2
	TestShop.On("ExecuteGetShopByName").Return(&ShopExample, nil)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), time, time, time, "Testing", "User", "", "", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnRows(Account)

	sqlMock.ExpectBegin()

	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), currentUserUUID.String()).WillReturnResult(sqlmock.NewResult(1, 1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT DO NOTHING RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "00000000-0000-0000-0000-000000000000", ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "account_shop_following" ("account_id","shop_id") VALUES ($1,$2) ON CONFLICT DO NOTHING`)).
		WithArgs(currentUserUUID.String(), 2).WillReturnResult(sqlmock.NewResult(1, 2))

	sqlMock.ExpectCommit()
	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetShopByName", 1)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUnFollowShop_InvalidJson(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
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

func TestUnFollowShop_Panic(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ctx, router, w := SetGinTestMode()
	Scraper := &MockScrapper{}
	implShop := controllers.Shop{DB: MockedDataBase, Scraper: Scraper}
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

func TestUnFollowShop_ShopNotFound(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(nil, errors.New("record not found"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "ExecuteGetShopByName")
	assert.Contains(t, w.Body.String(), "shop not found")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestUnFollowShop_GetShopByNameFail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	TestShop.On("ExecuteGetShopByName").Return(nil, errors.New("Error getting Shop"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetShopByName", 1)
	assert.Contains(t, w.Body.String(), "Error getting Shop")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
func TestUnFollowShop_GetAccountFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{}
	TestShop.On("ExecuteGetShopByName").Return(&ShopExample, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnError(errors.New("Error while getting account"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetShopByName", 1)
	assert.Contains(t, w.Body.String(), "Error while getting account")
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUnFollowShop_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	time := time.Now()
	currentUserUUID := uuid.New()
	Scraper := &scrap.Scraper{}
	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop, Scraper: Scraper}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{}
	ShopExample.ID = 2
	TestShop.On("ExecuteGetShopByName").Return(&ShopExample, nil)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), time, time, time, "Testing", "User", "", "", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnRows(Account)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1`)).
		WithArgs(currentUserUUID.String()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1 AND "account_shop_following"."shop_id" = $2`)).
		WithArgs(currentUserUUID.String(), 2).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlMock.ExpectCommit()

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetShopByName", 1)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetShopByID_AveragePriceFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetAverageItemPrice").Return(float64(0), errors.New("error getting Item average price"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "name"}).AddRow(10, ShopExample.ID, "Owner"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "ShopRating"}).AddRow(5, ShopExample.ID, 4.4))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews_topics" WHERE "reviews_topics"."reviews_id"`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 5, "Test1").AddRow(7, 5, "Test2"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	_, err := Shop.GetShopByID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error getting Item average price")
	assert.Error(t, sqlMock.ExpectationsWereMet())
}

func TestGetShopByID_RevenueFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetAverageItemPrice").Return(float64(15.5), nil)
	TestShop.On("ExecuteGetTotalRevenue").Return(float64(0), errors.New("Error while getting Total revenue"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "has_sold_history"}).AddRow(ShopExample.ID, ShopExample.Name, true))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "name"}).AddRow(10, ShopExample.ID, "Owner"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "ShopRating"}).AddRow(5, ShopExample.ID, 4.4))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews_topics" WHERE "reviews_topics"."reviews_id"`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 5, "Test1").AddRow(7, 5, "Test2"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	_, err := Shop.GetShopByID(ShopExample.ID)

	assert.Contains(t, err.Error(), "Error while getting Total revenue")
	assert.Error(t, sqlMock.ExpectationsWereMet())
}
func TestGetShopByID_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetAverageItemPrice").Return(float64(15.5), nil)
	TestShop.On("ExecuteGetTotalRevenue").Return(float64(120), nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "name"}).AddRow(10, ShopExample.ID, "Owner"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "ShopRating"}).AddRow(5, ShopExample.ID, 4.4))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews_topics" WHERE "reviews_topics"."reviews_id"`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 5, "Test1").AddRow(7, 5, "Test2"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	Shop.GetShopByID(ShopExample.ID)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetItemsCountByShopID_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("error while calculating item average price "))

	Shop.GetItemsCountByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetItemsCountByShopID_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{{}, {}}, nil)

	Shop.GetItemsCountByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetSoldItemsByShopID_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("error while calculating item average price "))

	Shop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetSoldItemsByShopID_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	Allitems := []models.Item{{ListingID: 1}, {ListingID: 2}, {ListingID: 3}}
	for i := range Allitems {
		Allitems[i].ID = uint(i + 1)
	}

	Solditems := sqlmock.NewRows([]string{"id", "listingID", "ItemID"}).AddRow(1, 1, 1).AddRow(2, 1, 1).AddRow(3, 3, 3)

	TestShop.On("ExecuteGetItemsByShopID").Return(Allitems, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sold_items" WHERE listing_id IN ($1,$2,$3) AND "sold_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(Solditems)

	Shop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetSoldItemsByShopID_NoSoldItemsInDB(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{{}, {}}, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sold_items" WHERE listing_id IN ($1,$2) AND "sold_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnError(errors.New("items were not found"))

	_, err := Shop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetItemsByShopID", 1)

	assert.Contains(t, err.Error(), "items were not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetTotalRevenue_Fail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	AverageItemPrice := 19.2
	TestShop.On("ExecuteGetSoldItemsByShopID").Return(nil, errors.New("Sold items where not found"))

	_, err := Shop.GetTotalRevenue(ShopExample.ID, AverageItemPrice)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSoldItemsByShopID", 1)

	assert.Contains(t, err.Error(), "Sold items where not found")

}
func TestGetTotalRevenue_Success(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	AverageItemPrice := 19.2
	revenueExpected := 485.68

	TestShop.On("ExecuteGetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{{Available: true, OriginalPrice: 15.2, SoldQuantity: 3}, {Available: true, OriginalPrice: 19.12, SoldQuantity: 10}, {Available: true, OriginalPrice: 124.44, SoldQuantity: 2}}, nil)

	Revenue, err := Shop.GetTotalRevenue(ShopExample.ID, AverageItemPrice)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSoldItemsByShopID", 1)
	assert.NoError(t, err)
	assert.Equal(t, revenueExpected, Revenue)

}

func TestSoldItemsTask(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	AverageItemPrice := 19.2
	revenueExpected := 485.68

	TestShop.On("ExecuteGetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{{Available: true, OriginalPrice: 15.2, SoldQuantity: 3}, {Available: true, OriginalPrice: 19.12, SoldQuantity: 10}, {Available: true, OriginalPrice: 124.44, SoldQuantity: 2}}, nil)

	Revenue, err := Shop.GetTotalRevenue(ShopExample.ID, AverageItemPrice)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSoldItemsByShopID", 1)
	assert.NoError(t, err)
	assert.Equal(t, revenueExpected, Revenue)
}

func TestProcessStatsRequest_InvalidPeriod(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	implShop := controllers.Shop{DB: MockedDataBase}
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
func TestProcessStatsRequest_GetSellingStatsByPeriod_Fail(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopID := uint(2)
	Period := "lastSevenDays"

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	TestShop.On("ExecuteGetSellingStatsByPeriod").Return(nil, errors.New("error while fetcheing data from db"))

	router.GET("/stats/:shopID/:period", func(ctx *gin.Context) {
		Shop.ProcessStatsRequest(ctx)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSellingStatsByPeriod", 1)
	assert.Contains(t, w.Body.String(), "error while handling stats")

}
func TestProcessStatsRequest_Success(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopID := uint(2)
	Period := "lastSevenDays"

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	stats := map[string]controllers.DailySoldStats{
		"2024-01-01": {
			TotalSales: 100,
			Items:      []models.Item{{}, {}},
		},
	}

	TestShop.On("ExecuteGetSellingStatsByPeriod").Return(stats, nil)

	router.GET("/stats/:shopID/:period", func(ctx *gin.Context) {
		Shop.ProcessStatsRequest(ctx)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSellingStatsByPeriod", 1)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSellingStatsByPeriod_SelectData(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}
	Shop := controllers.NewShopController(implShop)

	ShopID := uint(2)
	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	DailySales := sqlmock.NewRows([]string{"id", "created_at", "ShopID", "TotalSales"}).AddRow(1, now.AddDate(0, 0, -3), ShopID, 90).AddRow(2, now.AddDate(0, 0, -4), ShopID, 95)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "daily_shop_sales" WHERE (shop_id = $1 AND created_at > $2) AND "daily_shop_sales"."deleted_at" IS NULL`)).WithArgs(ShopID, Period).WillReturnRows(DailySales)

	Shop.GetSellingStatsByPeriod(ShopID, Period)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetSellingStatsByPeriod_SelectDataaFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}
	Shop := controllers.NewShopController(implShop)

	ShopID := uint(2)
	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "daily_shop_sales" WHERE (shop_id = $1 AND created_at > $2) AND "daily_shop_sales"."deleted_at" IS NULL`)).WillReturnError(errors.New("error fetching data"))

	_, err := Shop.GetSellingStatsByPeriod(ShopID, Period)

	assert.Contains(t, err.Error(), "error fetching data")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetSellingStatsByPeriod_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}
	Shop := controllers.NewShopController(implShop)

	ShopID := uint(2)

	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	DailySales := sqlmock.NewRows([]string{"id", "created_at", "ShopID", "TotalSales"}).AddRow(1, now.AddDate(0, 0, -3), ShopID, 90)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "daily_shop_sales" WHERE (shop_id = $1 AND created_at > $2) AND "daily_shop_sales"."deleted_at" IS NULL`)).WithArgs(ShopID, Period).WillReturnRows(DailySales)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 1).AddRow(2, 2))
	for i := 1; i <= 2; i++ {
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).WillReturnRows(sqlmock.NewRows([]string{"ID", "ListingID"}).AddRow(i, 12).AddRow(i, 13))

	}

	itemsCount := 0
	stats, _ := Shop.GetSellingStatsByPeriod(ShopID, Period)
	for _, value := range stats {
		itemsCount += len(value.Items)
	}

	assert.Equal(t, 1, len(stats))
	assert.Equal(t, 2, itemsCount)
	assert.IsTypef(t, map[string]controllers.DailySoldStats{}, stats, "GetSellingStatsByPeriod return map[string]controllers.DailySoldStats type")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSaveShopToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}
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

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := Shop.SaveShopToDB(ShopExample, ShopRequest)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSaveShopToDB_Failed(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

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

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).WillReturnError(errors.New("database Error"))
	sqlMock.ExpectRollback()

	err := implShop.SaveShopToDB(ShopExample, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database Error")
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateShopMenuToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

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

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_menus" ("created_at","updated_at","deleted_at","shop_id","total_items_amount") VALUES ($1,$2,$3,$4,$5) ON CONFLICT ("id") DO UPDATE SET "shop_id"="excluded"."shop_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 0).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT ("id") DO UPDATE SET "shop_menu_id"="excluded"."shop_menu_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, "All", "1191", "wwww.ExampleLink", 19).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectCommit()
	err := implShop.UpdateShopMenuToDB(ShopExample, ShopRequest)

	assert.NoError(t, err)

	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateShopMenuToDB_Fail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

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

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("database Error"))
	sqlMock.ExpectRollback()

	err := implShop.UpdateShopMenuToDB(ShopExample, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database Error")
	TestShop.AssertNumberOfCalls(t, "ExecuteCreateShopRequest", 1)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReverseSoldItems(t *testing.T) {
	SoldItems := []models.SoldItems{{Name: "1"}, {Name: "2"}, {Name: "3"}, {Name: "4"}, {Name: "5"}, {Name: "6"}}
	ReversedSoldItems := []models.SoldItems{{Name: "6"}, {Name: "5"}, {Name: "4"}, {Name: "3"}, {Name: "2"}, {Name: "1"}}

	result := controllers.ReverseSoldItems(SoldItems)

	assert.Equal(t, ReversedSoldItems, result, "checking if the slice got reversed")
}

func TestSaveSoldItemsToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 12, "1122", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 2, 13, "1122").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := implShop.SaveSoldItemsToDB(SoldItems)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSaveSoldItemsToDB_Fail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).WillReturnError(errors.New("error while saving sold item"))
	sqlMock.ExpectRollback()

	err := implShop.SaveSoldItemsToDB(SoldItems)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while saving sold item")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateDailySales_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}

	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ExampleShopID := uint(10)
	dailyRevenue := 98.9
	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "daily_shop_sales" SET "updated_at"=$1,"daily_revenue"=$2 WHERE created_at > $3 AND shop_id = $4 AND "daily_shop_sales"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), dailyRevenue, sqlmock.AnyArg(), ExampleShopID).WillReturnResult(sqlmock.NewResult(1, 3))
	sqlMock.ExpectCommit()

	err := Shop.UpdateDailySales(SoldItems, ExampleShopID, dailyRevenue)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateDailySales_Failed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}

	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ExampleShopID := uint(10)
	dailyRevenue := 98.9
	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "daily_shop_sales" SET "updated_at"=$1,"daily_revenue"=$2 WHERE created_at > $3 AND shop_id = $4 AND "daily_shop_sales"."deleted_at" IS NULL`)).
		WillReturnError(errors.New("error while saving data to dailyShopSales"))
	sqlMock.ExpectRollback()

	err := Shop.UpdateDailySales(SoldItems, ExampleShopID, dailyRevenue)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while saving data to dailyShopSales")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
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

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	AllMenus := []models.MenuItem{{Category: "All"}, {Category: "UnCategorized"}, {Category: "Out Of Production"}}

	SoldOutItems := []models.Item{{Name: "Example", ListingID: 12, DataShopID: "1122"}}
	SoldOutItems[0].ID = uint(1)
	ShopRequest := &models.ShopRequest{

		ShopName: "exampleShop",
		Status:   "Pending",
	}

	TestShop.On("ExecuteCreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Out Of Production", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT ("id") DO UPDATE SET "menu_item_id"="excluded"."menu_item_id" RETURNING "id"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	exists, err := implShop.CheckAndUpdateOutOfProdMenu(AllMenus, SoldOutItems, ShopRequest)

	assert.True(t, exists)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCheckAndUpdateOutOfProdMenu_NoExist(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

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

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

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

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_menus"`)).
		WillReturnRows(sqlmock.NewRows([]string{"shop_id"}).AddRow(1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Category"}).AddRow(1, "Out Of Production"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := implShop.CreateOutOfProdMenu(ShopExample, SoldOutItems, ShopRequest)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateNewOutOfProdMenu_Fail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

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

	TestShop.On("CreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops"`)).
		WillReturnError(errors.New("error while creating menu"))

	err := implShop.CreateOutOfProdMenu(ShopExample, SoldOutItems, ShopRequest)

	assert.Error(t, err)

	assert.Contains(t, err.Error(), "error while creating menu")
	assert.Nil(t, sqlMock.ExpectationsWereMet())
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

func TestUpdateAccountShopRelation(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}
	userID := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID.String()))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1`)).
		WithArgs(userID.String()).WillReturnRows(sqlmock.NewRows([]string{"shop_id", "account_id"}).AddRow(1, userID.String()))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE "shops"."id" = $1 AND "shops"."deleted_at" IS NULL`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "account_shop_following" WHERE "account_shop_following"."account_id" = $1 AND "account_shop_following"."shop_id" IN (NULL)`)).
		WithArgs(userID.String()).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := implShop.UpdateAccountShopRelation(ShopExample, userID)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestUpdateAccountShopRelation_Fail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}
	userID := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WillReturnError(errors.New("error while handling database operation"))

	err := implShop.UpdateAccountShopRelation(ShopExample, userID)

	assert.Error(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestEstablishAccountShopRelation(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}
	userID := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(userID.String(), 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID.String()))

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WillReturnResult(sqlmock.NewResult(1, 2))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT DO NOTHING RETURNING "id"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	sqlMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "account_shop_following" ("account_id","shop_id") VALUES ($1,$2) ON CONFLICT DO NOTHING`)).WithArgs(userID.String(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := implShop.EstablishAccountShopRelation(ShopExample, userID)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestRoundTwoDecimalDigits(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{
			name:     "several digits after decimal point",
			value:    19.32333333,
			expected: 19.32,
		},
		{
			name:     "two digits after decimal point",
			value:    19.32,
			expected: 19.32,
		},
		{
			name:     "one digits after decimal point",
			value:    19.3,
			expected: 19.3,
		},
		{
			name:     "no digits after decimal point",
			value:    19,
			expected: 19,
		},
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := controllers.RoundToTwoDecimalDigits(tc.value)
			if actual != tc.expected {
				t.Errorf("Expected RoundTwoDecimalDigits(%v) to be %v, but got %v", tc.value, tc.expected, actual)
			}
		})
	}
}

func TestGetItemsBySoldItems_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	exampleSoldItemIds := []uint{2000, 2001, 2002, 2003, 2004}
	SoldItems := make([]models.SoldItems, 5)

	for index := range exampleSoldItemIds {
		SoldItems[index].ID = exampleSoldItemIds[index]
		SoldItems[index].ItemID = uint(index + 1)
	}

	for i, itemID := range exampleSoldItemIds {
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).
			WithArgs(itemID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
	}

	items, err := implShop.GetItemsBySoldItems(SoldItems)

	assert.NoError(t, err)
	assert.Equal(t, len(exampleSoldItemIds), len(items))

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetItemsBySoldItems_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	exampleSoldItemIds := []uint{2000, 2001, 2002, 2003, 2004}
	SoldItems := make([]models.SoldItems, 5)

	for index := range exampleSoldItemIds {
		SoldItems[index].ID = exampleSoldItemIds[index]
		SoldItems[index].ItemID = uint(index + 1)
	}
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).
		WithArgs(exampleSoldItemIds[0]).WillReturnError(errors.New("error while processing database operations"))

	_, err := implShop.GetItemsBySoldItems(SoldItems)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while processing database operations")

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHandleHandleGetShopByID__NoShop(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	implShop := controllers.Shop{DB: MockedDataBase}
	router.GET("/testroute/:shopID", implShop.HandleGetShopByID)

	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnError(errors.New("failed to get shop"))
	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetShopByID found no shop in db")

}

func TestHandleHandleGetShopByID__fail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	implShop := controllers.Shop{DB: MockedDataBase}
	router.GET("/testroute/:shopID", implShop.HandleGetShopByID)

	req, err := http.NewRequest("GET", "/testroute", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code, "no id passed")

}

func TestHandleHandleGetShopByID__Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	router.GET("/testroute/:shopID", implShop.HandleGetShopByID)

	TestShop.On("ExecuteGetAverageItemPrice").Return(1.0, nil)
	TestShop.On("ExecuteGetTotalRevenue").Return(1.0, nil)
	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnRows(sqlmock.NewRows([]string{"id", "shop_id"}).AddRow(1, 1))
	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnRows(sqlmock.NewRows([]string{"id", "shop_id"}).AddRow(1, 1))
	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 1, "Test1"))
	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnRows(sqlmock.NewRows([]string{"id", "shop_id"}).AddRow(9, 1))
	sqlMock.ExpectQuery(regexp.QuoteMeta(``)).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID"}).AddRow(1, 9))

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestHandleGetItemsByShopID__NoShop(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	router.GET("/testroute/:shopID", implShop.HandleGetItemsByShopID)

	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("no shop found"))

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "ExecuteGetItemsByShopID found no shop in db")

}

func TestHandleGetItemsByShopID_Fail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	router.GET("/testroute/:shopID", implShop.HandleGetItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code, "no id passed")

}

func TestHandleGetItemsByShopID__Success(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{}, nil)
	router.GET("/testroute/:shopID", implShop.HandleGetItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute/1", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestHandleGetSoldItemsByShopID__NoShop(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	router.GET("/testroute/:shopID/all_sold_items", implShop.HandleGetSoldItemsByShopID)

	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("no shop found"))

	req, err := http.NewRequest("GET", "/testroute//all_sold_items", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetItemsByShopID found no shop in db")

}

func TestHandleGetSoldItemsByShopID_Fail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	TestShop.On("ExecuteGetSoldItemsByShopID").Return(nil, errors.New("error getting data"))

	router.GET("/testroute/:shopID/all_sold_items", implShop.HandleGetSoldItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute/1/all_sold_items", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "no id passed")

}

func TestHandleGetSoldItemsByShopID_Success(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}

	TestShop.On("ExecuteGetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{}, nil)

	router.GET("/testroute/:shopID/all_sold_items", implShop.HandleGetSoldItemsByShopID)

	req, err := http.NewRequest("GET", "/testroute/1/all_sold_items", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestHandleGetItemsCountByShopID__NoShop(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	router.GET("/testroute/:shopID/items_count", implShop.HandleGetItemsCountByShopID)

	req, err := http.NewRequest("GET", "/testroute/d/items_count", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "GetItemsByShopID found no shop in db")

}

func TestHandleGetItemsCountByShopID__Fail(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	router.GET("/testroute/:shopID/items_count", implShop.HandleGetItemsCountByShopID)

	TestShop.On("ExecuteGetItemsByShopID").Return(nil, errors.New("no shop found"))

	req, err := http.NewRequest("GET", "/testroute/1/items_count", nil)
	if err != nil {
		t.Fatalf("Failed to create test request: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

}

func TestHandleGetItemsCountByShopID_Success(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	_, router, w := SetGinTestMode()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	router.GET("/testroute/:shopID/items_count", implShop.HandleGetItemsCountByShopID)

	TestShop.On("ExecuteGetItemsByShopID").Return([]models.Item{}, nil)

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

func TestGetSoldItemsInRange_success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	ShopId := uint(2)
	fromDate := utils.TruncateDate(time.Now())

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopId, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	_, err := implShop.GetSoldItemsInRange(fromDate, ShopId)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetSoldItemsInRange_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	ShopId := uint(2)
	fromDate := utils.TruncateDate(time.Now())

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopId, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("internal error"))

	_, err := implShop.GetSoldItemsInRange(fromDate, ShopId)
	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateSoldStats_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	dailyShopSales := []models.DailyShopSales{
		{
			ShopID:       1,
			TotalSales:   100,
			DailyRevenue: 90.1,
		},
	}
	for i := range dailyShopSales {
		dailyShopSales[i].CreatedAt = time.Now().AddDate(0, 0, (-len(dailyShopSales) + i))
	}

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
		WithArgs(dailyShopSales[0].ShopID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("internal error"))

	_, err := implShop.CreateSoldStats(dailyShopSales)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateSoldStats_Success_with_Items(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

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
	for i, row := range dailyShopSales {
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
			WithArgs(row.ShopID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).
			WithArgs(i + 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
	}
	stats, err := implShop.CreateSoldStats(dailyShopSales)

	for _, record := range stats {
		assert.Equal(t, 1, len(record.Items))
	}

	assert.Equal(t, len(dailyShopSales), len(stats))
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateSoldStats_Success_with_No_Items(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

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
	for _, row := range dailyShopSales {
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
			WithArgs(row.ShopID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	}
	stats, err := implShop.CreateSoldStats(dailyShopSales)

	for _, record := range stats {
		assert.Equal(t, 0, len(record.Items))
	}

	assert.Equal(t, len(dailyShopSales), len(stats))
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetAverageItemPrice_Shop_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	rows := sqlmock.NewRows([]string{"average_price"}).AddRow(10.5)
	sqlMock.ExpectQuery("SELECT AVG\\(items.original_price\\) as average_price").
		WithArgs(2).WillReturnRows(rows)

	Average, err := implShop.GetAverageItemPrice(ShopExample.ID)

	assert.Equal(t, 10.5, Average)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestGetAverageItemPrice_Shop_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery("SELECT AVG\\(items.original_price\\) as average_price").
		WithArgs(2).WillReturnError(errors.New("Error generateing average price"))

	_, err := implShop.GetAverageItemPrice(ShopExample.ID)

	assert.Contains(t, err.Error(), "Error generateing average price")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestCreateShopRequest_TypeShop_FailNoAccount(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	implShop := controllers.Shop{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		ShopName: "exampleShop",
		Status:   "Pending",
	}

	err := implShop.CreateShopRequest(ShopRequest)

	assert.Contains(t, err.Error(), "no AccountID was passed")

}
func TestCreateShopRequest_TypeShop_FailSaveData(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	implShop := controllers.Shop{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_requests" ("created_at","updated_at","deleted_at","account_id","shop_name","status") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).WillReturnError(errors.New("Failed to save ShopRequest"))
	sqlMock.ExpectRollback()

	err := implShop.CreateShopRequest(ShopRequest)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "Failed to save ShopRequest")
}
func TestCreateShopRequest_TypeShop_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	implShop := controllers.Shop{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_requests" ("created_at","updated_at","deleted_at","account_id","shop_name","status") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()
	implShop.CreateShopRequest(ShopRequest)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetItemsByShopID_TypeShop_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	implShop := controllers.Shop{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	implShop.GetItemsByShopID(ShopExample.ID)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetItemsByShopID_TypeShop_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}
	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnError(errors.New("error while getting shop from DB"))

	_, err := implShop.GetItemsByShopID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error while getting shop from DB")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetShopByName_TypeShop_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE name = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.Name, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "name"}).AddRow(10, ShopExample.ID, "Owner"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "ShopRating"}).AddRow(5, ShopExample.ID, 4.4))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews_topics" WHERE "reviews_topics"."reviews_id"`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 5, "Test1").AddRow(7, 5, "Test2"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	implShop.GetShopByName("ExampleShop")

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetShopByName_TypeShop_fail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	implShop := controllers.Shop{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE name = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.Name, 1).WillReturnError(errors.New("Error getting shop data"))

	_, err := implShop.GetShopByName("ExampleShop")

	assert.Contains(t, err.Error(), "Error getting shop data")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

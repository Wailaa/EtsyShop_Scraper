package controllers_test

import (
	"EtsyScraper/controllers"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	setupMockServer "EtsyScraper/setupTests"
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
)

type MockedShop struct {
	mock.Mock
}

func (m *MockedShop) ExecuteCreateShop(dispatch controllers.ExecShopMethodProcess, ShopRequest *models.ShopRequest) {

}

func (m *MockedShop) GetAvarageItemPrice(ShopID uint) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}
func (m *MockedShop) ExecuteGetTotalRevenue(dispatch controllers.ExecShopMethodProcess, ShopID uint, AvarageItemPrice float64) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}
func (m *MockedShop) ExecuteGetSoldItemsByShopID(dispatch controllers.ExecShopMethodProcess, ID uint) (SoldItemInfos []controllers.ResponseSoldItemInfo, err error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []controllers.ResponseSoldItemInfo
	if shopInterface != nil {
		soldItems = shopInterface.([]controllers.ResponseSoldItemInfo)
	}
	return soldItems, args.Error(1)
}
func (m *MockedShop) ExecuteUpdateSellingHistory(controllers.ShopController, *models.Shop, *models.TaskSchedule, *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) ExecuteUpdateDiscontinuedItems(dispatch controllers.ShopController, Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []models.SoldItems
	if shopInterface != nil {
		soldItems = shopInterface.([]models.SoldItems)
	}
	return soldItems, args.Error(1)
}

func (m *MockedShop) ExecuteGetSellingStatsByPeriod(dispatch controllers.ExecShopMethodProcess, ShopID uint, timePeriod time.Time) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Stats map[string]controllers.DailySoldStats
	if shopInterface != nil {
		Stats = shopInterface.(map[string]controllers.DailySoldStats)
	}
	return Stats, args.Error(1)
}

func (m *MockedShop) GetItemsByShopID(ID uint) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var items []models.Item
	if shopInterface != nil {
		items = shopInterface.([]models.Item)
	}
	return items, args.Error(1)
}
func (m *MockedShop) CreateShopRequest(shopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockedShop) GetShopByName(name string) (*models.Shop, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
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

	TestShop.On("GetShopByName").Return(nil, errors.New("Error"))
	TestShop.On("CreateShopRequest").Return(errors.New("SecondError"))

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "CreateShopRequest")
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

	TestShop.On("GetShopByName").Return(&models.Shop{Name: "ShopName"}, nil)
	TestShop.On("CreateShopRequest").Return(errors.New("SecondError"))

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "CreateShopRequest")
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

	TestShop.On("GetShopByName").Return(nil, errors.New("record not found"))
	TestShop.On("CreateShopRequest").Return(nil)

	router.POST("/create_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.CreateNewShopRequest)

	body := []byte(`{"new_shop_name":"ShopExample"}`)
	req, _ := http.NewRequest("POST", "/create_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "CreateShopRequest")
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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
	TestShop.On("CreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("Failed to save shop"))
	sqlMock.ExpectRollback()

	err := Shop.CreateNewShop(ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save shop")
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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

	TestShop.On("CreateShopRequest").Return(nil)

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("failed to save new record"))
	sqlMock.ExpectRollback()

	err := Shop.CreateNewShop(ShopRequest)

	assert.Error(t, err)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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

	TestShop.On("CreateShopRequest").Return(nil)

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()
	sqlMock.ExpectBegin()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()

	err := Shop.CreateNewShop(ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 2)
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

	TestShop.On("CreateShopRequest").Return(nil)

	Scraper.On("ScrapShop").Return(ShopExample, nil)
	Scraper.On("ScrapAllMenuItems").Return(ShopExample)
	TestShop.On("ExecuteUpdateSellingHistory").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 10, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()
	sqlMock.ExpectBegin()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 10, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{userID.String()}))
	sqlMock.ExpectCommit()

	err := Shop.CreateNewShop(ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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
	TestShop.On("CreateShopRequest").Return(nil)

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("error getting Items"))

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting Items")
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

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
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)

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
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	TestShop.On("CreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}, {}}, nil)
	TestShop.On("CreateShopRequest").Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "daily_shop_sales" SET "updated_at"=$1,"sold_items"=$2 WHERE created_at > $3 AND shop_id = $4 AND "daily_shop_sales"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 14))
	sqlMock.ExpectCommit()

	err := Shop.UpdateSellingHistory(ShopExample, Task, ShopRequest)

	assert.NoError(t, err)
	TestShop.AssertNumberOfCalls(t, "ExecuteUpdateDiscontinuedItems", 1)
	TestShop.AssertNumberOfCalls(t, "CreateShopRequest", 1)
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
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("Error While fetching Shop's details"))

	_, err := Shop.UpdateDiscontinuedItems(ShopExample, Task, ShopRequest)

	assert.Error(t, err)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)
	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
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
	TestShop.On("GetItemsByShopID").Return(ShopItems, nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_menus" ("created_at","updated_at","deleted_at","shop_id","total_items_ammount") VALUES ($1,$2,$3,$4,$5) ON CONFLICT ("id") DO UPDATE SET "shop_id"="excluded"."shop_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9),($10,$11,$12,$13,$14,$15,$16,$17,DEFAULT) ON CONFLICT ("id") DO UPDATE SET "shop_menu_id"="excluded"."shop_menu_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT ("id") DO UPDATE SET "menu_item_id"="excluded"."menu_item_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1"}))

	sqlMock.ExpectCommit()

	_, err := Shop.UpdateDiscontinuedItems(&ExampleShop, Task, ShopRequest)

	assert.NoError(t, err)
	Scraper.AssertNumberOfCalls(t, "ScrapSalesHistory", 1)
	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

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

	TestShop.On("GetShopByName").Return(nil, errors.New("record not found"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "GetShopByName")
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

	TestShop.On("GetShopByName").Return(nil, errors.New("Error getting Shop"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetShopByName", 1)
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
	TestShop.On("GetShopByName").Return(&ShopExample, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnError(errors.New("Error while getting account"))

	router.POST("/follow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.FollowShop)

	body := []byte(`{"follow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/follow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetShopByName", 1)
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
	TestShop.On("GetShopByName").Return(&ShopExample, nil)

	Account := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password_hashed", "subscription_type", "email_verified", "email_verification_token", "request_change_pass", "account_pass_reset_token", "last_time_logged_in", "last_time_logged_out"}).
		AddRow(currentUserUUID.String(), time, time, time, "Testing", "User", "", "", "free", false, "", false, "", time, time)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE ID = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnRows(Account)

	sqlMock.ExpectBegin()

	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), currentUserUUID.String()).WillReturnResult(sqlmock.NewResult(1, 1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","social_media_links","has_sold_history","on_vacation","created_by_user_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT DO NOTHING RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "00000000-0000-0000-0000-000000000000", ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"1"}))

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

	TestShop.AssertNumberOfCalls(t, "GetShopByName", 1)

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

	TestShop.On("GetShopByName").Return(nil, errors.New("record not found"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertCalled(t, "GetShopByName")
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

	TestShop.On("GetShopByName").Return(nil, errors.New("Error getting Shop"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetShopByName", 1)
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
	TestShop.On("GetShopByName").Return(&ShopExample, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(currentUserUUID.String(), 1).WillReturnError(errors.New("Error while getting account"))

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetShopByName", 1)
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
	TestShop.On("GetShopByName").Return(&ShopExample, nil)

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

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET "created_at"=$1,"updated_at"=$2,"deleted_at"=$3,"first_name"=$4,"last_name"=$5,"email"=$6,"password_hashed"=$7,"subscription_type"=$8,"email_verified"=$9,"email_verification_token"=$10,"request_change_pass"=$11,"account_pass_reset_token"=$12,"last_time_logged_in"=$13,"last_time_logged_out"=$14 WHERE "accounts"."deleted_at" IS NULL AND "id" = $15`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), currentUserUUID.String()).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	router.POST("/unfollow_shop", func(ctx *gin.Context) {
		ctx.Set("currentUserUUID", currentUserUUID)
	}, Shop.UnFollowShop)

	body := []byte(`{"unfollow_shop":"ExampleShop"}`)
	req, _ := http.NewRequest("POST", "/unfollow_shop", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "GetShopByName", 1)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetShopByName_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

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
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	TestShop.GetShopByName("ExampleShop")

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetShopByName_Success_fail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE name = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.Name, 1).WillReturnError(errors.New("Error getting shop data"))

	_, err := TestShop.GetShopByName("ExampleShop")

	assert.Contains(t, err.Error(), "Error getting shop data")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetShopByID_AvaragePriceFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &MockedShop{}
	implShop := controllers.Shop{DB: MockedDataBase, Process: TestShop}
	Shop := controllers.NewShopController(implShop)

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)
	TestShop.On("GetAvarageItemPrice").Return(float64(0), errors.New("error getting Item avarage price"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "name"}).AddRow(10, ShopExample.ID, "Owner"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "ShopRating"}).AddRow(5, ShopExample.ID, 4.4))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews_topics" WHERE "reviews_topics"."reviews_id"`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 5, "Test1").AddRow(7, 5, "Test2"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	_, err := Shop.GetShopByID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error getting Item avarage price")
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
	TestShop.On("GetAvarageItemPrice").Return(float64(15.5), nil)
	TestShop.On("ExecuteGetTotalRevenue").Return(float64(0), errors.New("Error while getting Total revenue"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_members" WHERE "shop_members"."shop_id" = $1 AND "shop_members"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "name"}).AddRow(10, ShopExample.ID, "Owner"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE "reviews"."shop_id" = $1 AND "reviews"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "ShopRating"}).AddRow(5, ShopExample.ID, 4.4))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews_topics" WHERE "reviews_topics"."reviews_id"`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "ReviewsID", "Keyword"}).AddRow(5, 5, "Test1").AddRow(7, 5, "Test2"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

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
	TestShop.On("GetAvarageItemPrice").Return(float64(15.5), nil)
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
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	Shop.GetShopByID(ShopExample.ID)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestGetItemsByShopID_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(ShopExample.ID, ShopExample.Name))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" = $1 AND "shop_menus"."deleted_at" IS NULL`)).
		WithArgs(ShopExample.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "ShopID", "TotalItemsAmmount"}).AddRow(9, ShopExample.ID, 5).AddRow(11, ShopExample.ID, 10))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "ShopMenuID", "SectionID"}).AddRow(8, 9, "SelectionID"))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE "items"."menu_item_id" = $1 AND "items"."deleted_at" IS NULL`)).
		WithArgs(8).WillReturnRows(sqlmock.NewRows([]string{"id", "Name", "Available", "MenuItemID"}).AddRow(8, "ItemName", true, 8))

	TestShop.GetItemsByShopID(ShopExample.ID)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetItemsByShopID_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnError(errors.New("error while getting shop from DB"))

	_, err := TestShop.GetItemsByShopID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error while getting shop from DB")
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
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("error while calculating item avarage price "))

	Shop.GetItemsCountByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
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
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}}, nil)

	Shop.GetItemsCountByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
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
	TestShop.On("GetItemsByShopID").Return(nil, errors.New("error while calculating item avarage price "))

	Shop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
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

	TestShop.On("GetItemsByShopID").Return(Allitems, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sold_items" WHERE listing_id IN ($1,$2,$3) AND "sold_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(Solditems)

	Shop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)
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
	TestShop.On("GetItemsByShopID").Return([]models.Item{{}, {}}, nil)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sold_items" WHERE listing_id IN ($1,$2) AND "sold_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnError(errors.New("items were not found"))

	_, err := Shop.GetSoldItemsByShopID(ShopExample.ID)

	TestShop.AssertNumberOfCalls(t, "GetItemsByShopID", 1)

	assert.Contains(t, err.Error(), "items were not found")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetAvarageItemPrice_Success(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	rows := sqlmock.NewRows([]string{"average_price"}).AddRow(10.5)
	sqlMock.ExpectQuery("SELECT AVG\\(items.original_price\\) as average_price").
		WithArgs(2).WillReturnRows(rows)

	Avarage, err := TestShop.GetAvarageItemPrice(ShopExample.ID)

	assert.Equal(t, 10.5, Avarage)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestGetAvarageItemPrice_Fail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery("SELECT AVG\\(items.original_price\\) as average_price").
		WithArgs(2).WillReturnError(errors.New("Error generateing average price"))

	_, err := TestShop.GetAvarageItemPrice(ShopExample.ID)

	assert.Contains(t, err.Error(), "Error generateing average price")
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
	AvarageItemPrice := 19.2
	TestShop.On("ExecuteGetSoldItemsByShopID").Return(nil, errors.New("Sold items where not found"))

	_, err := Shop.GetTotalRevenue(ShopExample.ID, AvarageItemPrice)

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
	AvarageItemPrice := 19.2
	revenueExpected := 485.68

	TestShop.On("ExecuteGetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{{Available: true, OriginalPrice: 15.2, SoldQauntity: 3}, {Available: true, OriginalPrice: 19.12, SoldQauntity: 10}, {Available: true, OriginalPrice: 124.44, SoldQauntity: 2}}, nil)

	Revenue, err := Shop.GetTotalRevenue(ShopExample.ID, AvarageItemPrice)

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
	AvarageItemPrice := 19.2
	revenueExpected := 485.68

	TestShop.On("ExecuteGetSoldItemsByShopID").Return([]controllers.ResponseSoldItemInfo{{Available: true, OriginalPrice: 15.2, SoldQauntity: 3}, {Available: true, OriginalPrice: 19.12, SoldQauntity: 10}, {Available: true, OriginalPrice: 124.44, SoldQauntity: 2}}, nil)

	Revenue, err := Shop.GetTotalRevenue(ShopExample.ID, AvarageItemPrice)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSoldItemsByShopID", 1)
	assert.NoError(t, err)
	assert.Equal(t, revenueExpected, Revenue)
}

func TestCreateShopRequest_FailNoAccount(t *testing.T) {

	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		ShopName: "exampleShop",
		Status:   "Pending",
	}

	err := TestShop.CreateShopRequest(ShopRequest)

	assert.Contains(t, err.Error(), "no AccountID was passed")

}
func TestCreateShopRequest_FailSaveData(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_requests" ("created_at","updated_at","deleted_at","account_id","shop_name","status") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).WillReturnError(errors.New("Failed to save ShopRequest"))
	sqlMock.ExpectRollback()

	err := TestShop.CreateShopRequest(ShopRequest)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "Failed to save ShopRequest")
}
func TestCreateShopRequest_Success(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()
	TestShop := &controllers.ShopCreators{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_requests" ("created_at","updated_at","deleted_at","account_id","shop_name","status") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()
	TestShop.CreateShopRequest(ShopRequest)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

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
	var err error

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	router.GET(route, func(ctx *gin.Context) {
		err = Shop.ProcessStatsRequest(ctx, ShopID, Period)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Contains(t, err.Error(), "invalid period provided")
	assert.Equal(t, http.StatusOK, w.Code)

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
	var err error

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	TestShop.On("ExecuteGetSellingStatsByPeriod").Return(nil, errors.New("error while fetcheing data from db"))

	router.GET(route, func(ctx *gin.Context) {
		err = Shop.ProcessStatsRequest(ctx, ShopID, Period)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSellingStatsByPeriod", 1)
	assert.Contains(t, err.Error(), "error while fetcheing data from db")

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
	var err error

	route := fmt.Sprintf("/stats/%v/%s", ShopID, Period)

	stats := map[string]controllers.DailySoldStats{
		"2024-01-01": {
			TotalSales: 100,
			Items:      []models.Item{{}, {}},
		},
	}

	TestShop.On("ExecuteGetSellingStatsByPeriod").Return(stats, nil)

	router.GET(route, func(ctx *gin.Context) {
		err = Shop.ProcessStatsRequest(ctx, ShopID, Period)
	})

	req, _ := http.NewRequest("GET", route, nil)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	TestShop.AssertNumberOfCalls(t, "ExecuteGetSellingStatsByPeriod", 1)
	assert.NoError(t, err)
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

	DailySales := sqlmock.NewRows([]string{"id", "created_at", "ShopID", "TotalSales", "SoldItems"}).AddRow(1, now.AddDate(0, 0, -3), ShopID, 90, []byte{}).AddRow(2, now.AddDate(0, 0, -4), ShopID, 95, []byte(`[29669, 29670, 29671 ,29672 ,29673]`))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "daily_shop_sales" WHERE (shop_id = $1 AND created_at > $2) AND "daily_shop_sales"."deleted_at" IS NULL`)).WithArgs(ShopID, Period).WillReturnRows(DailySales)
	for i := 1; i <= 5; i++ {
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).WillReturnRows(sqlmock.NewRows([]string{"ID", "ListingID"}).AddRow(i, 1234+i))
	}
	itemsCount := 0
	stats, _ := Shop.GetSellingStatsByPeriod(ShopID, Period)
	for _, value := range stats {
		itemsCount += len(value.Items)
	}

	assert.Equal(t, 2, len(stats))
	assert.Equal(t, 5, itemsCount)
	assert.IsTypef(t, map[string]controllers.DailySoldStats{}, stats, "GetSellingStatsByPeriod return map[string]controllers.DailySoldStats type")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
package scheduleUpdates_test

import (
	"EtsyScraper/models"
	scheduleUpdates "EtsyScraper/scheduleUpdateTask"
	setupMockServer "EtsyScraper/setupTests"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/gorm"
)

type MockCronJob struct {
	AddFuncCalled bool
	AddFuncArg1   string
	StartCalled   bool
}

func (m *MockCronJob) AddFunc(spec string, cmd func()) {
	m.AddFuncCalled = true
	m.AddFuncArg1 = spec
}

func (m *MockCronJob) Start() {
	m.StartCalled = true
}

type MockShopController struct {
	mock.Mock
}

func (m *MockShopController) UpdateSellingHistory(shop *models.Shop, task *models.TaskSchedule, shopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockShopController) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	args := m.Called()
	return args.Get(0).([]models.SoldItems), args.Error(1)
}

func TestScheduleScrapUpdate_SchedulesCronJob(t *testing.T) {

	cronJob := &MockCronJob{}

	err := scheduleUpdates.ScheduleScrapUpdate(cronJob)

	assert.Nil(t, err)

	assert.True(t, cronJob.AddFuncCalled)
	assert.True(t, cronJob.StartCalled)
	assert.Equal(t, "12 15 * * *", cronJob.AddFuncArg1)
}

func TestUpdateSoldItems_ShopParameterNil(t *testing.T) {

	shopController := &MockShopController{}

	queue := scheduleUpdates.UpdateSoldItemsQueue{
		Shop: models.Shop{},
		Task: models.TaskSchedule{},
	}
	shopController.On("UpdateSellingHistory").Return(nil)

	scheduleUpdates.UpdateSoldItems(queue, shopController)

	shopController.AssertNumberOfCalls(t, "UpdateSellingHistory", 1)

}

func TestGetAllShops_Success(t *testing.T) {
	mock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	defer testDB.Close()

	shopRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Shop 1").
		AddRow(2, "Shop 2")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"shops\"")).WillReturnRows(shopRows)

	shopMenuRows := sqlmock.NewRows([]string{"id", "shop_id", "total_items_amount"}).
		AddRow(1, 1, 2).
		AddRow(2, 2, 2)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" IN ($1,$2) AND "shop_menus"."deleted_at" IS NULL`)).
		WillReturnRows(shopMenuRows)

	menuRows := sqlmock.NewRows([]string{"id", "shop_menu_id", "category"}).
		AddRow(1, 1, "Category 1").
		AddRow(2, 1, "Category 2").
		AddRow(3, 2, "Category 1").
		AddRow(4, 2, "Category 2")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WillReturnRows(menuRows)

	updateDB := scheduleUpdates.NewUpdateDB(MockedDataBase)

	_, err := updateDB.GetAllShops()
	if err != nil {
		t.Errorf("An Error accured While testing getAllShops()")
	}

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestGetAllShops_Error(t *testing.T) {
	mock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	defer testDB.Close()

	expextedQuery := `SELECT * FROM "shops" WHERE "shops"."deleted_at"`

	mock.MatchExpectationsInOrder(true)
	mock.ExpectQuery(regexp.QuoteMeta(expextedQuery)).WillReturnError(gorm.ErrRecordNotFound)

	updateDB := scheduleUpdates.NewUpdateDB(MockedDataBase)

	_, err := updateDB.GetAllShops()

	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
func TestMenuExists(t *testing.T) {
	Menu := "Test"
	ListOfMenus := []string{"this", "is", "a", "Test"}
	result := scheduleUpdates.MenuExists(Menu, ListOfMenus)
	assert.True(t, result)
}
func TestMenuExists_NotFound(t *testing.T) {
	Menu := "test"
	ListOfMenus := []string{"this", "is", "a", "Test"}
	result := scheduleUpdates.MenuExists(Menu, ListOfMenus)
	assert.False(t, result)
}

type MockUpdate struct {
	mock.Mock
}

func (s *MockUpdate) GetAllShops() (*[]models.Shop, error) {
	args := s.Called()
	return args.Get(0).(*[]models.Shop), args.Error(1)
}
func (s *MockUpdate) ShopItemsUpdate(Shop, updatedShop *models.Shop) error {

	return nil
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
	return args.Get(0).(*models.Shop)
}
func (m *MockScrapper) ScrapShop(shopName string) (*models.Shop, error) {
	args := m.Called()
	return args.Get(0).(*models.Shop), args.Error(1)
}
func (m *MockScrapper) ScrapSalesHistory(ShopName string, Task *models.TaskSchedule) ([]models.SoldItems, *models.TaskSchedule) {
	args := m.Called()
	return args.Get(0).([]models.SoldItems), args.Get(1).(*models.TaskSchedule)
}

func TestStartShopUpdate_UpdatesSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	updateDB := &scheduleUpdates.UpdateDB{DB: MockedDataBase}

	MockedScrapper := &MockScrapper{}
	expectedShop := &models.Shop{TotalSales: 101, Admirers: 10, HasSoldHistory: false}
	MockedScrapper.On("CheckForUpdates").Return(expectedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	shopRows := sqlmock.NewRows([]string{"id", "name", "total_sales", "admirers"}).
		AddRow(1, "Shop 1", 100, 2).
		AddRow(2, "Shop 2", 100, 2)
	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"shops\"")).WillReturnRows(shopRows)

	shopMenuRows := sqlmock.NewRows([]string{"id", "shop_id", "total_items_amount"}).
		AddRow(1, 1, 2).
		AddRow(2, 2, 2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" IN ($1,$2) AND "shop_menus"."deleted_at" IS NULL`)).
		WillReturnRows(shopMenuRows)

	menuRows := sqlmock.NewRows([]string{"id", "shop_menu_id", "category"}).
		AddRow(1, 1, "Category 1").
		AddRow(2, 1, "Category 2").
		AddRow(3, 2, "Category 1").
		AddRow(4, 2, "Category 2")

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WillReturnRows(menuRows)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "daily_shop_sales" ("created_at","updated_at","deleted_at","shop_id","total_sales","admirers","daily_revenue","sold_items") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 101, 10, float64(0), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "shops"`)).
		WithArgs(10, 101, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_menus" ("created_at","updated_at","deleted_at","shop_id","total_items_ammount","id") VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT ("id") DO UPDATE SET "shop_id"="excluded"."shop_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))

	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9),($10,$11,$12,$13,$14,$15,$16,$17,$18) ON CONFLICT ("id") DO UPDATE SET "shop_menu_id"="excluded"."shop_menu_id" RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "2", "3", "4"}))

	sqlMock.ExpectCommit()

	err := updateDB.StartShopUpdate(false, MockedScrapper)
	if err != nil {
		t.Errorf("error '%s' was not expected", err)
	}
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStartShopUpdate_OneUpdate(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	updateDB := &scheduleUpdates.UpdateDB{DB: MockedDataBase}

	MockedScrapper := &MockScrapper{}
	expectedShop := &models.Shop{TotalSales: 100, Admirers: 2, HasSoldHistory: false}
	MockedScrapper.On("CheckForUpdates").Return(expectedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	shopRows := sqlmock.NewRows([]string{"id", "name", "total_sales", "admirers"}).
		AddRow(1, "Shop 1", 100, 2).
		AddRow(2, "Shop 2", 100, 2)
	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"shops\"")).WillReturnRows(shopRows)

	shopMenuRows := sqlmock.NewRows([]string{"id", "shop_id", "total_items_amount"}).
		AddRow(1, 1, 2).
		AddRow(2, 2, 2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_menus" WHERE "shop_menus"."shop_id" IN ($1,$2) AND "shop_menus"."deleted_at" IS NULL`)).
		WillReturnRows(shopMenuRows)

	menuRows := sqlmock.NewRows([]string{"id", "shop_menu_id", "category"}).
		AddRow(1, 1, "Category 1").
		AddRow(2, 1, "Category 2").
		AddRow(3, 2, "Category 1").
		AddRow(4, 2, "Category 2")

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "menu_items" WHERE "menu_items"."shop_menu_id" IN ($1,$2) AND "menu_items"."deleted_at" IS NULL`)).
		WillReturnRows(menuRows)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "daily_shop_sales" ("created_at","updated_at","deleted_at","shop_id","total_sales","admirers","daily_revenue","sold_items") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 100, 2, float64(0), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))
	sqlMock.ExpectCommit()

	err := updateDB.StartShopUpdate(false, MockedScrapper)
	if err != nil {
		t.Errorf("error '%s' was not expected", err)
	}
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdate_NoUpdates(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	updateDB := &scheduleUpdates.UpdateDB{DB: MockedDataBase}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{}

	UpdatedShop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Amount:    45,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{{ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{{ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}},
				},
			},
		},
	}

	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for i := 1; i < 10; i++ {
		itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
			AddRow(uint(i), time.Now(), time.Now(), time.Now(), fmt.Sprintf("Shop %v", i), 10, "€", -1, "", true, fmt.Sprintf("itemLink%v", i), uint(i+1), uint(i), "101")

		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemRows)
	}

	itemsRowDataShop := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
		AddRow(uint(1), time.Now(), time.Now(), time.Now(), "Shop1", 10, "€", -1, "", true, "itemLink1", 2, uint(1), "101").
		AddRow(uint(2), time.Now(), time.Now(), time.Now(), "Shop2", 10, "€", -1, "", true, "itemLink2", 2, uint(2), "101").
		AddRow(uint(3), time.Now(), time.Now(), time.Now(), "Shop3", 10, "€", -1, "", true, "itemLink3", 2, uint(3), "101").
		AddRow(uint(4), time.Now(), time.Now(), time.Now(), "Shop4", 10, "€", -1, "", true, "itemLink4", 2, uint(4), "101").
		AddRow(uint(5), time.Now(), time.Now(), time.Now(), "Shop5", 10, "€", -1, "", true, "itemLink5", 2, uint(5), "101").
		AddRow(uint(6), time.Now(), time.Now(), time.Now(), "Shop6", 10, "€", -1, "", true, "itemLink6", 2, uint(6), "101").
		AddRow(uint(7), time.Now(), time.Now(), time.Now(), "Shop7", 10, "€", -1, "", true, "itemLink7", 2, uint(7), "101").
		AddRow(uint(8), time.Now(), time.Now(), time.Now(), "Shop8", 10, "€", -1, "", true, "itemLink8", 2, uint(8), "101").
		AddRow(uint(9), time.Now(), time.Now(), time.Now(), "Shop9", 10, "€", -1, "", true, "itemLink9", 2, uint(9), "101")
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemsRowDataShop)

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdate_FewUpdates(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	updateDB := &scheduleUpdates.UpdateDB{DB: MockedDataBase}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{}

	UpdatedShop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Amount:    45,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 20}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{{ListingID: 4, DataShopID: "101", OriginalPrice: 20}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{{ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 20}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}},
				},
			},
		},
	}

	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for i := 1; i < 10; i++ {
		itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
			AddRow(uint(i), time.Now(), time.Now(), time.Now(), fmt.Sprintf("Shop %v", i), 10, "€", -1, "", true, fmt.Sprintf("itemLink%v", i), uint(i+1), uint(i), "101")

		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemRows)

		if i == 1 || i == 4 || i == 7 {
			sqlMock.ExpectBegin()
			sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, i, false, float64(10), float64(20), true, false, i+1, 0).WillReturnRows(sqlmock.NewRows([]string{"1"}))
			sqlMock.ExpectCommit()

			sqlMock.ExpectBegin()
			sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "updated_at"=$1,"original_price"=$2 WHERE "items"."deleted_at" IS NULL AND "id" = $3`)).
				WithArgs(sqlmock.AnyArg(), float64(20), i).WillReturnResult(sqlmock.NewResult(1, 1))
			sqlMock.ExpectCommit()
		}
	}

	itemsRowDataShop := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
		AddRow(uint(1), time.Now(), time.Now(), time.Now(), "Shop1", 10, "€", -1, "", true, "itemLink1", 2, uint(1), "101").
		AddRow(uint(2), time.Now(), time.Now(), time.Now(), "Shop2", 10, "€", -1, "", true, "itemLink2", 2, uint(2), "101").
		AddRow(uint(3), time.Now(), time.Now(), time.Now(), "Shop3", 10, "€", -1, "", true, "itemLink3", 2, uint(3), "101").
		AddRow(uint(4), time.Now(), time.Now(), time.Now(), "Shop4", 10, "€", -1, "", true, "itemLink4", 2, uint(4), "101").
		AddRow(uint(5), time.Now(), time.Now(), time.Now(), "Shop5", 10, "€", -1, "", true, "itemLink5", 2, uint(5), "101").
		AddRow(uint(6), time.Now(), time.Now(), time.Now(), "Shop6", 10, "€", -1, "", true, "itemLink6", 2, uint(6), "101").
		AddRow(uint(7), time.Now(), time.Now(), time.Now(), "Shop7", 10, "€", -1, "", true, "itemLink7", 2, uint(7), "101").
		AddRow(uint(8), time.Now(), time.Now(), time.Now(), "Shop8", 10, "€", -1, "", true, "itemLink8", 2, uint(8), "101").
		AddRow(uint(9), time.Now(), time.Now(), time.Now(), "Shop9", 10, "€", -1, "", true, "itemLink9", 2, uint(9), "101")
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemsRowDataShop)

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdate_NewItemAdded(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	updateDB := &scheduleUpdates.UpdateDB{DB: MockedDataBase}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{}

	UpdatedShop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Amount:    45,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{{ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{{ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}, {ListingID: 10, MenuItemID: 11, DataShopID: "101", OriginalPrice: 100}},
				},
			},
		},
	}

	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for i := 1; i <= 10; i++ {
		if i != 10 {
			itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
				AddRow(uint(i), time.Now(), time.Now(), time.Now(), fmt.Sprintf("Shop %v", i), 10, "€", -1, "", true, fmt.Sprintf("itemLink%v", i), uint(i+1), uint(i), "101")
			sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemRows)
		} else {
			itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"})
			sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemRows)

		}

	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "", float64(100), "", float64(0), "", false, "", 0, 10, "101").WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 0, true, float64(0), float64(100), false, true, 0, 0).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	itemsRowDataShop := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
		AddRow(uint(1), time.Now(), time.Now(), time.Now(), "Shop1", 10, "€", -1, "", true, "itemLink1", 2, uint(1), "101").
		AddRow(uint(2), time.Now(), time.Now(), time.Now(), "Shop2", 10, "€", -1, "", true, "itemLink2", 2, uint(2), "101").
		AddRow(uint(3), time.Now(), time.Now(), time.Now(), "Shop3", 10, "€", -1, "", true, "itemLink3", 2, uint(3), "101").
		AddRow(uint(4), time.Now(), time.Now(), time.Now(), "Shop4", 10, "€", -1, "", true, "itemLink4", 2, uint(4), "101").
		AddRow(uint(5), time.Now(), time.Now(), time.Now(), "Shop5", 10, "€", -1, "", true, "itemLink5", 2, uint(5), "101").
		AddRow(uint(6), time.Now(), time.Now(), time.Now(), "Shop6", 10, "€", -1, "", true, "itemLink6", 2, uint(6), "101").
		AddRow(uint(7), time.Now(), time.Now(), time.Now(), "Shop7", 10, "€", -1, "", true, "itemLink7", 2, uint(7), "101").
		AddRow(uint(8), time.Now(), time.Now(), time.Now(), "Shop8", 10, "€", -1, "", true, "itemLink8", 2, uint(8), "101").
		AddRow(uint(9), time.Now(), time.Now(), time.Now(), "Shop9", 10, "€", -1, "", true, "itemLink9", 2, uint(9), "101").
		AddRow(uint(10), time.Now(), time.Now(), time.Now(), "Shop10", 101, "€", -1, "", true, "itemLink10", 11, uint(1), "101")
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemsRowDataShop)

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdate_CreateNewMenu(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	updateDB := &scheduleUpdates.UpdateDB{DB: MockedDataBase}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{}

	UpdatedShop := &models.Shop{
		ShopMenu: models.ShopMenu{
			Menu: []models.MenuItem{
				{
					Category:  "All",
					SectionID: "0",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "On sale",
					SectionID: "1",
					Amount:    0,
					Items:     []models.Item{},
				},
				{
					Category:  "shelving",
					SectionID: "46696458",
					Amount:    45,
					Items:     []models.Item{{ListingID: 1, DataShopID: "101", OriginalPrice: 10}, {ListingID: 2, DataShopID: "101", OriginalPrice: 10}, {ListingID: 3, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{{ListingID: 4, DataShopID: "101", OriginalPrice: 10}, {ListingID: 5, DataShopID: "101", OriginalPrice: 10}},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{{ListingID: 6, DataShopID: "101", OriginalPrice: 10}, {ListingID: 7, DataShopID: "101", OriginalPrice: 10}, {ListingID: 8, DataShopID: "101", OriginalPrice: 10}, {ListingID: 9, DataShopID: "101", OriginalPrice: 10}, {ListingID: 10, DataShopID: "101", OriginalPrice: 10}},
				},
			},
		},
	}

	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for i := 1; i <= 10; i++ {

		itemRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
			AddRow(uint(i), time.Now(), time.Now(), time.Now(), fmt.Sprintf("Shop %v", i), 10, "€", -1, "", true, fmt.Sprintf("itemLink%v", i), uint(i+1), uint(i), "101")
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemRows)

	}

	itemsRowDataShop := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
		AddRow(uint(1), time.Now(), time.Now(), time.Now(), "Shop1", 10, "€", -1, "", true, "itemLink1", 2, uint(1), "101").
		AddRow(uint(2), time.Now(), time.Now(), time.Now(), "Shop2", 10, "€", -1, "", true, "itemLink2", 2, uint(2), "101").
		AddRow(uint(3), time.Now(), time.Now(), time.Now(), "Shop3", 10, "€", -1, "", true, "itemLink3", 2, uint(3), "101").
		AddRow(uint(4), time.Now(), time.Now(), time.Now(), "Shop4", 10, "€", -1, "", true, "itemLink4", 2, uint(4), "101").
		AddRow(uint(5), time.Now(), time.Now(), time.Now(), "Shop5", 10, "€", -1, "", true, "itemLink5", 2, uint(5), "101").
		AddRow(uint(6), time.Now(), time.Now(), time.Now(), "Shop6", 10, "€", -1, "", true, "itemLink6", 2, uint(6), "101").
		AddRow(uint(7), time.Now(), time.Now(), time.Now(), "Shop7", 10, "€", -1, "", true, "itemLink7", 2, uint(7), "101").
		AddRow(uint(8), time.Now(), time.Now(), time.Now(), "Shop8", 10, "€", -1, "", true, "itemLink8", 2, uint(8), "101").
		AddRow(uint(9), time.Now(), time.Now(), time.Now(), "Shop9", 10, "€", -1, "", true, "itemLink9", 2, uint(9), "101").
		AddRow(uint(10), time.Now(), time.Now(), time.Now(), "Shop10", 101, "€", -1, "", true, "itemLink10", 2, uint(10), "101").
		AddRow(uint(11), time.Now(), time.Now(), time.Now(), "Shop10", 101, "€", -1, "", true, "itemLink10", 11, uint(11), "101")
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WillReturnRows(itemsRowDataShop)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 0, "Out Of Production", "0", "", 0).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 11, false, float64(101), float64(101), true, false, 11, 0).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"updated_at"=$3 WHERE "items"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs(false, 0, sqlmock.AnyArg(), 11).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

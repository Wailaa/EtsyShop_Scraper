package scheduleUpdates_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"EtsyScraper/controllers"
	"EtsyScraper/models"
	"EtsyScraper/repository"
	scheduleUpdates "EtsyScraper/scheduleUpdateTask"
	setupMockServer "EtsyScraper/setupTests"
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

type MockShopUpdater struct {
	mock.Mock
}

func (m *MockShopUpdater) GetShopByID(ID uint) (*models.Shop, error) {

	args := m.Called()
	shopInterface := args.Get(0)
	var shop *models.Shop
	if shopInterface != nil {
		shop = shopInterface.(*models.Shop)
	}
	return shop, args.Error(1)
}

func (m *MockShopUpdater) CreateNewShop(ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockShopUpdater) GetItemsByShopID(ID uint) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Items []models.Item
	if shopInterface != nil {
		Items = shopInterface.([]models.Item)
	}
	return Items, args.Error(1)
}
func (m *MockShopUpdater) GetItemsBySoldItems(SoldItems []models.SoldItems) ([]models.Item, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Items []models.Item
	if shopInterface != nil {
		Items = shopInterface.([]models.Item)
	}
	return Items, args.Error(1)
}

func (m *MockShopUpdater) CreateSoldStats(dailyShopSales []models.DailyShopSales) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()

	return args.Get(0).(map[string]controllers.DailySoldStats), args.Error(1)
}

func (m *MockShopUpdater) CreateOutOfProdMenu(Shop *models.Shop, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockShopUpdater) CreateShopRequest(ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockShopUpdater) GetTotalRevenue(ShopID uint, AverageItemPrice float64) (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}
func (m *MockShopUpdater) CheckAndUpdateOutOfProdMenu(AllMenus []models.MenuItem, SoldOutItems []models.Item, ShopRequest *models.ShopRequest) (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}
func (m *MockShopUpdater) EstablishAccountShopRelation(requestedShop *models.Shop, userID uuid.UUID) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockShopUpdater) GetSoldItemsByShopID(ID uint) (SoldItemInfos []controllers.ResponseSoldItemInfo, err error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []controllers.ResponseSoldItemInfo
	if shopInterface != nil {
		soldItems = shopInterface.([]controllers.ResponseSoldItemInfo)
	}
	return soldItems, args.Error(1)
}

func (m *MockShopUpdater) GetSellingStatsByPeriod(ShopID uint, timePeriod time.Time) (map[string]controllers.DailySoldStats, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var Stats map[string]controllers.DailySoldStats
	if shopInterface != nil {
		Stats = shopInterface.(map[string]controllers.DailySoldStats)
	}
	return Stats, args.Error(1)
}

func (m *MockShopUpdater) UpdateSellingHistory(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockShopUpdater) UpdateDiscontinuedItems(Shop *models.Shop, Task *models.TaskSchedule, ShopRequest *models.ShopRequest) ([]models.SoldItems, error) {
	args := m.Called()
	shopInterface := args.Get(0)
	var soldItems []models.SoldItems
	if shopInterface != nil {
		soldItems = shopInterface.([]models.SoldItems)
	}
	return soldItems, args.Error(1)
}

func (m *MockShopUpdater) SaveShopToDB(scrappedShop *models.Shop, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockShopUpdater) UpdateShopMenuToDB(Shop *models.Shop, ShopRequest *models.ShopRequest) error {
	args := m.Called()
	return args.Error(0)
}

func TestScheduleScrapUpdateSchedulesCronJob(t *testing.T) {

	cronJob := &MockCronJob{}

	updateDB := &scheduleUpdates.UpdateDB{}
	err := scheduleUpdates.ScheduleScrapUpdate(cronJob, updateDB)

	assert.Nil(t, err)

	assert.True(t, cronJob.AddFuncCalled)
	assert.True(t, cronJob.StartCalled)
	assert.Equal(t, "12 15 * * *", cronJob.AddFuncArg1)
}

func TestUpdateSoldItemsShopParameterNil(t *testing.T) {
	_, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	shopController := &MockShopUpdater{}
	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo, Shop: shopController}

	queue := scheduleUpdates.UpdateSoldItemsQueue{
		Shop: models.Shop{},
		Task: models.TaskSchedule{},
	}
	shopController.On("UpdateSellingHistory").Return(nil)

	updateDB.UpdateSoldItems(queue)

	shopController.AssertNumberOfCalls(t, "UpdateSellingHistory", 1)

}

func TestMenuExists(t *testing.T) {
	Menu := "Test"
	ListOfMenus := []string{"this", "is", "a", "Test"}
	result := scheduleUpdates.MenuExists(Menu, ListOfMenus)
	assert.True(t, result)
}
func TestMenuExistsNotFound(t *testing.T) {
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

func TestStartShopUpdateUpdatesSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

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

	for i := 1; i < 3; i++ {
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "daily_shop_sales" ("created_at","updated_at","deleted_at","shop_id","total_sales","admirers","daily_revenue") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), i, 101, 10, float64(0)).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))
		sqlMock.ExpectCommit()

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "shops"`)).
			WithArgs(10, 101, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_menus" ("created_at","updated_at","deleted_at","shop_id","total_items_amount","id") VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT ("id") DO UPDATE SET "shop_id"="excluded"."shop_id" RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))

		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9),($10,$11,$12,$13,$14,$15,$16,$17,$18) ON CONFLICT ("id") DO UPDATE SET "shop_menu_id"="excluded"."shop_menu_id" RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"1", "2", "3", "4"}))

		sqlMock.ExpectCommit()
	}
	err := updateDB.StartShopUpdate(false, MockedScrapper)
	if err != nil {
		t.Errorf("error '%s' was not expected", err)
	}
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestStartShopUpdateOneUpdate(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

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
	for i := 1; i < 3; i++ {
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "daily_shop_sales" ("created_at","updated_at","deleted_at","shop_id","total_sales","admirers","daily_revenue") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), i, 100, 2, float64(0)).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))
		sqlMock.ExpectCommit()
	}
	err := updateDB.StartShopUpdate(false, MockedScrapper)
	if err != nil {
		t.Errorf("error '%s' was not expected", err)
	}
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdateNoUpdates(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{
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
		}}

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
	for _, Menu := range UpdatedShop.ShopMenu.Menu {
		for index := range Menu.Items {
			Menu.Items[index].ID = uint(index + 1)
		}
	}
	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for _, Menu := range UpdatedShop.ShopMenu.Menu {
		for index, item := range Menu.Items {
			sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL ORDER BY "items"."id" LIMIT $2`)).WithArgs(item.ListingID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "ListingID", "DataShopID", "OriginalPrice"}).AddRow(index+1, item.ListingID, item.DataShopID, item.OriginalPrice))
		}

	}
	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdateFewUpdates(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{
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
		}}

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
	for _, Menu := range UpdatedShop.ShopMenu.Menu {
		for index := range Menu.Items {
			Menu.Items[index].ID = uint(index + 1)
		}
	}
	for _, Menu := range ExistingShop.ShopMenu.Menu {
		for index := range Menu.Items {
			Menu.Items[index].ID = uint(index + 1)
		}
	}
	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for _, UpdatedMenu := range UpdatedShop.ShopMenu.Menu {

		if len(UpdatedMenu.Items) == 0 {
			continue
		}

		Category := UpdatedMenu.Category
		for _, ExistingMenu := range ExistingShop.ShopMenu.Menu {
			if ExistingMenu.Category != Category {
				continue
			}

			for _, UpdatedItem := range UpdatedMenu.Items {
				for _, ExistingItem := range ExistingMenu.Items {
					if UpdatedItem.ListingID == ExistingItem.ListingID {

						sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL ORDER BY "items"."id" LIMIT $2`)).WithArgs(ExistingItem.ListingID, 1).
							WillReturnRows(sqlmock.NewRows([]string{"id", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
								AddRow(ExistingItem.ID, ExistingItem.Name, ExistingItem.OriginalPrice, ExistingItem.CurrencySymbol, ExistingItem.SalePrice, ExistingItem.DiscoutPercent, ExistingItem.Available, ExistingItem.ItemLink, ExistingItem.MenuItemID, ExistingItem.ListingID, ExistingItem.DataShopID))
						if UpdatedItem.OriginalPrice != ExistingItem.OriginalPrice {
							sqlMock.ExpectBegin()
							sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
								WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, ExistingItem.ID, false, ExistingItem.OriginalPrice, UpdatedItem.OriginalPrice, ExistingItem.Available, UpdatedItem.Available, ExistingItem.MenuItemID, ExistingItem.MenuItemID).WillReturnRows(sqlmock.NewRows([]string{"1"}))
							sqlMock.ExpectCommit()
							sqlMock.ExpectBegin()
							sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"original_price"=$3,"updated_at"=$4 WHERE "items"."deleted_at" IS NULL AND "id" = $5`)).
								WithArgs(UpdatedItem.Available, ExistingItem.MenuItemID, UpdatedItem.OriginalPrice, sqlmock.AnyArg(), ExistingItem.ID).WillReturnResult(sqlmock.NewResult(1, 1))
							sqlMock.ExpectCommit()
						}
					}
				}
			}
		}
	}

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdateUpdatedItemAdded(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{
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
		}}
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

	for _, Menu := range UpdatedShop.ShopMenu.Menu {
		for index := range Menu.Items {
			if index > 8 {
				Menu.Items[index].ID = 0
			}

		}
	}
	for _, Menu := range ExistingShop.ShopMenu.Menu {
		for index := range Menu.Items {
			Menu.Items[index].ID = uint(index + 1)
		}
	}

	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	for _, UpdatedMenu := range UpdatedShop.ShopMenu.Menu {

		if len(UpdatedMenu.Items) == 0 {
			continue
		}

		Category := UpdatedMenu.Category
		for _, ExistingMenu := range ExistingShop.ShopMenu.Menu {
			if ExistingMenu.Category != Category {
				continue
			}

			for _, UpdatedItem := range UpdatedMenu.Items {
				for _, ExistingItem := range ExistingMenu.Items {
					if UpdatedItem.ListingID == ExistingItem.ListingID {
						sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL ORDER BY "items"."id" LIMIT $2`)).WithArgs(ExistingItem.ListingID, 1).
							WillReturnRows(sqlmock.NewRows([]string{"id", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
								AddRow(ExistingItem.ID, ExistingItem.Name, ExistingItem.OriginalPrice, ExistingItem.CurrencySymbol, ExistingItem.SalePrice, ExistingItem.DiscoutPercent, ExistingItem.Available, ExistingItem.ItemLink, ExistingItem.MenuItemID, ExistingItem.ListingID, ExistingItem.DataShopID))

					}
				}
				if UpdatedItem.ListingID == 10 {
					sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL ORDER BY "items"."id" LIMIT $2`)).WithArgs(UpdatedItem.ListingID, 1).
						WillReturnRows(sqlmock.NewRows([]string{"id", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
							AddRow(UpdatedItem.ID, UpdatedItem.Name, UpdatedItem.OriginalPrice, UpdatedItem.CurrencySymbol, UpdatedItem.SalePrice, UpdatedItem.DiscoutPercent, UpdatedItem.Available, UpdatedItem.ItemLink, UpdatedItem.MenuItemID, UpdatedItem.ListingID, UpdatedItem.DataShopID))
				}

			}

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

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShopItemsUpdateCreateNewMenu(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	MockedScrapper := &MockScrapper{}
	ExistingShop := &models.Shop{
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
					Items:     []models.Item{},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{},
				},
			},
		}}
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
					Items:     []models.Item{},
				},
				{
					Category:  "tables",
					SectionID: "46704593",
					Amount:    44,
					Items:     []models.Item{},
				},
				{
					Category:  "coat racks",
					SectionID: "46704591",
					Amount:    46,
					Items:     []models.Item{},
				},
				{
					Category:  "chairs",
					SectionID: "46704599",
					Amount:    46,
					Items:     []models.Item{},
				},
			},
		},
	}

	for _, Menu := range UpdatedShop.ShopMenu.Menu {
		for index := range Menu.Items {
			Menu.Items[index].ID = uint(index + 1)
		}
	}
	for _, Menu := range ExistingShop.ShopMenu.Menu {
		for index := range Menu.Items {
			Menu.Items[index].ID = uint(index + 1)
		}
	}

	MockedScrapper.On("ScrapAllMenuItems").Return(UpdatedShop, nil)

	sqlMock.MatchExpectationsInOrder(true)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 0, "chairs", "46704599", "", 46).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	updateDB.ShopItemsUpdate(ExistingShop, UpdatedShop, MockedScrapper)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestShouldUpdateItem(t *testing.T) {

	tests := []struct {
		name          string
		existingPrice float64
		newPrice      float64
		expected      bool
	}{
		{
			name:          "Price discrepancy is greater than tolerated ",
			existingPrice: 100.0,
			newPrice:      105.0,
			expected:      true,
		},
		{
			name:          "Price discrepancy is greater than tolerated ",
			existingPrice: 100.0,
			newPrice:      103.0,
			expected:      true,
		},
		{
			name:          "Price discrepancy is greater than tolerated ",
			existingPrice: 100.0,
			newPrice:      96.0,
			expected:      true,
		},
		{
			name:          "Price discrepancy is  tolerated ",
			existingPrice: 100.0,
			newPrice:      99.0,
			expected:      false,
		},
		{
			name:          "Price discrepancy is  tolerated ",
			existingPrice: 100.0,
			newPrice:      101,
			expected:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := scheduleUpdates.ShouldUpdateItem(tc.existingPrice, tc.newPrice)
			if actual != tc.expected {
				t.Errorf("Expected ShouldUpdateItem(%f, %f) to be %t, but got %t", tc.existingPrice, tc.newPrice, tc.expected, actual)
			}
		})
	}
}

func TestApplyUpdatedSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	sqlMock.MatchExpectationsInOrder(true)

	NewItem := models.Item{
		OriginalPrice: 40,
		Available:     true,
	}

	ExistingItem := models.Item{
		Name:           "testItem",
		OriginalPrice:  20.0,
		CurrencySymbol: "$",
		SalePrice:      -1,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.examplelink.com",
		MenuItemID:     10,
		ListingID:      101010,
		DataShopID:     "1234",
	}
	ExistingItem.ID = 7

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, ExistingItem.ID, false, ExistingItem.OriginalPrice, NewItem.OriginalPrice, ExistingItem.Available, NewItem.Available, ExistingItem.MenuItemID, ExistingItem.MenuItemID).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"original_price"=$3,"updated_at"=$4 WHERE "items"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs(NewItem.Available, ExistingItem.MenuItemID, NewItem.OriginalPrice, sqlmock.AnyArg(), ExistingItem.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	updateDB.ApplyItemUpdates(ExistingItem, NewItem, ExistingItem.MenuItemID)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestApplyUpdatedFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	sqlMock.MatchExpectationsInOrder(true)

	NewItem := models.Item{
		OriginalPrice: 40,
		Available:     true,
	}
	NewItem.ID = uint(2)

	ExistingItem := models.Item{}
	ExistingItem.ID = 1

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, ExistingItem.ID, false, ExistingItem.OriginalPrice, NewItem.OriginalPrice, ExistingItem.Available, NewItem.Available, ExistingItem.MenuItemID, ExistingItem.MenuItemID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ExistingItem.ID))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"original_price"=$3,"updated_at"=$4 WHERE "items"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs(NewItem.Available, ExistingItem.MenuItemID, NewItem.OriginalPrice, sqlmock.AnyArg(), ExistingItem.ID).WillReturnError(errors.New("no item id"))
	sqlMock.ExpectRollback()

	updateDB.ApplyItemUpdates(ExistingItem, NewItem, ExistingItem.MenuItemID)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestHandleOutOfProductionItemsCreateNewMenu(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	dataShopId := "101"
	OutOfProductionID := uint(0)
	ShopMenuID := uint(2)
	existingItemMap := make(map[uint]bool)

	for i := uint(1); i < 9; i++ {
		existingItemMap[i] = true
	}

	itemsRowDataShop := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
		AddRow(uint(1), time.Now(), time.Now(), time.Now(), "Shop1", 10, "€", -1, "", true, "itemLink1", ShopMenuID, uint(1), dataShopId).
		AddRow(uint(2), time.Now(), time.Now(), time.Now(), "Shop2", 10, "€", -1, "", true, "itemLink2", ShopMenuID, uint(2), dataShopId).
		AddRow(uint(3), time.Now(), time.Now(), time.Now(), "Shop3", 10, "€", -1, "", true, "itemLink3", ShopMenuID, uint(3), dataShopId).
		AddRow(uint(4), time.Now(), time.Now(), time.Now(), "Shop4", 10, "€", -1, "", true, "itemLink4", ShopMenuID, uint(4), dataShopId).
		AddRow(uint(5), time.Now(), time.Now(), time.Now(), "Shop5", 10, "€", -1, "", true, "itemLink5", ShopMenuID, uint(5), dataShopId).
		AddRow(uint(6), time.Now(), time.Now(), time.Now(), "Shop6", 10, "€", -1, "", true, "itemLink6", ShopMenuID, uint(6), dataShopId).
		AddRow(uint(7), time.Now(), time.Now(), time.Now(), "Shop7", 10, "€", -1, "", true, "itemLink7", ShopMenuID, uint(7), dataShopId).
		AddRow(uint(8), time.Now(), time.Now(), time.Now(), "Shop8", 10, "€", -1, "", true, "itemLink8", ShopMenuID, uint(8), dataShopId).
		AddRow(uint(9), time.Now(), time.Now(), time.Now(), "Shop9", 10, "€", -1, "", true, "itemLink9", ShopMenuID, uint(9), dataShopId).
		AddRow(uint(10), time.Now(), time.Now(), time.Now(), "Shop10", 10, "€", -1, "", true, "itemLink10", ShopMenuID, uint(10), dataShopId).
		AddRow(uint(11), time.Now(), time.Now(), time.Now(), "Shop10", 10, "€", -1, "", true, "itemLink10", ShopMenuID, uint(11), dataShopId)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WithArgs(dataShopId).WillReturnRows(itemsRowDataShop)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, ShopMenuID, "Out Of Production", "0", "", 0).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	for i := 9; i <= 11; i++ {

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, i, false, float64(10), float64(10), true, false, ShopMenuID, int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		sqlMock.ExpectCommit()

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"updated_at"=$3 WHERE "items"."deleted_at" IS NULL AND "id" = $4`)).
			WithArgs(false, 1, sqlmock.AnyArg(), i).WillReturnResult(sqlmock.NewResult(1, 2))
		sqlMock.ExpectCommit()
	}

	updateDB.HandleOutOfProductionItems(dataShopId, OutOfProductionID, ShopMenuID, existingItemMap)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestHandleOutOfProductionItemsExists(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	dataShopId := "101"
	OutOfProductionID := uint(10)
	ShopMenuID := uint(2)
	existingItemMap := make(map[uint]bool)

	for i := uint(1); i < 9; i++ {
		existingItemMap[i] = true
	}

	itemsRowDataShop := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
		AddRow(uint(1), time.Now(), time.Now(), time.Now(), "Shop1", 10, "€", -1, "", true, "itemLink1", ShopMenuID, uint(1), dataShopId).
		AddRow(uint(2), time.Now(), time.Now(), time.Now(), "Shop2", 10, "€", -1, "", true, "itemLink2", ShopMenuID, uint(2), dataShopId).
		AddRow(uint(3), time.Now(), time.Now(), time.Now(), "Shop3", 10, "€", -1, "", true, "itemLink3", ShopMenuID, uint(3), dataShopId).
		AddRow(uint(4), time.Now(), time.Now(), time.Now(), "Shop4", 10, "€", -1, "", true, "itemLink4", ShopMenuID, uint(4), dataShopId).
		AddRow(uint(5), time.Now(), time.Now(), time.Now(), "Shop5", 10, "€", -1, "", true, "itemLink5", ShopMenuID, uint(5), dataShopId).
		AddRow(uint(6), time.Now(), time.Now(), time.Now(), "Shop6", 10, "€", -1, "", true, "itemLink6", ShopMenuID, uint(6), dataShopId).
		AddRow(uint(7), time.Now(), time.Now(), time.Now(), "Shop7", 10, "€", -1, "", true, "itemLink7", ShopMenuID, uint(7), dataShopId).
		AddRow(uint(8), time.Now(), time.Now(), time.Now(), "Shop8", 10, "€", -1, "", true, "itemLink8", ShopMenuID, uint(8), dataShopId).
		AddRow(uint(9), time.Now(), time.Now(), time.Now(), "Shop9", 10, "€", -1, "", true, "itemLink9", ShopMenuID, uint(9), dataShopId).
		AddRow(uint(10), time.Now(), time.Now(), time.Now(), "Shop10", 10, "€", -1, "", true, "itemLink10", ShopMenuID, uint(10), dataShopId).
		AddRow(uint(11), time.Now(), time.Now(), time.Now(), "Shop10", 10, "€", -1, "", true, "itemLink10", ShopMenuID, uint(11), dataShopId)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WithArgs(dataShopId).WillReturnRows(itemsRowDataShop)

	for i := 9; i <= 11; i++ {

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, i, false, float64(10), float64(10), true, false, ShopMenuID, OutOfProductionID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		sqlMock.ExpectCommit()

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"updated_at"=$3 WHERE "items"."deleted_at" IS NULL AND "id" = $4`)).
			WithArgs(false, OutOfProductionID, sqlmock.AnyArg(), i).WillReturnResult(sqlmock.NewResult(1, 2))
		sqlMock.ExpectCommit()
	}

	updateDB.HandleOutOfProductionItems(dataShopId, OutOfProductionID, ShopMenuID, existingItemMap)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestAddNewItemSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	sqlMock.MatchExpectationsInOrder(true)
	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	Item := models.Item{
		Name:           "testItem",
		OriginalPrice:  20.0,
		CurrencySymbol: "$",
		SalePrice:      -1,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.examplelink.com",
		MenuItemID:     10,
		ListingID:      101010,
		DataShopID:     "1234",
	}
	Item.ID = 7

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`NSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, Item.Name, Item.OriginalPrice, Item.CurrencySymbol, Item.SalePrice, Item.DiscoutPercent, Item.Available, Item.ItemLink, Item.MenuItemID, Item.ListingID, Item.DataShopID, Item.ID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, Item.ID, sqlmock.AnyArg(), float64(0), Item.OriginalPrice, false, true, 0, Item.MenuItemID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	_ = updateDB.AddNewItem(Item)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestAddNewItemFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	sqlMock.MatchExpectationsInOrder(true)
	ShopRepo := &repository.DataBase{DB: MockedDataBase}
	updateDB := &scheduleUpdates.UpdateDB{Repo: ShopRepo}

	Item := models.Item{
		Name:           "testItem",
		OriginalPrice:  20.0,
		CurrencySymbol: "$",
		SalePrice:      -1,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.examplelink.com",
		MenuItemID:     10,
		ListingID:      101010,
		DataShopID:     "1234",
	}
	Item.ID = 7

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WillReturnError(errors.New("error while handling db operation"))

	sqlMock.ExpectRollback()

	err := updateDB.AddNewItem(Item)

	assert.Contains(t, err.Error(), "error while handling db operation")
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestAddSoldItemsQueueList(t *testing.T) {
	SoldItemsQueueList := []scheduleUpdates.UpdateSoldItemsQueue{}
	NewSoldItems := 5
	Shop := models.Shop{Name: "ExampleShop"}

	SoldItemsQueueList = scheduleUpdates.AddSoldItemsQueueList(SoldItemsQueueList, NewSoldItems, Shop)

	assert.Equal(t, 1, len(SoldItemsQueueList), "a new SoldItemsQueue should be added")
	assert.Equal(t, Shop, SoldItemsQueueList[0].Shop)
	assert.Equal(t, NewSoldItems, SoldItemsQueueList[0].Task.UpdateSoldItems)
}

package repository_test

import (
	"EtsyScraper/models"
	"EtsyScraper/repository"
	setupMockServer "EtsyScraper/setupTests"
	"EtsyScraper/utils"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCreateShopToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	sqlMock.ExpectCommit()

	err := ShopRepo.CreateShop(ShopExample)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestCreateShopToDBFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("Failed to save shop"))
	sqlMock.ExpectRollback()

	err := ShopRepo.CreateShop(ShopExample)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save shop")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestSaveShopToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := ShopRepo.SaveShop(ShopExample)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestSaveShopToDBFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := &models.Shop{
		Name: "exampleShop",
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shops" ("created_at","updated_at","deleted_at","name","description","location","total_sales","joined_since","last_update_time","admirers","has_sold_history","on_vacation","created_by_user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopExample.Name, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("error while saving data"))
	sqlMock.ExpectRollback()

	err := ShopRepo.SaveShop(ShopExample)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestSaveSoldItemsToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 12, "1122", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 2, 13, "1122").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := ShopRepo.SaveSoldItemsToDB(SoldItems)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestSaveSoldItemsToDBFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sold_items" ("created_at","updated_at","deleted_at","item_id","listing_id","data_shop_id") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).WillReturnError(errors.New("error while saving sold item"))
	sqlMock.ExpectRollback()

	err := ShopRepo.SaveSoldItemsToDB(SoldItems)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while saving sold item")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateDailySalesSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ExampleShopID := uint(10)
	dailyRevenue := 98.9
	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "daily_shop_sales" SET "updated_at"=$1,"daily_revenue"=$2 WHERE created_at > $3 AND shop_id = $4 AND "daily_shop_sales"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), dailyRevenue, sqlmock.AnyArg(), ExampleShopID).WillReturnResult(sqlmock.NewResult(1, 3))
	sqlMock.ExpectCommit()

	err := ShopRepo.UpdateDailySales(SoldItems, ExampleShopID, dailyRevenue)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateDailySalesFailed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ExampleShopID := uint(10)
	dailyRevenue := 98.9
	SoldItems := []models.SoldItems{{Name: "Example", ItemID: 1, ListingID: 12, DataShopID: "1122"}, {Name: "Example2", ItemID: 2, ListingID: 13, DataShopID: "1122"}}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "daily_shop_sales" SET "updated_at"=$1,"daily_revenue"=$2 WHERE created_at > $3 AND shop_id = $4 AND "daily_shop_sales"."deleted_at" IS NULL`)).
		WillReturnError(errors.New("error while saving data to dailyShopSales"))
	sqlMock.ExpectRollback()

	err := ShopRepo.UpdateDailySales(SoldItems, ExampleShopID, dailyRevenue)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while saving data to dailyShopSales")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestSaveShopToMenu(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	AllMenus := models.MenuItem{Category: "All"}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := ShopRepo.SaveMenu(AllMenus)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestSaveShopToDBMenu(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	AllMenus := models.MenuItem{Category: "All"}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WillReturnError(errors.New("error while saving data"))
	sqlMock.ExpectRollback()

	err := ShopRepo.SaveMenu(AllMenus)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestFetchShopByID(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

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

	_, err := ShopRepo.FetchShopByID(ShopExample.ID)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestFetchShopByIDFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnError(errors.New("error hhandling db"))

	_, err := ShopRepo.FetchShopByID(ShopExample.ID)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestFetchStatsByPeriod(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopID := uint(2)
	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	DailySales := sqlmock.NewRows([]string{"id", "created_at", "ShopID", "TotalSales"}).AddRow(1, now.AddDate(0, 0, -3), ShopID, 90).AddRow(2, now.AddDate(0, 0, -4), ShopID, 95)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "daily_shop_sales" WHERE (shop_id = $1 AND created_at > $2) AND "daily_shop_sales"."deleted_at" IS NULL`)).WithArgs(ShopID, Period).WillReturnRows(DailySales)

	_, err := ShopRepo.FetchStatsByPeriod(ShopID, Period)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestFetchStatsByPeriodFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopID := uint(2)
	now := time.Now()
	Period := now.AddDate(0, 0, -6)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "daily_shop_sales" WHERE (shop_id = $1 AND created_at > $2) AND "daily_shop_sales"."deleted_at" IS NULL`)).WithArgs(ShopID, Period).WillReturnError(errors.New("error while handling db"))

	_, err := ShopRepo.FetchStatsByPeriod(ShopID, Period)

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestFetchSoldItemsByListingID(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	Allitems := []models.Item{{ListingID: 1}, {ListingID: 2}, {ListingID: 3}}
	for i := range Allitems {
		Allitems[i].ID = uint(i + 1)
	}

	Solditems := sqlmock.NewRows([]string{"id", "listingID", "ItemID"}).AddRow(1, 1, 1).AddRow(2, 1, 1).AddRow(3, 3, 3)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sold_items" WHERE listing_id IN ($1,$2,$3) AND "sold_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnRows(Solditems)

	_, err := ShopRepo.FetchSoldItemsByListingID([]uint{1, 2, 3})

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestFetchSoldItemsByListingIDFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sold_items" WHERE listing_id IN ($1,$2,$3) AND "sold_items"."deleted_at" IS NULL`)).
		WithArgs().WillReturnError(errors.New("error While handling DB"))

	_, err := ShopRepo.FetchSoldItemsByListingID([]uint{1, 2, 3})

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestFetchItemsBySoldItems(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	SolditemID := uint(2)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).
		WithArgs(SolditemID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	_, err := ShopRepo.FetchItemsBySoldItems(uint(2))

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestFetchItemsBySoldItemsFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	SolditemID := uint(2)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = ($1)`)).
		WithArgs(SolditemID).WillReturnError(errors.New("error while hanfdling data"))

	_, err := ShopRepo.FetchItemsBySoldItems(uint(2))

	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetSoldItemsInRange(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopId := uint(2)
	fromDate := utils.TruncateDate(time.Now())

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopId, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	_, err := ShopRepo.GetSoldItemsInRange(fromDate, ShopId)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetSoldItemsInRangeFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopId := uint(2)
	fromDate := utils.TruncateDate(time.Now())

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT sold_items.* FROM "shops" JOIN shop_menus ON shops.id = shop_menus.shop_id JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id JOIN items ON menu_items.id = items.menu_item_id JOIN sold_items ON items.id = sold_items.item_id WHERE (shops.id = $1 AND sold_items.created_at BETWEEN $2 AND $3) AND "shops"."deleted_at" IS NULL`)).
		WithArgs(ShopId, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("internal error"))

	_, err := ShopRepo.GetSoldItemsInRange(fromDate, ShopId)
	assert.Error(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestUpdateAccountShopRelation(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

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

	err := ShopRepo.UpdateAccountShopRelation(ShopExample, userID)

	assert.NoError(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestUpdateAccountShopRelationFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := &models.Shop{
		Name:           "exampleShop",
		TotalSales:     10,
		HasSoldHistory: true,
	}
	userID := uuid.New()

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 AND "accounts"."deleted_at" IS NULL ORDER BY "accounts"."id" LIMIT $2`)).
		WillReturnError(errors.New("error while handling database operation"))

	err := ShopRepo.UpdateAccountShopRelation(ShopExample, userID)

	assert.Error(t, err)

	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetAverageItemPrice(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	rows := sqlmock.NewRows([]string{"average_price"}).AddRow(10.5)
	sqlMock.ExpectQuery("SELECT AVG\\(items.original_price\\) as average_price").
		WithArgs(2).WillReturnRows(rows)

	Average, err := ShopRepo.GetAverageItemPrice(ShopExample.ID)

	assert.Equal(t, 10.5, Average)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestGetAverageItemPriceShopFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery("SELECT AVG\\(items.original_price\\) as average_price").
		WithArgs(2).WillReturnError(errors.New("Error generateing average price"))

	_, err := ShopRepo.GetAverageItemPrice(ShopExample.ID)

	assert.Contains(t, err.Error(), "Error generateing average price")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestSaveShopRequestToDB(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_requests" ("created_at","updated_at","deleted_at","account_id","shop_name","status") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).WillReturnError(errors.New("Failed to save ShopRequest"))
	sqlMock.ExpectRollback()

	err := ShopRepo.SaveShopRequestToDB(ShopRequest)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "Failed to save ShopRequest")
}

func TestCreateShopRequestTypeShopSuccess(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopRequest := &models.ShopRequest{
		AccountID: uuid.New(),
		ShopName:  "exampleShop",
		Status:    "Pending",
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shop_requests" ("created_at","updated_at","deleted_at","account_id","shop_name","status") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := ShopRepo.SaveShopRequestToDB(ShopRequest)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestGetShopWithItemsByShopID(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

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

	_, err := ShopRepo.GetShopWithItemsByShopID(ShopExample.ID)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}
func TestGetShopWithItemsByShopIDFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE id = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.ID, 1).WillReturnError(errors.New("error while getting shop from DB"))

	_, err := ShopRepo.GetShopWithItemsByShopID(ShopExample.ID)

	assert.Contains(t, err.Error(), "error while getting shop from DB")
	assert.NoError(t, sqlMock.ExpectationsWereMet())

}

func TestGetShopByNameTypeShopSuccess(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

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

	ShopRepo.GetShopByName("ExampleShop")

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
func TestGetShopByNameTypeShopfail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopExample := models.Shop{Name: "ExampleShop"}
	ShopExample.ID = uint(2)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shops" WHERE name = $1 AND "shops"."deleted_at" IS NULL ORDER BY "shops"."id" LIMIT $2`)).
		WithArgs(ShopExample.Name, 1).WillReturnError(errors.New("Error getting shop data"))

	_, err := ShopRepo.GetShopByName("ExampleShop")

	assert.Contains(t, err.Error(), "Error getting shop data")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
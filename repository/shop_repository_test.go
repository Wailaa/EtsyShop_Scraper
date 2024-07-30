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
	"gorm.io/gorm"
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

	ShopExample := models.Shop{}
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

	ShopExample := models.Shop{}
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

	ShopExample := models.Shop{}
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

	ShopExample := models.Shop{}
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

func TestGetAllShopsSuccess(t *testing.T) {
	mock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

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

	_, err := ShopRepo.GetAllShops()
	if err != nil {
		t.Errorf("An Error occurred While testing getAllShops()")
	}

	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestGetAllShopsError(t *testing.T) {
	mock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	expectedQuery := `SELECT * FROM "shops" WHERE "shops"."deleted_at"`

	mock.MatchExpectationsInOrder(true)
	mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).WillReturnError(gorm.ErrRecordNotFound)

	_, err := ShopRepo.GetAllShops()

	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestCreateDailySales(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopID := uint(10)
	TotalSales := 100
	Admirers := 90

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "daily_shop_sales" ("created_at","updated_at","deleted_at","shop_id","total_sales","admirers","daily_revenue") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ShopID, TotalSales, Admirers, float64(0)).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))
	sqlMock.ExpectCommit()

	ShopRepo.CreateDailySales(ShopID, TotalSales, Admirers)

	assert.Nil(t, sqlMock.ExpectationsWereMet())

}

func TestCreateDailySalesFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ShopID := uint(10)
	TotalSales := 100
	Admirers := 90

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "daily_shop_sales" ("created_at","updated_at","deleted_at","shop_id","total_sales","admirers","daily_revenue") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
		WillReturnError(errors.New("error while handling database operation"))
	sqlMock.ExpectRollback()

	err := ShopRepo.CreateDailySales(ShopID, TotalSales, Admirers)

	assert.Contains(t, err.Error(), "error while handling database operation")

	assert.Nil(t, sqlMock.ExpectationsWereMet())

}
func TestUpdateColumnsInShopSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	Shop := models.Shop{}
	Shop.ID = uint(2)
	updateData := map[string]interface{}{
		"total_sales": 201,
		"admirers":    101,
	}

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "shops" SET "admirers"=$1,"total_sales"=$2,"updated_at"=$3 WHERE "shops"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs(101, 201, sqlmock.AnyArg(), Shop.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	sqlMock.ExpectCommit()

	err := ShopRepo.UpdateColumnsInShop(Shop, updateData)

	assert.NoError(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}
func TestUpdateColumnsInShopFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	Shop := models.Shop{}
	Shop.ID = uint(2)
	updateData := map[string]interface{}{
		"total_sales": 201,
		"admirers":    101,
	}

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "shops" SET "admirers"=$1,"total_sales"=$2,"updated_at"=$3 WHERE "shops"."deleted_at" IS NULL AND "id" = $4`)).
		WillReturnError(errors.New("error while handling database operation"))

	sqlMock.ExpectRollback()

	err := ShopRepo.UpdateColumnsInShop(Shop, updateData)

	assert.Contains(t, err.Error(), "error while handling database operation")
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestCreateMenuFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}
	menu := models.MenuItem{

		Category:  "Out Of Production",
		SectionID: "0",
		Amount:    0,
		Items:     []models.Item{},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WillReturnError(errors.New("error while handling database operation"))
	sqlMock.ExpectRollback()

	_, err := ShopRepo.CreateMenu(menu)
	assert.Contains(t, err.Error(), "error while handling database operation")
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestCreateMenuSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}
	menu := models.MenuItem{

		Category:  "Out Of Production",
		SectionID: "0",
		Link:      "JustALink.com",
		Amount:    0,
		Items:     []models.Item{},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "menu_items" ("created_at","updated_at","deleted_at","shop_menu_id","category","section_id","link","amount") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 0, menu.Category, menu.SectionID, menu.Link, menu.Amount).WillReturnRows(sqlmock.NewRows([]string{"1", "2"}))
	sqlMock.ExpectCommit()

	_, err := ShopRepo.CreateMenu(menu)
	assert.NoError(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}
func TestGetItemByListingIDSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ExistingItem := models.Item{
		Name:           "ExampeItem",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}
	ExistingItem.ID = uint(15)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL ORDER BY "items"."id" LIMIT $2`)).WithArgs(ExistingItem.ListingID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
			AddRow(ExistingItem.ID, ExistingItem.Name, ExistingItem.OriginalPrice, ExistingItem.CurrencySymbol, ExistingItem.SalePrice, ExistingItem.DiscoutPercent, ExistingItem.Available, ExistingItem.ItemLink, ExistingItem.MenuItemID, ExistingItem.ListingID, ExistingItem.DataShopID))

	result, err := ShopRepo.GetItemByListingID(ExistingItem.ListingID)
	assert.Equal(t, result, &ExistingItem)

	assert.NoError(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}
func TestGetAllItemsByDataShopIDSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	DataShopID := "98889"

	ExistingItems := []models.Item{{
		Name:           "ExampeItem",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}, {

		Name:           "ExampeItem2",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}}

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WithArgs(DataShopID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "original_price", "currency_symbol", "sale_price", "discount_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id"}).
			AddRow(ExistingItems[0].ID, ExistingItems[0].Name, ExistingItems[0].OriginalPrice, ExistingItems[0].CurrencySymbol, ExistingItems[0].SalePrice, ExistingItems[0].DiscoutPercent, ExistingItems[0].Available, ExistingItems[0].ItemLink, ExistingItems[0].MenuItemID, ExistingItems[0].ListingID, ExistingItems[0].DataShopID).
			AddRow(ExistingItems[1].ID, ExistingItems[1].Name, ExistingItems[1].OriginalPrice, ExistingItems[1].CurrencySymbol, ExistingItems[1].SalePrice, ExistingItems[1].DiscoutPercent, ExistingItems[1].Available, ExistingItems[1].ItemLink, ExistingItems[1].MenuItemID, ExistingItems[1].ListingID, ExistingItems[1].DataShopID))

	result, err := ShopRepo.GetAllItemsByDataShopID(DataShopID)
	assert.Equal(t, result, ExistingItems)

	assert.NoError(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}
func TestGetItemByDataShopIDFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	DataShopID := "98889"

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE data_shop_id = $1 AND "items"."deleted_at" IS NULL`)).WithArgs(DataShopID).
		WillReturnError(errors.New("error while handling database"))

	_, err := ShopRepo.GetAllItemsByDataShopID(DataShopID)

	assert.Error(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

func TestGetItemByListingIDFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ExistingItem := models.Item{
		ListingID: uint(9),
	}
	ExistingItem.ID = uint(15)

	sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "items" WHERE Listing_id = $1 AND "items"."deleted_at" IS NULL ORDER BY "items"."id" LIMIT $2`)).WithArgs(ExistingItem.ListingID, 1).WillReturnError(errors.New("error while fetching database"))
	_, err := ShopRepo.GetItemByListingID(ExistingItem.ListingID)

	assert.Error(t, err, "error while fetching database")
	assert.Nil(t, sqlMock.ExpectationsWereMet())
}
func TestCreateItemHistoryChange(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	UpdatedMenuID := uint(15)

	NewItem := models.Item{
		OriginalPrice: 40,
		Available:     true,
	}

	ExistingItem := models.Item{
		OriginalPrice: 20.0,
		Available:     true,
		MenuItemID:    UpdatedMenuID,
	}

	Change := models.ItemHistoryChange{
		ItemID:         ExistingItem.ID,
		NewItemCreated: false,
		OldPrice:       ExistingItem.OriginalPrice,
		NewPrice:       NewItem.OriginalPrice,
		OldAvailable:   ExistingItem.Available,
		NewAvailable:   NewItem.Available,
		OldMenuItemID:  ExistingItem.MenuItemID,
		NewMenuItemID:  UpdatedMenuID,
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, ExistingItem.ID, false, ExistingItem.OriginalPrice, NewItem.OriginalPrice, ExistingItem.Available, NewItem.Available, ExistingItem.MenuItemID, UpdatedMenuID).WillReturnRows(sqlmock.NewRows([]string{"1"}))
	sqlMock.ExpectCommit()

	err := ShopRepo.CreateItemHistoryChange(Change)

	assert.NoError(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())

}
func TestCreateItemHistoryChangeFail(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	UpdatedMenuID := uint(15)

	NewItem := models.Item{
		OriginalPrice: 40,
		Available:     true,
	}

	ExistingItem := models.Item{

		OriginalPrice: 20.0,
		Available:     true,
		MenuItemID:    UpdatedMenuID,
	}
	Change := models.ItemHistoryChange{
		ItemID:         ExistingItem.ID,
		NewItemCreated: false,
		OldPrice:       ExistingItem.OriginalPrice,
		NewPrice:       NewItem.OriginalPrice,
		OldAvailable:   ExistingItem.Available,
		NewAvailable:   NewItem.Available,
		OldMenuItemID:  ExistingItem.MenuItemID,
		NewMenuItemID:  UpdatedMenuID,
	}
	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_history_changes" ("created_at","updated_at","deleted_at","item_id","new_item_created","old_price","new_price","old_available","new_available","old_menu_item_id","new_menu_item_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WillReturnError(errors.New("error while handling DB"))
	sqlMock.ExpectRollback()

	err := ShopRepo.CreateItemHistoryChange(Change)

	assert.Error(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())

}

func TestUpdateItemSuccess(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ExistingItem := models.Item{
		Name:           "ExampeItem",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}
	ExistingItem.ID = uint(15)

	NewItem := map[string]interface{}{
		"original_price": 40,
		"available":      true,
		"menu_item_id":   ExistingItem.MenuItemID,
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"original_price"=$3,"updated_at"=$4 WHERE "items"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs(true, ExistingItem.MenuItemID, 40, sqlmock.AnyArg(), ExistingItem.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := ShopRepo.UpdateItem(ExistingItem, NewItem)

	assert.NoError(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())

}
func TestUpdateItemFail(t *testing.T) {

	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	ExistingItem := models.Item{
		Name:           "ExampeItem",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}
	ExistingItem.ID = uint(15)

	NewItem := map[string]interface{}{
		"original_price": 40,
		"available":      true,
		"menu_item_id":   ExistingItem.MenuItemID,
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE "items" SET "available"=$1,"menu_item_id"=$2,"original_price"=$3,"updated_at"=$4 WHERE "items"."deleted_at" IS NULL AND "id" = $5`)).
		WithArgs(true, ExistingItem.MenuItemID, 40, sqlmock.AnyArg(), ExistingItem.ID).WillReturnError(errors.New("error while handling DB"))
	sqlMock.ExpectRollback()

	err := ShopRepo.UpdateItem(ExistingItem, NewItem)

	assert.Error(t, err)
	assert.Nil(t, sqlMock.ExpectationsWereMet())

}

func TestCreateNewItemFailed(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	Item := models.Item{
		Name:           "ExampeItem",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}
	Item.ID = uint(15)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, Item.Name, Item.OriginalPrice, Item.CurrencySymbol, Item.SalePrice, Item.DiscoutPercent, Item.Available, Item.ItemLink, Item.MenuItemID, Item.ListingID, Item.DataShopID, Item.ID).WillReturnError(errors.New("error while handling DB"))
	sqlMock.ExpectRollback()

	_, err := ShopRepo.CreateNewItem(Item)

	assert.Error(t, err)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}
func TestCreateNewItemSuccess(t *testing.T) {
	sqlMock, testDB, MockedDataBase := setupMockServer.StartMockedDataBase()
	testDB.Begin()
	defer testDB.Close()

	ShopRepo := repository.DataBase{DB: MockedDataBase}

	Item := models.Item{
		Name:           "ExampeItem",
		OriginalPrice:  10.0,
		CurrencySymbol: "€",
		SalePrice:      10.0,
		DiscoutPercent: "",
		Available:      true,
		ItemLink:       "www.ExampleLink.com",
		MenuItemID:     uint(2),
		ListingID:      uint(9),
		DataShopID:     "98889",
	}
	Item.ID = uint(15)

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items" ("created_at","updated_at","deleted_at","name","original_price","currency_symbol","sale_price","discout_percent","available","item_link","menu_item_id","listing_id","data_shop_id","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, Item.Name, Item.OriginalPrice, Item.CurrencySymbol, Item.SalePrice, Item.DiscoutPercent, Item.Available, Item.ItemLink, Item.MenuItemID, Item.ListingID, Item.DataShopID, Item.ID).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at", "deleted_at", "name", "original_price", "currency_symbol", "sale_price", "discout_percent", "available", "item_link", "menu_item_id", "listing_id", "data_shop_id", "id"}))
	sqlMock.ExpectCommit()

	_, err := ShopRepo.CreateNewItem(Item)

	assert.NoError(t, err)

	assert.Nil(t, sqlMock.ExpectationsWereMet())
}

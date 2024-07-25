package repository

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"time"

	"github.com/google/uuid"
)

type ShopRepository interface {
	CreateShop(scrappedShop *models.Shop) error
	SaveShop(Shop *models.Shop) error
	SaveSoldItemsToDB(ScrappedSoldItems []models.SoldItems) error
	UpdateDailySales(ScrappedSoldItems []models.SoldItems, ShopID uint, dailyRevenue float64) error
	SaveMenu(Menus models.MenuItem) error
	FetchShopByID(ID uint) (*models.Shop, error)
	FetchStatsByPeriod(ShopID uint, timePeriod time.Time) ([]models.DailyShopSales, error)
	FetchSoldItemsByListingID(listingIDs []uint) ([]models.SoldItems, error)
	FetchItemsBySoldItems(soldItemID uint) (models.Item, error)
	GetSoldItemsInRange(fromDate time.Time, ShopID uint) ([]models.SoldItems, error)
	UpdateAccountShopRelation(requestedShop *models.Shop, UserID uuid.UUID) error
	GetAverageItemPrice(ShopID uint) (float64, error)
	SaveShopRequestToDB(ShopRequest *models.ShopRequest) error
	GetShopWithItemsByShopID(ID uint) (*models.Shop, error)
	GetShopByName(ShopName string) (shop *models.Shop, err error)
	GetAllShops() (*[]models.Shop, error)
	CreateDailySales(ShopID uint, TotalSales, Admirers int) error
	UpdateColumnsInShop(Shop models.Shop, updateData map[string]interface{}) error
	CreateMenu(Menus models.MenuItem) error
	GetItemByListingID(ID uint) (*models.Item, error)
}

func (d *DataBase) GetItemByListingID(ID uint) (*models.Item, error) {
	existingItem := models.Item{}
	if err := d.DB.Where("Listing_id = ? ", ID).First(&existingItem).Error; err != nil {
		return nil, utils.HandleError(err)
	}
	return &existingItem, nil
}
func (d *DataBase) UpdateColumnsInShop(Shop models.Shop, updateData map[string]interface{}) error {
	if err := d.DB.Model(&Shop).Updates(updateData).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (d *DataBase) CreateDailySales(ShopID uint, TotalSales, Admirers int) error {
	dailySales := models.DailyShopSales{
		ShopID:     ShopID,
		TotalSales: TotalSales,
		Admirers:   Admirers,
	}

	if err := d.DB.Create(&dailySales).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}
func (d *DataBase) CreateShop(scrappedShop *models.Shop) error {
	if err := d.DB.Create(scrappedShop).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (d *DataBase) SaveShop(Shop *models.Shop) error {
	if err := d.DB.Save(Shop).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (d *DataBase) SaveSoldItemsToDB(ScrappedSoldItems []models.SoldItems) error {
	if err := d.DB.Create(&ScrappedSoldItems).Error; err != nil {
		return utils.HandleError(err, "Shop's selling history failed while saving to database")
	}
	return nil
}

func (d *DataBase) UpdateDailySales(ScrappedSoldItems []models.SoldItems, ShopID uint, dailyRevenue float64) error {

	now := utils.TruncateDate(time.Now())
	dailyRevenue = utils.RoundToTwoDecimalDigits(dailyRevenue)

	if err := d.DB.Model(&models.DailyShopSales{}).Where("created_at > ?", now).Where("shop_id = ?", ShopID).Updates(&models.DailyShopSales{DailyRevenue: dailyRevenue}).Error; err != nil {
		return utils.HandleError(err)
	}

	return nil
}

func (d *DataBase) SaveMenu(Menus models.MenuItem) error {
	if err := d.DB.Save(&Menus).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}
func (d *DataBase) CreateMenu(Menus models.MenuItem) error {
	if err := d.DB.Create(&Menus).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (d *DataBase) FetchShopByID(ID uint) (*models.Shop, error) {
	shop := models.Shop{}
	if err := d.DB.Preload("Member").Preload("ShopMenu.Menu").Preload("Reviews.ReviewsTopic").Where("id = ?", ID).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ")
	}
	return &shop, nil
}

func (d *DataBase) FetchStatsByPeriod(ShopID uint, timePeriod time.Time) ([]models.DailyShopSales, error) {
	dailyShopSales := []models.DailyShopSales{}

	if err := d.DB.Where("shop_id = ? AND created_at > ?", ShopID, timePeriod).Find(&dailyShopSales).Error; err != nil {
		return nil, utils.HandleError(err)
	}
	return dailyShopSales, nil
}

func (d *DataBase) FetchSoldItemsByListingID(listingIDs []uint) ([]models.SoldItems, error) {
	Solditems := []models.SoldItems{}
	if err := d.DB.Where("listing_id IN ?", listingIDs).Find(&Solditems).Error; err != nil {
		return nil, utils.HandleError(err, "items were not found ")
	}

	return Solditems, nil
}

func (d *DataBase) FetchItemsBySoldItems(soldItemID uint) (models.Item, error) {

	item := models.Item{}
	if err := d.DB.Raw("SELECT items.* FROM items JOIN sold_items ON items.id = sold_items.item_id WHERE sold_items.id = (?)", soldItemID).Scan(&item).Error; err != nil {
		return item, utils.HandleError(err, "error parsing sold items")
	}

	return item, nil
}

func (d *DataBase) GetSoldItemsInRange(fromDate time.Time, ShopID uint) ([]models.SoldItems, error) {
	soldItems := []models.SoldItems{}
	tillDate := fromDate.Add(24 * time.Hour)

	if err := d.DB.Table("shops").
		Select("sold_items.*").
		Joins("JOIN shop_menus ON shops.id = shop_menus.shop_id").
		Joins("JOIN menu_items ON shop_menus.id = menu_items.shop_menu_id").
		Joins("JOIN items ON menu_items.id = items.menu_item_id").
		Joins("JOIN sold_items ON items.id = sold_items.item_id").
		Where("shops.id = ? AND sold_items.created_at BETWEEN ? AND ?", ShopID, fromDate, tillDate).
		Find(&soldItems).Error; err != nil {
		return nil, utils.HandleError(err)
	}

	return soldItems, nil
}

func (d *DataBase) UpdateAccountShopRelation(requestedShop *models.Shop, UserID uuid.UUID) error {
	account := &models.Account{}
	account.ID = UserID

	result, err := d.GetAccountWithShops(account.ID)
	if err != nil {
		return utils.HandleError(err)
	}

	if err := d.DB.Model(result).Association("ShopsFollowing").Delete(requestedShop); err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (d *DataBase) GetAverageItemPrice(ShopID uint) (float64, error) {
	var averagePrice float64

	if err := d.DB.Table("items").
		Joins("JOIN menu_items ON items.menu_item_id = menu_items.id").
		Joins("JOIN shop_menus ON menu_items.shop_menu_id = shop_menus.id").
		Joins("JOIN shops ON shop_menus.shop_id = shops.id").
		Where("shops.id = ? AND items.original_price > 0 ", ShopID).
		Select("AVG(items.original_price) as average_price").
		Row().Scan(&averagePrice); err != nil {

		return 0, utils.HandleError(err)
	}
	averagePrice = utils.RoundToTwoDecimalDigits(averagePrice)

	return averagePrice, nil
}

func (d *DataBase) SaveShopRequestToDB(ShopRequest *models.ShopRequest) error {
	if err := d.DB.Save(ShopRequest).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (d *DataBase) GetShopWithItemsByShopID(ID uint) (*models.Shop, error) {
	shop := &models.Shop{}
	if err := d.DB.Preload("ShopMenu.Menu.Items").Where("id = ?", ID).First(shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found")
	}
	return shop, nil
}

func (d *DataBase) GetShopByName(ShopName string) (shop *models.Shop, err error) {

	if err = d.DB.Preload("Member").Preload("ShopMenu.Menu.Items").Preload("Reviews.ReviewsTopic").Where("name = ?", ShopName).First(&shop).Error; err != nil {
		return nil, utils.HandleError(err, "no Shop was Found ,error")
	}
	return
}
func (d *DataBase) GetAllShops() (*[]models.Shop, error) {
	AllShops := &[]models.Shop{}

	if err := d.DB.Preload("ShopMenu.Menu").Find(AllShops).Error; err != nil {
		return nil, utils.HandleError(err, "error while retrieving shops data")
	}

	return AllShops, nil
}

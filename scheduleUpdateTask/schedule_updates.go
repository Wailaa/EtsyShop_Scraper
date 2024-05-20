package scheduleUpdates

import (
	"log"
	"math"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"EtsyScraper/utils"
)

type UpdateDB struct {
	DB *gorm.DB
}

type UpdateSoldItemsQueue struct {
	Shop models.Shop
	Task models.TaskSchedule
}

func NewUpdateDB(DB *gorm.DB) *UpdateDB {
	return &UpdateDB{DB: DB}
}

type CustomCronJob struct {
	cronJob *cron.Cron
}

func NewCustomCronJob() *CustomCronJob {
	return &CustomCronJob{
		cronJob: cron.New(),
	}
}

func (c *CustomCronJob) AddFunc(spec string, cmd func()) {
	c.cronJob.AddFunc(spec, cmd)
}

func (c *CustomCronJob) Start() {
	c.cronJob.Start()
}

func (c *CustomCronJob) Stop() {
	c.cronJob.Stop()
}

type CronJob interface {
	AddFunc(spec string, cmd func())
	Start()
}

func StartScheduleScrapUpdate() {
	c := NewCustomCronJob()
	ScheduleScrapUpdate(c)
}
func ScheduleScrapUpdate(c CronJob) error {
	scraper := &scrap.Scraper{}
	var FuncError error
	c.AddFunc("12 15 * * *", func() {
		log.Println("ScheduleScrapUpdate executed at", time.Now())
		needUpdateItems := false
		if time.Now().Weekday() == time.Tuesday {
			needUpdateItems = true
		}
		if err := NewUpdateDB(initializer.DB).StartShopUpdate(needUpdateItems, scraper); err != nil {
			FuncError = err
		}
	})
	if FuncError != nil {
		return utils.HandleError(FuncError)
	}
	c.Start()
	return nil
}

func (u *UpdateDB) StartShopUpdate(needUpdateItems bool, scraper scrap.ScrapeUpdateProcess) error {

	SoldItemsQueueList := []UpdateSoldItemsQueue{}

	Shops, err := u.GetAllShops()
	if err != nil {
		return utils.HandleError(err, "error while retrieving Shops rows.")
	}

	for _, Shop := range *Shops {

		updatedShop, err := scraper.CheckForUpdates(Shop.Name, needUpdateItems)
		if err != nil {
			return utils.HandleError(err, "error while scraping Shop. error")
		}

		NewSoldItems := updatedShop.TotalSales - Shop.TotalSales
		NewAdmirers := updatedShop.Admirers - Shop.Admirers

		if NewSoldItems > 0 && Shop.HasSoldHistory {
			SoldItemsQueueList = AddSoldItemsQueueList(SoldItemsQueueList, NewSoldItems, Shop)
		}

		if updatedShop.OnVacation {
			updatedShop.TotalSales = Shop.TotalSales
			updatedShop.Admirers = Shop.Admirers
		}

		updateData := map[string]interface{}{
			"total_sales": updatedShop.TotalSales,
			"admirers":    updatedShop.Admirers,
		}

		if err := u.CreateDailySales(Shop.ID, updatedShop.TotalSales, updatedShop.Admirers); err != nil {
			return utils.HandleError(err)
		}

		if NewAdmirers > 0 || NewSoldItems > 0 {
			log.Printf("Shop's name: %s , TotalSales was: %v , TotalSales now: %v \n", Shop.Name, Shop.TotalSales, updatedShop.TotalSales)
			u.DB.Model(&Shop).Updates(updateData)
		}

		if needUpdateItems {
			log.Println("ShopItemsUpdate executed at", time.Now())
			u.ShopItemsUpdate(&Shop, updatedShop, scraper)
		}

	}
	if len(SoldItemsQueueList) > 0 {
		for _, queue := range SoldItemsQueueList {

			newController := controllers.NewShopController(controllers.Shop{DB: u.DB, Scraper: &scrap.Scraper{}})
			UpdateSoldItems(queue, newController)
			log.Printf("added %v new SoldItems to Shop: %s\n", queue.Task.UpdateSoldItems, queue.Shop.Name)
		}
	}
	log.Println("finished updating Shops")

	return nil
}

func UpdateSoldItems(queue UpdateSoldItemsQueue, newController controllers.ShopUpdater) {
	ShopRequest := &models.ShopRequest{}
	newController.UpdateSellingHistory(&queue.Shop, &queue.Task, ShopRequest)
}

func (u *UpdateDB) GetAllShops() (*[]models.Shop, error) {
	AllShops := &[]models.Shop{}

	if err := u.DB.Preload("ShopMenu.Menu").Find(AllShops).Error; err != nil {
		return nil, utils.HandleError(err, "error while retrieving shops data")
	}

	return AllShops, nil
}

func MenuExists(Menu string, ListOfMenus []string) bool {
	for _, newMenu := range ListOfMenus {
		if Menu == newMenu {
			return true
		}
	}
	return false

}

func (u *UpdateDB) ShopItemsUpdate(Shop, updatedShop *models.Shop, scraper scrap.ScrapeUpdateProcess) error {

	dataShopID := ""
	existingItemMap := make(map[uint]bool)
	ListOfMenus := []string{}
	var OutOfProductionID uint

	updatedShop = scraper.ScrapAllMenuItems(updatedShop)
	for _, UpdatedMenu := range updatedShop.ShopMenu.Menu {
		for _, Menu := range Shop.ShopMenu.Menu {

			ListOfMenus = append(ListOfMenus, Menu.Category)

			if Menu.Category == "Out Of Production" {
				OutOfProductionID = Menu.ID
				continue
			}
			if Menu.Category == UpdatedMenu.Category {
				UpdatedMenu.ID = Menu.ID

			}

		}

		if Exists := MenuExists(UpdatedMenu.Category, ListOfMenus); !Exists {
			NewMenu := models.CreateMenuItem(UpdatedMenu)
			NewMenu.ShopMenuID = Shop.ShopMenu.ID
			u.DB.Create(&NewMenu)
			UpdatedMenu.ID = NewMenu.ID
		}

		for _, item := range UpdatedMenu.Items {
			existingItem := models.Item{}
			existingItemMap[item.ListingID] = true

			u.DB.Where("Listing_id = ? ", item.ListingID).First(&existingItem)
			dataShopID = existingItem.DataShopID

			if existingItem.ID == 0 {
				item.MenuItemID = UpdatedMenu.ID
				u.AddNewItem(item)

			} else if ShouldUpdateItem(existingItem.OriginalPrice, item.OriginalPrice) {
				ApplyUpdated(u.DB, existingItem, item, UpdatedMenu.ID)
			}
		}

	}

	u.HandleOutOfProductionItems(dataShopID, OutOfProductionID, Shop.ShopMenu.ID, existingItemMap)
	return nil
}

func ShouldUpdateItem(existingPrice, newPrice float64) bool {
	PriceDiscrepancy := 3.0
	PriceChange := math.Abs((existingPrice / newPrice) - 1)
	PriceChangePerc := math.Round(PriceChange * 100)
	return PriceChangePerc >= PriceDiscrepancy
}

func ApplyUpdated(DB *gorm.DB, existingItem, item models.Item, UpdatedMenuID uint) {

	DB.Create(&models.ItemHistoryChange{
		ItemID:         existingItem.ID,
		NewItemCreated: false,
		OldPrice:       existingItem.OriginalPrice,
		NewPrice:       item.OriginalPrice,
		OldAvailable:   existingItem.Available,
		NewAvailable:   item.Available,
		OldMenuItemID:  existingItem.MenuItemID,
		NewMenuItemID:  UpdatedMenuID,
	})

	DB.Model(&existingItem).Updates(models.Item{
		OriginalPrice: item.OriginalPrice,
		Available:     item.Available,
		MenuItemID:    UpdatedMenuID,
	})

}

func (u *UpdateDB) HandleOutOfProductionItems(dataShopID string, OutOfProductionID, ShopMenuID uint, existingItemMap map[uint]bool) {
	existingItems := []models.Item{}
	u.DB.Where("data_shop_id = ?", dataShopID).Find(&existingItems)

	for _, item := range existingItems {
		if _, ok := existingItemMap[item.ListingID]; !ok && item.MenuItemID != OutOfProductionID {
			if OutOfProductionID == 0 {
				Menu := models.CreateMenuItem(models.MenuItem{
					ShopMenuID: ShopMenuID,
					Category:   "Out Of Production",
					SectionID:  "0",
				})

				u.DB.Create(&Menu)
				OutOfProductionID = Menu.ID
				log.Println("Out Of Production is created , id : ", OutOfProductionID)

			}

			u.DB.Create(&models.ItemHistoryChange{
				ItemID:       item.ID,
				OldPrice:     item.OriginalPrice,
				NewPrice:     item.OriginalPrice,
				OldAvailable: item.Available,
				NewAvailable: false,

				OldMenuItemID: item.MenuItemID,
				NewMenuItemID: OutOfProductionID,
			})

			u.DB.Model(&item).Updates(map[string]interface{}{
				"available":    false,
				"menu_item_id": OutOfProductionID,
			})

			log.Println("item  not available anymore: ", item)
		}
	}
}

func (u *UpdateDB) AddNewItem(item models.Item) error {

	if err := u.DB.Create(&item).Error; err != nil {
		return utils.HandleError(err)
	}

	log.Println("new item created : ", item)

	changeRecords := &models.ItemHistoryChange{
		ItemID:         item.ID,
		NewItemCreated: true,
		OldPrice:       0,
		NewPrice:       item.OriginalPrice,
		OldAvailable:   false,
		NewAvailable:   true,

		NewMenuItemID: item.MenuItemID,
	}

	if err := u.DB.Create(changeRecords).Error; err != nil {
		return utils.HandleError(err)
	}

	return nil
}

func AddSoldItemsQueueList(SoldItemsQueueList []UpdateSoldItemsQueue, NewSoldItems int, Shop models.Shop) []UpdateSoldItemsQueue {
	SoldItemsQueue := UpdateSoldItemsQueue{}
	Task := models.TaskSchedule{
		IsScrapeFinished:     false,
		IsPaginationScrapped: false,
		CurrentPage:          0,
		LastPage:             0,
		UpdateSoldItems:      NewSoldItems,
	}
	SoldItemsQueue.Shop = Shop
	SoldItemsQueue.Task = Task
	SoldItemsQueueList = append(SoldItemsQueueList, SoldItemsQueue)

	return SoldItemsQueueList
}

func (u *UpdateDB) CreateDailySales(ShopID uint, TotalSales, Admirers int) error {
	dailySales := models.DailyShopSales{
		ShopID:     ShopID,
		TotalSales: TotalSales,
		Admirers:   Admirers,
	}

	if err := u.DB.Create(&dailySales).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

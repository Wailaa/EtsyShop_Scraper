package scheduleUpdates

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"log"
	"math"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
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
		return FuncError
	}
	c.Start()
	return nil
}

func (u *UpdateDB) StartShopUpdate(needUpdateItems bool, scraper scrap.ScrapeUpdateProcess) error {
	SoldItemsQueue := UpdateSoldItemsQueue{}
	AddSoldItemsQueue := []UpdateSoldItemsQueue{}

	Shops, err := u.GetAllShops()
	if err != nil {
		log.Println("error while retrieving Shops rows. error :", err)
		return err
	}

	for _, Shop := range *Shops {

		updatedShop, err := scraper.CheckForUpdates(Shop.Name, needUpdateItems)
		if err != nil {
			log.Println("error while scraping Shop. error :", err)
			return err
		}

		NewSoldItems := updatedShop.TotalSales - Shop.TotalSales
		NewAdmirers := updatedShop.Admirers - Shop.Admirers

		if NewSoldItems > 0 && Shop.HasSoldHistory {

			Task := models.TaskSchedule{
				IsScrapeFinished:     false,
				IsPaginationScrapped: false,
				CurrentPage:          0,
				LastPage:             0,
				UpdateSoldItems:      NewSoldItems,
			}
			SoldItemsQueue.Shop = Shop
			SoldItemsQueue.Task = Task
			AddSoldItemsQueue = append(AddSoldItemsQueue, SoldItemsQueue)

		}

		if updatedShop.OnVacation {
			updatedShop.TotalSales = Shop.TotalSales
			updatedShop.Admirers = Shop.Admirers
		}

		updateData := map[string]interface{}{
			"total_sales": updatedShop.TotalSales,
			"admirers":    updatedShop.Admirers,
		}

		dailySales := models.DailyShopSales{
			ShopID:     Shop.ID,
			TotalSales: updatedShop.TotalSales,
			Admirers:   updatedShop.Admirers,
		}

		u.DB.Create(&dailySales)

		if NewAdmirers > 0 || NewSoldItems > 0 {
			log.Printf("Shop's name: %s , TotalSales was: %v , TotalSales now: %v \n", Shop.Name, Shop.TotalSales, updatedShop.TotalSales)
			u.DB.Model(&Shop).Updates(updateData)
		}

		if needUpdateItems {
			log.Println("ShopItemsUpdate executed at", time.Now())
			u.ShopItemsUpdate(&Shop, updatedShop, scraper)
		}

	}
	if len(AddSoldItemsQueue) > 0 {
		for _, queue := range AddSoldItemsQueue {

			newController := controllers.NewShopController(controllers.Shop{DB: u.DB, Process: &controllers.ShopCreators{DB: u.DB}, Scraper: &scrap.Scraper{}})
			UpdateSoldItems(queue, newController)
			log.Printf("added %v new SoldItems to Shop: %s\n", queue.Task.UpdateSoldItems, queue.Shop.Name)
		}
	}
	log.Println("finished updating Shops")

	return nil
}

func UpdateSoldItems(queue UpdateSoldItemsQueue, newController controllers.ShopController) {
	ShopRequest := &models.ShopRequest{}
	newController.UpdateSellingHistory(&queue.Shop, &queue.Task, ShopRequest)
}

func (u *UpdateDB) GetAllShops() (*[]models.Shop, error) {
	AllShops := &[]models.Shop{}

	if err := u.DB.Preload("ShopMenu.Menu").Find(AllShops).Error; err != nil {
		log.Println("error while retrieving shops data , error :", err)
		return nil, err
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
				u.DB.Create(&item)

				log.Println("new item created : ", item)

				u.DB.Create(&models.ItemHistoryChange{
					ItemID:         item.ID,
					NewItemCreated: true,
					OldPrice:       0,
					NewPrice:       item.OriginalPrice,
					OldAvailable:   false,
					NewAvailable:   true,

					NewMenuItemID: item.MenuItemID,
				})

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

package scheduleUpdates

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"fmt"
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
	return &UpdateDB{DB}
}

func ScheduleScrapUpdate() error {
	c := cron.New()

	_, err := c.AddFunc("54 15 * * * ", func() {
		log.Println("ScheduleScrapUpdate executed at", time.Now())
		NewUpdateDB(initializer.DB).StartShopUpdate()
	})
	if err != nil {
		fmt.Println("Error scheduling task:", err)
		return err
	}

	c.Start()
	return nil
}

func (u *UpdateDB) StartShopUpdate() error {
	SoldItemsQueue := UpdateSoldItemsQueue{}
	AddSoldItemsQueue := []UpdateSoldItemsQueue{}
	Shops, err := u.getAllShops()
	if err != nil {
		log.Println("error while retreiving Shops rows. error :", err)
	}

	for _, Shop := range *Shops {

		updatedShop, err := scrap.CheckForUpdates(Shop.Name)
		if err != nil {
			log.Println("error while scraping Shop. error :", err)
			return err
		}
		u.ShopItemsUpdate(&Shop, updatedShop)
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

		if time.Now().Weekday() == time.Wednesday {

			log.Println("ShopItemsUpdate executed at", time.Now())
			u.ShopItemsUpdate(&Shop, updatedShop)
		}

	}
	if len(AddSoldItemsQueue) > 0 {
		for _, queue := range AddSoldItemsQueue {
			UpdateSoldItems(queue)
			log.Printf("added %v new SoldItems to Shop: %s\n", queue.Task.UpdateSoldItems, queue.Shop.Name)
		}
	}

	log.Println("finished updating Shops")

	return nil
}

func UpdateSoldItems(queue UpdateSoldItemsQueue) {
	ShopRequest := &models.ShopRequest{}
	controllers.NewShopController(initializer.DB).UpdateSellingHistory(&queue.Shop, &queue.Task, ShopRequest)
}

func (u *UpdateDB) getAllShops() (*[]models.Shop, error) {
	AllShops := &[]models.Shop{}

	result := u.DB.Preload("ShopMenu.Menu").Find(AllShops)
	if result.Error != nil {
		log.Println("error while retreiving shops data , error :", result.Error)
		return nil, result.Error
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

func (u *UpdateDB) ShopItemsUpdate(Shop, updatedShop *models.Shop) error {

	dataShopID := ""
	existingItems := []models.Item{}
	existingItemMap := make(map[uint]bool)
	ListOfMenus := []string{}
	var OutOfProductionID uint

	updatedShop = scrap.ScrapAllMenuItems(updatedShop)

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

			PriceDiscrepancy := 3.0
			PriceChange := math.Abs((existingItem.OriginalPrice / item.OriginalPrice) - 1)
			PriceChangePerc := math.Round(PriceChange * 100)

			log.Println("the price Change is :", PriceChangePerc)

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

			} else if PriceChangePerc >= PriceDiscrepancy {

				log.Println("item before update : ", existingItem)

				u.DB.Create(&models.ItemHistoryChange{
					ItemID:         existingItem.ID,
					NewItemCreated: false,
					OldPrice:       existingItem.OriginalPrice,
					NewPrice:       item.OriginalPrice,
					OldAvailable:   existingItem.Available,
					NewAvailable:   item.Available,
					OldMenuItemID:  existingItem.MenuItemID,
					NewMenuItemID:  UpdatedMenu.ID,
				})

				u.DB.Model(&existingItem).Updates(models.Item{
					OriginalPrice: item.OriginalPrice,
					Available:     item.Available,
					MenuItemID:    UpdatedMenu.ID,
				})

				log.Println("updated item  : ", existingItem)
			}
		}

	}

	u.DB.Where("data_shop_id = ?", dataShopID).Find(&existingItems)

	for _, item := range existingItems {
		if _, ok := existingItemMap[item.ListingID]; !ok && item.MenuItemID != OutOfProductionID {
			if OutOfProductionID == 0 {
				Menu := models.CreateMenuItem(models.MenuItem{
					ShopMenuID: Shop.ShopMenu.ID,
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

	return nil
}

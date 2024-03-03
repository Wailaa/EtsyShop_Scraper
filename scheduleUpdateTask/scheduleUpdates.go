package scheduleUpdates

import (
	"EtsyScraper/controllers"
	initializer "EtsyScraper/init"
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type UpdateDB struct {
	DB *gorm.DB
}

func NewUpdateDB(DB *gorm.DB) *UpdateDB {
	return &UpdateDB{DB}
}

func ScheduleScrapUpdate() error {
	c := cron.New()
	_, err := c.AddFunc("41 12 * * * ", func() {
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

		NewSoldItems := updatedShop.TotalSales - Shop.TotalSales
		NewAdmirers := updatedShop.Admirers - Shop.Admirers

		if NewSoldItems > 0 && Shop.HasSoldHistory {

			Task := &models.TaskSchedule{
				IsScrapeFinished:     false,
				IsPaginationScrapped: false,
				CurrentPage:          0,
				LastPage:             0,
				UpdateSoldItems:      NewSoldItems,
			}
			UpdateSoldItems(&Shop, Task)
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

		time.Sleep(10 * time.Second)

	}
	log.Println("finished updating Shops")

	return nil
}

func UpdateSoldItems(Shop *models.Shop, Task *models.TaskSchedule) {
	ShopRequest := &models.ShopRequest{}
	controllers.NewShopController(initializer.DB).UpdateSellingHistory(Shop, Task, ShopRequest)
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

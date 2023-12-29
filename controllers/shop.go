package controllers

import (
	"EtsyScraper/models"
	scrap "EtsyScraper/scraping"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shop struct {
	DB *gorm.DB
}

func NewShopController(DB *gorm.DB) *Shop {
	return &Shop{DB}
}
func (s *Shop) CreateNewShop(ctx *gin.Context) {
	currentUserUUID := ctx.MustGet("currentUserUUID").(uuid.UUID)
	var shop *models.CreateNewShopReuest
	if err := ctx.ShouldBindJSON(&shop); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	scrappedShop, err := scrap.ScrapShop(shop.ShopName)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	scrappedShop.CreatedByUserID = currentUserUUID

	tx := s.DB.Begin()

	result := tx.Create(scrappedShop)
	if result.Error != nil {
		tx.Rollback()
		log.Println(err)
		return
	}

	tx.Commit()

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "result": scrappedShop})

}

package models_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"EtsyScraper/models"
)

func TestCreateShop_ValidInput(t *testing.T) {

	inputShop := &models.Shop{
		Name:             "TopSeller",
		Description:      "Art is the most beautiful of all lies",
		Location:         "Mars",
		TotalSales:       100,
		JoinedSince:      "2021-01-01",
		LastUpdateTime:   "2022-01-01",
		CreatedByUserID:  uuid.New(),
		Admirers:         124,
		SocialMediaLinks: []string{"@TopSeller.blabla", "TopSeller_21"},
	}

	result := models.CreateShop(inputShop)

	assert.NotNil(t, result)
	assert.Equal(t, inputShop.Name, result.Name)
	assert.Equal(t, inputShop.Description, result.Description)
	assert.Equal(t, inputShop.Location, result.Location)
	assert.Equal(t, inputShop.TotalSales, result.TotalSales)
	assert.Equal(t, inputShop.JoinedSince, result.JoinedSince)
	assert.Equal(t, inputShop.LastUpdateTime, result.LastUpdateTime)
	assert.Equal(t, inputShop.CreatedByUserID, result.CreatedByUserID)
	assert.Equal(t, inputShop.Admirers, result.Admirers)
	assert.Equal(t, inputShop.SocialMediaLinks, result.SocialMediaLinks)
}

func TestCreateShopMenu_SameShopID(t *testing.T) {

	newShopMenu := &models.ShopMenu{
		ShopID:           1,
		TotalItemsAmount: 10,
		Menu:             []models.MenuItem{{}, {}},
	}

	result := models.CreateShopMenu(newShopMenu)

	assert.Equal(t, newShopMenu.ShopID, result.ShopID)
	assert.Equal(t, newShopMenu.TotalItemsAmount, result.TotalItemsAmount)
	assert.Equal(t, len(newShopMenu.Menu), len(result.Menu))

}

func TestCreateMenuItem_ValidInput(t *testing.T) {

	menuItem := models.MenuItem{
		ShopMenuID: 1,
		Category:   "Category",
		SectionID:  "Section",
		Link:       "Link",
		Amount:     10,
	}

	result := models.CreateMenuItem(menuItem)

	assert.Equal(t, menuItem.ShopMenuID, result.ShopMenuID)
	assert.Equal(t, menuItem.Category, result.Category)
	assert.Equal(t, menuItem.SectionID, result.SectionID)
	assert.Equal(t, menuItem.Link, result.Link)
	assert.Equal(t, menuItem.Amount, result.Amount)
}

func TestCreateMenuItem_EmptyInput(t *testing.T) {

	menuItem := models.MenuItem{}

	result := models.CreateMenuItem(menuItem)

	assert.Equal(t, uint(0), result.ShopMenuID)
	assert.Equal(t, "", result.Category)
	assert.Equal(t, "", result.SectionID)
	assert.Equal(t, "", result.Link)
	assert.Equal(t, 0, result.Amount)
}

func TestCreateShopReviews_ValidInput(t *testing.T) {

	input := &models.Reviews{
		ShopID:       100,
		ShopRating:   4.11112222,
		ReviewsCount: 1124,
		ReviewsTopic: []models.ReviewsTopic{{}, {}, {}, {}},
	}

	result := models.CreateShopReviews(input)

	assert.NotNil(t, result)
	assert.Equal(t, input.ShopID, result.ShopID)
	assert.Equal(t, input.ShopRating, result.ShopRating)
	assert.Equal(t, input.ReviewsCount, result.ReviewsCount)
	assert.Equal(t, len(input.ReviewsTopic), len(result.ReviewsTopic))
}

func TestCreateShopMember_ValidInput(t *testing.T) {
	shopMember := &models.ShopMember{
		ShopID: 1,
		Name:   "Biggie",
		Role:   "Chied Of Happinnes",
	}

	result := models.CreateShopMember(shopMember)

	assert.NotNil(t, result)
	assert.Equal(t, shopMember.ShopID, result.ShopID)
	assert.Equal(t, shopMember.Name, result.Name)
	assert.Equal(t, shopMember.Role, result.Role)
}

func TestCreateSoldOutItemWithAllFields(t *testing.T) {

	item := &models.SoldItems{
		Name:       "the Best Item",
		ItemLink:   "https://MyItem.blabla/the_best_item",
		ListingID:  123,
		DataShopID: "12344321",
	}

	result := models.CreateSoldOutItem(item)

	assert.NotNil(t, result)
	assert.Equal(t, "the Best Item", result.Name)
	assert.Equal(t, "https://MyItem.blabla/the_best_item", result.ItemLink)
	assert.False(t, result.Available)
	assert.Equal(t, uint(123), result.ListingID)
	assert.Equal(t, "12344321", result.DataShopID)
}

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ModelsGroup = []interface{}{
	&Account{},
	&Shop{},
	&ShopMenu{},
	&MenuItem{},
	&Reviews{},
	&ShopMember{},
	&ReviewsTopic{},
	&Item{},
	&SoldItems{},
}

type Shop struct {
	gorm.Model
	Name             string   `json:"shop_name" gorm:"type:varchar(100);not null"`
	Description      string   `json:"shop_description" gorm:"type:varchar(255);not null"`
	Location         string   `json:"location" gorm:"type:varchar(50);not null"`
	TotalSales       int      `json:"shop_total_sales" gorm:"not null"`
	JoinedSince      string   `json:"joined_since" gorm:"type:varchar(100);not null"`
	LastUpdateTime   string   `json:"last_update_time" gorm:"type:varchar(155);not null"`
	Admirers         int      `json:"admirers" gorm:"not null"`
	SocialMediaLinks []string `json:"social_media_links" gorm:"serializer:json"`
	HasSoldHistory   bool     `json:"-" `

	Member   []ShopMember `json:"shop_member" gorm:"foreignKey:ShopID;references:ID"`
	ShopMenu ShopMenu     `json:"shop_menu" gorm:"foreignKey:ShopID;references:ID"`
	Reviews  Reviews      `json:"shop_reviews" gorm:"foreignKey:ShopID;references:ID"`

	CreatedByUserID uuid.UUID `json:"-" gorm:"type:uuid"`
	Followers       []Account `json:"-" gorm:"many2many:account_shop_following;"`
}

type ResponseSoldItemInfo struct {
	Name           string
	ItemID         uint
	OriginalPrice  float64
	CurrencySymbol string
	SalePrice      float64
	DiscoutPercent string
	ItemLink       string
	SoldQauntity   int
}
type SoldItems struct {
	gorm.Model
	Name       string `gorm:"-"`
	ItemLink   string `gorm:"-"`
	ItemID     uint   `gorm:"index"`
	ListingID  uint
	DataShopID string
}

type Item struct {
	gorm.Model
	Name           string
	OriginalPrice  float64
	CurrencySymbol string
	SalePrice      float64
	DiscoutPercent string
	Available      bool
	ItemLink       string
	MenuItemID     uint
	ListingID      uint
	DataShopID     string
	SoldUnits      []SoldItems `gorm:"foreignKey:ItemID"`
}

type MenuItem struct {
	gorm.Model
	ShopMenuID uint   `json:"-"`
	Category   string `json:"category_name"`
	SectionID  string `json:"selection_id"`
	Link       string `json:"link"`
	Amount     int    `json:"item_amount"`
	Items      []Item `json:"category_item" gorm:"foreignKey:MenuItemID;"`
}

type ShopMenu struct {
	gorm.Model
	ShopID            uint       `json:"-"`
	TotalItemsAmmount int        `json:"total_items_amount"`
	Menu              []MenuItem `json:"items_category" gorm:"foreignKey:ShopMenuID"`
}
type Reviews struct {
	gorm.Model
	ShopID       uint           `json:"-"`
	ShopRating   float64        `json:"shop_rate"`
	ReviewsCount int            `json:"reviews_count"`
	ReviewsTopic []ReviewsTopic `json:"reviews_mentions" gorm:"foreignKey:ReviewsID"`
}

type ReviewsTopic struct {
	gorm.Model
	ReviewsID    uint   `json:"-"`
	Keyword      string `json:"keyword"`
	KeywordCount int    `json:"keyword_count"`
}

type ShopMember struct {
	gorm.Model
	ShopID uint   `json:"-"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type CreateNewShopReuest struct {
	ShopName string `json:"new_shop_name"`
}

type FollowShopRequest struct {
	FollowShopName string `json:"follow_shop"`
}

type UnFollowShopRequest struct {
	UnFollowShopName string `json:"unfollow_shop"`
}
type TaskSchedule struct {
	IsScrapped bool
	FirstPage  int
	LastPage   int
}

func CreateShop(newShop *Shop) *Shop {
	Shop := &Shop{
		Name:             newShop.Name,
		Description:      newShop.Description,
		Location:         newShop.Location,
		TotalSales:       newShop.TotalSales,
		JoinedSince:      newShop.JoinedSince,
		LastUpdateTime:   newShop.LastUpdateTime,
		CreatedByUserID:  newShop.CreatedByUserID,
		Admirers:         newShop.Admirers,
		SocialMediaLinks: newShop.SocialMediaLinks,
	}
	return Shop
}

func CreateShopMenu(newShopMenu *ShopMenu) *ShopMenu {
	NewShopMenu := &ShopMenu{
		ShopID:            newShopMenu.ShopID,
		TotalItemsAmmount: newShopMenu.TotalItemsAmmount,
		Menu:              newShopMenu.Menu,
	}
	return NewShopMenu
}

func CreateMenuItem(menuItem *MenuItem) *MenuItem {
	newMenuItem := &MenuItem{
		ShopMenuID: menuItem.ShopMenuID,
		Category:   menuItem.Category,
		SectionID:  menuItem.SectionID,
		Link:       menuItem.Link,
		Amount:     menuItem.Amount,
	}
	return newMenuItem
}

func CreateShopReviews(shopReviews *Reviews) *Reviews {
	newShopReviews := &Reviews{
		ShopID:       shopReviews.ShopID,
		ShopRating:   shopReviews.ShopRating,
		ReviewsCount: shopReviews.ReviewsCount,
		ReviewsTopic: shopReviews.ReviewsTopic,
	}
	return newShopReviews
}

func CreateShopMember(shopMember *ShopMember) *ShopMember {
	NewMember := &ShopMember{
		ShopID: shopMember.ShopID,
		Name:   shopMember.Name,
		Role:   shopMember.Role,
	}
	return NewMember
}

func CreateSoldItemInfo(Item *Item) *ResponseSoldItemInfo {
	newSoldItem := &ResponseSoldItemInfo{
		Name:           Item.Name,
		ItemID:         Item.ID,
		OriginalPrice:  Item.OriginalPrice,
		CurrencySymbol: Item.CurrencySymbol,
		SalePrice:      Item.SalePrice,
		DiscoutPercent: Item.DiscoutPercent,
		ItemLink:       Item.ItemLink,
	}
	return newSoldItem
}
func CreateSoldOutItem(item *SoldItems) *Item {
	SoldOutItem := &Item{
		Name:       item.Name,
		ItemLink:   item.ItemLink,
		Available:  false,
		ListingID:  item.ListingID,
		DataShopID: item.DataShopID,
	}
	return SoldOutItem
}

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
	&CreateShopTaskQueue{},
	&ShopRequest{},
	&DailyShopSales{},
}

type Shop struct {
	gorm.Model
	Name              string   `json:"shop_name" gorm:"type:varchar(100);not null"`
	Description       string   `json:"shop_description" gorm:"type:varchar(255);not null"`
	Location          string   `json:"location" gorm:"type:varchar(50);not null"`
	TotalSales        int      `json:"shop_total_sales" gorm:"not null"`
	JoinedSince       string   `json:"joined_since" gorm:"type:varchar(100);not null"`
	LastUpdateTime    string   `json:"last_update_time" gorm:"type:varchar(155);not null"`
	Admirers          int      `json:"admirers" gorm:"not null"`
	SocialMediaLinks  []string `json:"social_media_links" gorm:"serializer:json"`
	HasSoldHistory    bool     `json:"-" `
	OnVacation        bool     `json:"-" `
	Revenue           float64  `json:"revenue" gorm:"-"`
	AvarageItemsPrice float64  `json:"avarage_item_price" gorm:"-"`

	Member   []ShopMember `json:"shop_member" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`
	ShopMenu ShopMenu     `json:"shop_menu" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`
	Reviews  Reviews      `json:"shop_reviews" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`

	CreatedByUserID uuid.UUID `json:"-" gorm:"type:uuid"`
	Followers       []Account `json:"-" gorm:"many2many:account_shop_following;constraint:OnDelete:CASCADE;"`
}

type ShopRequest struct {
	gorm.Model
	AccountID uuid.UUID
	ShopName  string `json:"shop_name"`
	Status    string
}

type CreateShopTaskQueue struct {
	gorm.Model
	ShopName  string `json:"shop_name"`
	ShopID    uint
	AccountID uuid.UUID `json:"-" gorm:"type:uuid"`
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
	gorm.Model     `json:"-"`
	Name           string
	OriginalPrice  float64
	CurrencySymbol string
	SalePrice      float64
	DiscoutPercent string
	Available      bool
	ItemLink       string
	MenuItemID     uint `json:"-"`
	ListingID      uint
	DataShopID     string      `json:"-"`
	SoldUnits      []SoldItems `json:"-" gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE;"`
}

type MenuItem struct {
	gorm.Model
	ShopMenuID uint   `json:"-"`
	Category   string `json:"category_name"`
	SectionID  string `json:"selection_id"`
	Link       string `json:"link"`
	Amount     int    `json:"item_amount"`
	Items      []Item `json:"category_item" gorm:"foreignKey:MenuItemID;constraint:OnDelete:CASCADE;"`
}

type ShopMenu struct {
	gorm.Model
	ShopID            uint       `json:"-" `
	TotalItemsAmmount int        `json:"total_items_amount"`
	Menu              []MenuItem `json:"items_category" gorm:"foreignKey:ShopMenuID;constraint:OnDelete:CASCADE;"`
}
type Reviews struct {
	gorm.Model
	ShopID       uint           `json:"-"`
	ShopRating   float64        `json:"shop_rate"`
	ReviewsCount int            `json:"reviews_count"`
	ReviewsTopic []ReviewsTopic `json:"reviews_mentions" gorm:"foreignKey:ReviewsID;constraint:OnDelete:CASCADE;"`
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

type TaskSchedule struct {
	IsScrapeFinished     bool
	IsPaginationScrapped bool
	CurrentPage          int
	LastPage             int
	UpdateSoldItems      int
}

type DailyShopSales struct {
	gorm.Model
	ShopID     uint
	TotalSales int
	Admirers   int
	SoldItems  []byte `gorm:"type:jsonb"`
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

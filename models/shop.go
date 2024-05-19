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
	&ShopRequest{},
	&DailyShopSales{},
	&ItemHistoryChange{},
}

type Shop struct {
	gorm.Model
	Name              string    `json:"shop_name" gorm:"type:varchar(100);not null"`
	Description       string    `json:"shop_description" gorm:"type:varchar(255);not null"`
	Location          string    `json:"location" gorm:"type:varchar(50);not null"`
	TotalSales        int       `json:"shop_total_sales" gorm:"not null"`
	JoinedSince       string    `json:"joined_since" gorm:"type:varchar(100);not null"`
	LastUpdateTime    string    `json:"last_update_time" gorm:"type:varchar(155);not null"`
	Admirers          int       `json:"admirers" gorm:"not null"`
	HasSoldHistory    bool      `json:"-" `
	OnVacation        bool      `json:"-" `
	Revenue           float64   `json:"revenue" gorm:"-"`
	AverageItemsPrice float64   `json:"average_item_price" gorm:"-"`
	CreatedByUserID   uuid.UUID `json:"-" gorm:"type:uuid"`

	SocialMediaLinks []SocialMediaLinks `json:"social_media_links" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`
	Member           []ShopMember       `json:"shop_member" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`
	ShopMenu         ShopMenu           `json:"shop_menu" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`
	Reviews          Reviews            `json:"shop_reviews" gorm:"foreignKey:ShopID;references:ID;constraint:OnDelete:CASCADE;"`
	Followers        []Account          `json:"-" gorm:"many2many:account_shop_following;constraint:OnDelete:CASCADE;"`
}

type SocialMediaLinks struct {
	gorm.Model `json:"-"`
	ShopID     uint   `json:"-"`
	Link       string `json:"link"`
}

type ShopRequest struct {
	gorm.Model
	AccountID uuid.UUID
	ShopName  string `json:"shop_name"`
	Status    string
}

type SoldItems struct {
	gorm.Model `json:"-"`
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
	PriceHistory   []ItemHistoryChange
}

type MenuItem struct {
	gorm.Model `json:"-"`
	ShopMenuID uint   `json:"-"`
	Category   string `json:"category_name"`
	SectionID  string `json:"-"`
	Link       string `json:"link"`
	Amount     int    `json:"item_amount"`
	Items      []Item `json:"-" gorm:"foreignKey:MenuItemID;constraint:OnDelete:CASCADE;"`
}

type ShopMenu struct {
	gorm.Model       `json:"-"`
	ShopID           uint       `json:"-" `
	TotalItemsAmount int        `json:"total_items_amount"`
	Menu             []MenuItem `json:"items_category" gorm:"foreignKey:ShopMenuID;constraint:OnDelete:CASCADE;"`
}
type Reviews struct {
	gorm.Model   `json:"-"`
	ShopID       uint           `json:"-"`
	ShopRating   float64        `json:"shop_rate"`
	ReviewsCount int            `json:"reviews_count"`
	ReviewsTopic []ReviewsTopic `json:"reviews_mentions" gorm:"foreignKey:ReviewsID;constraint:OnDelete:CASCADE;"`
}

type ReviewsTopic struct {
	gorm.Model   `json:"-"`
	ReviewsID    uint   `json:"-"`
	Keyword      string `json:"keyword"`
	KeywordCount int    `json:"keyword_count"`
}

type ShopMember struct {
	gorm.Model `json:"-"`
	ShopID     uint   `json:"-"`
	Name       string `json:"name"`
	Role       string `json:"role"`
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
	ShopID       uint
	TotalSales   int
	Admirers     int
	DailyRevenue float64
	Shop         Shop `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE;"`
}

type ItemHistoryChange struct {
	gorm.Model
	ItemID         uint
	NewItemCreated bool
	OldPrice       float64
	NewPrice       float64
	OldAvailable   bool
	NewAvailable   bool
	OldMenuItemID  uint
	NewMenuItemID  uint
}

func CreateMenuItem(menuItem MenuItem) MenuItem {
	newMenuItem := MenuItem{
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

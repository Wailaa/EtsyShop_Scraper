package models

import "gorm.io/gorm"

type Shop struct {
	gorm.Model
	Name             string   `json:"shop_name" gorm:"type:varchar(100);not null"`
	Description      string   `json:"shop_description" gorm:"type:varchar(255);not null"`
	Location         string   `json:"location" gorm:"type:varchar(50);not null"`
	TotalSales       int      `json:"shop_total_sales" gorm:"not null"`
	JoinedSince      string   `json:"joined_since" gorm:"type:varchar(100);not null"`
	ShopMenu         ShopMenu `json:"shop_menu" gorm:"-"`
	LastUpdateTime   string   `json:"last_update_time" gorm:"type:varchar(155);not null"`
	Admirers         int      `json:"admirers" gorm:"not null"`
	Reviews          Reviews  `json:"shop_reviews" gorm:"-"`
	SocialMediaLinks []string `json:"social_media_links" gorm:"serializer:json"`
	Member           Members  `json:"shop_member" gorm:"-"`
}

type MenuItem struct {
	gorm.Model
	ShopMenuID uint
	Category   string `json:"category_name"`
	SectionID  string `json:"selection_id"`
	Link       string `json:"link"`
	Amount     int    `json:"item_amount"`
}

type ShopMenu struct {
	gorm.Model
	ShopID            uint
	TotalItemsAmmount int        `json:"total_items_amount"`
	Menu              []MenuItem `json:"shop_item_id" gorm:"serializer:json"`
}
type Reviews struct {
	ShopRating   float64        `json:"shop_rate"`
	ReviewsCount int            `json:"reviews_count"`
	ReviewsTopic map[string]int `json:"reviews_mentions"`
}

type Members struct {
	Amount  int             `json:"amount"`
	Members map[int]*Member `json:"members"`
}
type Member struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type CreateNewShopReuest struct {
	ShopName string `json:"new_shop_name"`
}

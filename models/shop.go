package models

type Shop struct {
	Name             string    `json:"shop_name"`
	Description      string    `json:"shop_description"`
	Location         string    `json:"location"`
	TotalSales       int       `json:"shop_total_sales"`
	JoinedSince      string    `json:"joined_since"`
	ShopMenu         *ShopMenu `json:"shop_menu"`
	LastUpdateTime   string    `json:"last_update_time"`
	Admirers         int       `json:"admirers"`
	Reviews          *Reviews  `json:"shop_reviews"`
	SocialMediaLinks []string  `json:"social_media_links"`
	Member           *Members  `json:"shop_member"`
}

type MenuItem struct {
	Category  string `json:"category_name"`
	SectionID string `json:"selection_id"`
	Link      string `json:"link"`
	Amount    int    `json:"item_amount"`
}

type ShopMenu struct {
	Menu map[int]*MenuItem `json:"shop_item_id"`
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

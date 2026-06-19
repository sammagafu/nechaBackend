package dto

type HotelResponse struct {
	ID           string   `json:"id"`
	Code         string   `json:"code"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Address      string   `json:"address"`
	City         string   `json:"city"`
	Location     string   `json:"location"`
	Country      string   `json:"country"`
	Zone         string   `json:"zone"`
	Phone        string   `json:"phone"`
	Initials     string   `json:"initials"`
	LogoURL      string   `json:"logo_url"`
	ReferralCode string   `json:"referral_code"`
	Services     []string `json:"services"`
	IsVerified   bool     `json:"is_verified"`
}

type ProductResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	BrandName   string `json:"brand_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Badge       string `json:"badge"`
	Price       int64  `json:"price"`
	Currency    string `json:"currency"`
	ImageURL    string `json:"image_url"`
	Stock       int    `json:"stock"`
	IsFeatured  bool   `json:"is_featured"`
}

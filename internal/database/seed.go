package database

import (
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const defaultHotelCode = "SEACLIFF24"

func Seed(db *gorm.DB, seedDemoUsers bool) error {
	if err := seedHotels(db); err != nil {
		return err
	}
	if err := upgradeDiscoverService(db); err != nil {
		return err
	}
	if err := seedCatalogSync(db); err != nil {
		return err
	}
	if err := seedDiscovery(db); err != nil {
		return err
	}
	if err := seedAdmin(db, seedDemoUsers); err != nil {
		return err
	}
	if seedDemoUsers {
		if err := seedTestUsers(db); err != nil {
			return err
		}
	}
	return deactivateWelcomeAlert(db)
}

func deactivateWelcomeAlert(db *gorm.DB) error {
	return db.Model(&models.SystemAlert{}).
		Where("title = ?", "Welcome to Necha").
		Update("is_active", false).Error
}

func upgradeDiscoverService(db *gorm.DB) error {
	var hotels []models.Hotel
	if err := db.Find(&hotels).Error; err != nil {
		return err
	}
	for _, hotel := range hotels {
		hasDiscover := false
		for _, s := range hotel.Services {
			if s == "discover" {
				hasDiscover = true
				break
			}
		}
		if hasDiscover {
			continue
		}
		for _, s := range hotel.Services {
			if s == "nearby" {
				hotel.Services = append(hotel.Services, "discover")
				if err := db.Model(&hotel).Update("services", hotel.Services).Error; err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func seedHotels(db *gorm.DB) error {
	return ensureCanonicalHotels(db)
}

func seaCliffHotel() models.Hotel {
	return models.Hotel{
		Code:         "SEACLIFF24",
		Slug:         "sea-cliff",
		Name:         "Sea Cliff Hotel",
		Description:  "Personal care, beauty & wellness — delivered to your room.",
		Address:      "House No. 2, Salous St, Africana, Mbezi Beach",
		City:         "Dar es Salaam",
		Location:     "Masaki",
		Country:      "Tanzania",
		Zone:         "E",
		Phone:        "+255784455439",
		Initials:     "SC",
		LogoURL:      "",
		ReferralCode: "SEACLIFF24",
		Services:     models.StringSlice{"spa", "restaurant", "bar", "gym", "discover"},
		IsVerified:   true,
		KkooappID:    "kkoo-prop-seacliff-001",
		IsActive:     true,
	}
}

func nechaDemoHotel() models.Hotel {
	return models.Hotel{
		Code:         "NECHA-DEMO",
		Slug:         "necha-demo",
		Name:         "Necha Demo Hotel",
		Description:  "Demo partner property for sales and QA — same catalogue as Sea Cliff.",
		Address:      "House No. 2, Salous St, Africana, Mbezi Beach",
		City:         "Dar es Salaam",
		Location:     "Mbezi Beach",
		Country:      "Tanzania",
		Zone:         "E",
		Phone:        "+255784455439",
		Initials:     "ND",
		LogoURL:      "/necha-logo.svg",
		ReferralCode: "NECHA-DEMO",
		Services:     models.StringSlice{"spa", "restaurant", "bar", "shop"},
		IsVerified:   true,
		KkooappID:    "kkoo-prop-necha-demo-001",
		IsActive:     true,
	}
}

func partnerSeedHotels() []models.Hotel {
	return []models.Hotel{
		{
			Code: "SERENA24", Slug: "serena-dsm", Name: "Serena Hotel Dar es Salaam",
			Description: "Five-star waterfront hotel in the heart of Oysterbay.",
			Address: "Ohio Street", City: "Dar es Salaam", Location: "Oysterbay", Country: "Tanzania", Zone: "E",
			Phone: "+255222112233", Initials: "SE", LogoURL: "", ReferralCode: "SERENA24",
			Services: models.StringSlice{"spa", "restaurant", "bar", "discover"}, IsVerified: true,
			KkooappID: "kkoo-prop-serena-001", IsActive: true,
		},
		{
			Code: "HYATT24", Slug: "hyatt-regency-dsm", Name: "Hyatt Regency Dar es Salaam",
			Description: "Business and leisure hotel on the Kivukoni waterfront.",
			Address: "Kivukoni Front", City: "Dar es Salaam", Location: "City Centre", Country: "Tanzania", Zone: "D",
			Phone: "+255222113344", Initials: "HY", LogoURL: "", ReferralCode: "HYATT24",
			Services: models.StringSlice{"spa", "restaurant", "bar", "gym", "discover"}, IsVerified: true,
			KkooappID: "kkoo-prop-hyatt-001", IsActive: true,
		},
		{
			Code: "SLIPWAY24", Slug: "slipway-hotel", Name: "Slipway Hotel",
			Description: "Boutique hotel on the Msasani peninsula with marina views.",
			Address: "Slipway Road", City: "Dar es Salaam", Location: "Msasani", Country: "Tanzania", Zone: "E",
			Phone: "+255222115566", Initials: "SL", LogoURL: "", ReferralCode: "SLIPWAY24",
			Services: models.StringSlice{"restaurant", "bar", "discover"}, IsVerified: true,
			KkooappID: "kkoo-prop-slipway-001", IsActive: true,
		},
		{
			Code: "RAMADAN24", Slug: "ramada-resort", Name: "Ramada Resort Dar es Salaam",
			Description: "Resort property on the Jangwani beach strip.",
			Address: "Jangwani Beach", City: "Dar es Salaam", Location: "Kunduchi", Country: "Tanzania", Zone: "B",
			Phone: "+255222117788", Initials: "RA", LogoURL: "", ReferralCode: "RAMADAN24",
			Services: models.StringSlice{"spa", "restaurant", "gym", "discover"}, IsVerified: true,
			KkooappID: "kkoo-prop-ramada-001", IsActive: true,
		},
	}
}

// ensureCanonicalHotels creates missing seed hotels on every boot (idempotent).
func ensureCanonicalHotels(db *gorm.DB) error {
	if err := ensureHotel(db, seaCliffHotel()); err != nil {
		return err
	}
	if err := ensureHotel(db, nechaDemoHotel()); err != nil {
		return err
	}
	for _, h := range partnerSeedHotels() {
		if err := ensureHotel(db, h); err != nil {
			return err
		}
	}
	return nil
}

func ensureHotel(db *gorm.DB, seed models.Hotel) error {
	var existing models.Hotel
	err := db.Where("slug = ?", seed.Slug).First(&existing).Error
	if err == nil {
		if existing.LogoURL == "/necha-logo.png" || existing.LogoURL == "/logo.svg" {
			if err := db.Model(&existing).Update("logo_url", seed.LogoURL).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	if err := db.Create(&seed).Error; err != nil {
		return err
	}
	return createCatalogProducts(db, seed.ID)
}

type seedProduct struct {
	Slug        string
	BrandName   string
	Name        string
	Category    string
	Badge       string
	Description string
	Price       int64
	Stock       int
	IsFeatured  bool
	ImageURL    string
}

func catalogProducts() []seedProduct {
	return []seedProduct{
		{
			Slug: "lemongrass-body-butter", BrandName: "NECHA NATURALS", Name: "Lemongrass body butter",
			Category: "skin_care", Badge: "african_brand",
			Description: "Whipped shea & lemongrass — deeply nourishing with a fresh citrus scent",
			Price: 44000, Stock: 50, IsFeatured: true, ImageURL: "/assets/assets-3.jpg",
		},
		{
			Slug: "baobab-face-oil", BrandName: "NECHA NATURALS", Name: "Baobab face oil",
			Category: "skin_care", Badge: "african_brand",
			Description: "Pure cold-pressed baobab — rich in vitamins A, D & E for skin and hair",
			Price: 52000, Stock: 40, IsFeatured: true, ImageURL: "/assets/assets-5.jpg",
		},
		{
			Slug: "turmeric-soap", BrandName: "NECHA NATURALS", Name: "Turmeric soap",
			Category: "personal_care", Badge: "best_seller",
			Description: "Brightening turmeric bar — reduces dark spots and evens skin tone naturally",
			Price: 18000, Stock: 80, IsFeatured: true, ImageURL: "/assets/assets-4.jpg",
		},
		{
			Slug: "clove-soap", BrandName: "NECHA NATURALS", Name: "Clove soap",
			Category: "personal_care",
			Description: "Natural clove oil soap — antibacterial, antifungal and deeply cleansing",
			Price: 16000, Stock: 60, IsFeatured: true, ImageURL: "/assets/1.jpg",
		},
		{
			Slug: "lemongrass-bar-soap", BrandName: "NECHA NATURALS", Name: "Lemongrass bar soap",
			Category: "personal_care", Badge: "new",
			Description: "Refreshing antiseptic bar soap with natural lemongrass and a light citrus scent",
			Price: 15000, Stock: 70, ImageURL: "/assets/2.jpg",
		},
		{
			Slug: "baobab-bar-soap", BrandName: "NECHA NATURALS", Name: "Baobab bar soap",
			Category: "personal_care",
			Description: "Natural bar soap with baobab extract — deeply cleansing and moisturising",
			Price: 17000, Stock: 55, ImageURL: "/assets/3.jpg",
		},
		{
			Slug: "hibiscus-toner", BrandName: "NECHA NATURALS", Name: "Hibiscus glow toner",
			Category: "skin_care",
			Description: "Botanical toner with hibiscus and rose water for a balanced, radiant complexion",
			Price: 32000, Stock: 35, IsFeatured: true, ImageURL: "/assets/assets-3.jpg",
		},
		{
			Slug: "moringa-wellness-tea", BrandName: "NECHA NATURALS", Name: "Moringa wellness tea",
			Category: "wellness",
			Description: "Organic moringa leaf blend — antioxidants and calm for evening rituals",
			Price: 22000, Stock: 45, IsFeatured: true, ImageURL: "/assets/3.jpg",
		},
		{
			Slug: "shea-hair-mask", BrandName: "NECHA NATURALS", Name: "Shea repair hair mask",
			Category: "hair_care", Badge: "new",
			Description: "Deep-conditioning mask with shea butter and coconut for dry or treated hair",
			Price: 38000, Stock: 30, IsFeatured: true, ImageURL: "/assets/1.jpg",
		},
		{
			Slug: "citronella-room-mist", BrandName: "NECHA NATURALS", Name: "Citronella room mist",
			Category: "fragrance",
			Description: "Light botanical room spray — citronella, lemongrass and warm cedar notes",
			Price: 28000, Stock: 25, ImageURL: "/assets/assets-5.jpg",
		},
	}
}

func createCatalogProducts(db *gorm.DB, hotelID uuid.UUID) error {
	products := make([]models.Product, 0, len(catalogProducts()))
	for _, p := range catalogProducts() {
		products = append(products, models.Product{
			HotelID:     hotelID,
			Slug:        p.Slug,
			BrandName:   p.BrandName,
			Name:        p.Name,
			Category:    p.Category,
			Badge:       p.Badge,
			Description: p.Description,
			Price:       p.Price,
			Currency:    "TZS",
			Stock:       p.Stock,
			IsFeatured:  p.IsFeatured,
			IsActive:    true,
			ImageURL:    p.ImageURL,
		})
	}
	return db.Create(&products).Error
}

// seedCatalogSync keeps demo data aligned on every API boot (idempotent).
func seedCatalogSync(db *gorm.DB) error {
	if err := ensureDemoHotel(db); err != nil {
		return err
	}
	if err := syncProductImages(db); err != nil {
		return err
	}
	return upsertMissingCatalogProducts(db)
}

func ensureDemoHotel(db *gorm.DB) error {
	return ensureHotel(db, nechaDemoHotel())
}

func syncProductImages(db *gorm.DB) error {
	imageBySlug := map[string]string{}
	for _, p := range catalogProducts() {
		imageBySlug[p.Slug] = p.ImageURL
	}

	var products []models.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}

	for _, product := range products {
		localImage, ok := imageBySlug[product.Slug]
		if !ok {
			continue
		}
		needsUpdate := product.ImageURL == "" ||
			strings.HasPrefix(product.ImageURL, "https://necha.africa") ||
			strings.HasPrefix(product.ImageURL, "http://")
		if needsUpdate {
			if err := db.Model(&product).Update("image_url", localImage).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func upsertMissingCatalogProducts(db *gorm.DB) error {
	var hotel models.Hotel
	if err := db.Where("code = ?", defaultHotelCode).First(&hotel).Error; err != nil {
		return nil
	}

	for _, seed := range catalogProducts() {
		var existing models.Product
		err := db.Where("hotel_id = ? AND slug = ?", hotel.ID, seed.Slug).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		row := models.Product{
			HotelID:     hotel.ID,
			Slug:        seed.Slug,
			BrandName:   seed.BrandName,
			Name:        seed.Name,
			Category:    seed.Category,
			Badge:       seed.Badge,
			Description: seed.Description,
			Price:       seed.Price,
			Currency:    "TZS",
			Stock:       seed.Stock,
			IsFeatured:  seed.IsFeatured,
			IsActive:    true,
			ImageURL:    seed.ImageURL,
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	}

	// Mirror new products onto necha-demo when present.
	var demo models.Hotel
	if err := db.Where("slug = ?", "necha-demo").First(&demo).Error; err != nil {
		return nil
	}
	for _, seed := range catalogProducts() {
		var existing models.Product
		err := db.Where("hotel_id = ? AND slug = ?", demo.ID, seed.Slug).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		row := models.Product{
			HotelID:     demo.ID,
			Slug:        seed.Slug,
			BrandName:   seed.BrandName,
			Name:        seed.Name,
			Category:    seed.Category,
			Badge:       seed.Badge,
			Description: seed.Description,
			Price:       seed.Price,
			Currency:    "TZS",
			Stock:       seed.Stock,
			IsFeatured:  seed.IsFeatured,
			IsActive:    true,
			ImageURL:    seed.ImageURL,
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedDiscovery(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.DiscoveryItem{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	weekend := now.AddDate(0, 0, 3)
	nextWeek := now.AddDate(0, 0, 10)

	items := []models.DiscoveryItem{
		{
			Section: models.DiscoverySectionEvents, Subcategory: "music", Slug: "jazz-by-the-bay",
			Name: "Jazz by the Bay", Description: "Live jazz on the Oysterbay waterfront with local artists and sunset views.",
			Venue: "Oysterbay Beach Club", Location: "Oysterbay", Distance: "1.5km",
			EventStartsAt: &weekend, TicketMode: models.TicketModeReferral, TicketURL: "https://example.com/jazz-bay",
			IsFeatured: true, Status: models.DiscoveryStatusActive, SortOrder: 1,
		},
		{
			Section: models.DiscoverySectionEvents, Subcategory: "cultural", Slug: "makumbusho-art-night",
			Name: "National Museum Art Night", Description: "Evening exhibition of contemporary Tanzanian artists with guided tours.",
			Venue: "National Museum", Location: "City Centre", Distance: "6km",
			EventStartsAt: &nextWeek, TicketMode: models.TicketModeReferral,
			Status: models.DiscoveryStatusActive, SortOrder: 2,
		},
		{
			Section: models.DiscoverySectionEvents, Subcategory: "nightlife", Slug: "jahazi-live-friday",
			Name: "Jahazi Lounge Live", Description: "Friday live music and craft cocktails in Masaki.",
			Venue: "Jahazi Lounge", Location: "Masaki", Distance: "0.6km", Phone: "+255712000003",
			EventStartsAt: &weekend, TicketMode: models.TicketModeNone,
			Status: models.DiscoveryStatusActive, SortOrder: 3,
		},
		{
			Section: models.DiscoverySectionRestaurants, Subcategory: "fine_dining", Slug: "the-rock-restaurant",
			Name: "The Rock Restaurant", Description: "Iconic seafood restaurant perched on a rock off the Masaki shore.",
			Location: "Masaki", Distance: "1.2km", Phone: "+255712000001", PriceHint: "From TZS 80,000",
			IsFeatured: true, Status: models.DiscoveryStatusActive, SortOrder: 1,
		},
		{
			Section: models.DiscoverySectionRestaurants, Subcategory: "rooftop", Slug: "azure-rooftop",
			Name: "Azure Rooftop", Description: "Rooftop dining with panoramic views over Dar es Salaam bay.",
			Location: "Masaki", Distance: "1.0km", PriceHint: "From TZS 65,000",
			Status: models.DiscoveryStatusActive, SortOrder: 2,
		},
		{
			Section: models.DiscoverySectionRestaurants, Subcategory: "local", Slug: "mama-nyama-choma",
			Name: "Mama Nyama Choma", Description: "Authentic Tanzanian grilled meats and ugali — a local favourite.",
			Location: "Kinondoni", Distance: "4km", PriceHint: "From TZS 25,000",
			Status: models.DiscoveryStatusActive, SortOrder: 3,
		},
		{
			Section: models.DiscoverySectionRestaurants, Subcategory: "beachfront", Slug: "coco-beach-grill",
			Name: "Coco Beach Grill", Description: "Beachfront seafood and sundowners on the Msasani peninsula.",
			Location: "Msasani", Distance: "2km", PriceHint: "From TZS 45,000",
			Status: models.DiscoveryStatusActive, SortOrder: 4,
		},
		{
			Section: models.DiscoverySectionTours, Subcategory: "island", Slug: "mbudya-island-day-trip",
			Name: "Mbudya Island Day Trip", Description: "White sand, snorkelling and fresh seafood — a half-day escape from the city.",
			Location: "Mbudya Island", Distance: "Boat from Kunduchi", PriceHint: "From TZS 120,000",
			Phone: "+255712000010", IsFeatured: true, Status: models.DiscoveryStatusActive, SortOrder: 1,
		},
		{
			Section: models.DiscoverySectionTours, Subcategory: "city", Slug: "dar-city-highlights",
			Name: "Dar City Highlights Tour", Description: "Half-day guided tour: Kariakoo market, Askari Monument and harbour views.",
			Location: "City Centre", Distance: "Pickup from hotel", PriceHint: "From TZS 95,000",
			Status: models.DiscoveryStatusActive, SortOrder: 2,
		},
		{
			Section: models.DiscoverySectionTours, Subcategory: "cultural", Slug: "makumbusho-cultural-tour",
			Name: "Makumbusho Cultural Tour", Description: "Village museum visit with traditional dance performances and craft workshops.",
			Location: "Kijitonyama", Distance: "5km", PriceHint: "From TZS 55,000",
			Status: models.DiscoveryStatusActive, SortOrder: 3,
		},
		{
			Section: models.DiscoverySectionTours, Subcategory: "yacht", Slug: "yacht-club-sunset",
			Name: "Yacht Club Sunset Cruise", Description: "Evening cruise on the Indian Ocean with drinks and coastal views.",
			Location: "Msasani Bay", Distance: "2.5km", PriceHint: "From TZS 180,000",
			Phone: "+255712000011", Status: models.DiscoveryStatusActive, SortOrder: 4,
		},
	}

	for i := range items {
		if err := db.Create(&items[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

type seedUser struct {
	Email    string
	Password string
	FullName string
	Phone    string
	Role     models.UserRole
}

func seedAdmin(db *gorm.DB, allowDefaults bool) error {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		if !allowDefaults {
			return nil
		}
		email = "admin@necha.africa"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		if !allowDefaults {
			return nil
		}
		password = "admin12345"
	}
	return ensureUser(db, seedUser{
		Email:    email,
		Password: password,
		FullName: "Necha Admin",
		Role:     models.UserRoleAdmin,
	})
}

func seedTestUsers(db *gorm.DB) error {
	users := []seedUser{
		{
			Email:    envOr("DEMO_GUEST_EMAIL", "guest@necha.africa"),
			Password: envOr("DEMO_GUEST_PASSWORD", "guest12345"),
			FullName: "Demo Guest",
			Phone:    "+255700000001",
			Role:     models.UserRoleCustomer,
		},
		{
			Email:    envOr("DEMO_SHOPPER_EMAIL", "shopper@necha.africa"),
			Password: envOr("DEMO_SHOPPER_PASSWORD", "shop12345"),
			FullName: "Test Shopper",
			Phone:    "+255700000002",
			Role:     models.UserRoleCustomer,
		},
	}
	for _, user := range users {
		if err := ensureUser(db, user); err != nil {
			return err
		}
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func ensureUser(db *gorm.DB, seed seedUser) error {
	var count int64
	if err := db.Model(&models.User{}).Where("email = ?", seed.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Email:        seed.Email,
		PasswordHash: string(hash),
		FullName:     seed.FullName,
		Phone:        seed.Phone,
		Role:         seed.Role,
	}
	return db.Create(&user).Error
}

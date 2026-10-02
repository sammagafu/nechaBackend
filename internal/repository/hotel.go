package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type HotelRepository struct {
	db *gorm.DB
}

func NewHotelRepository(db *gorm.DB) *HotelRepository {
	return &HotelRepository{db: db}
}

func (r *HotelRepository) FindBySlug(slug string) (*models.Hotel, error) {
	var hotel models.Hotel
	err := r.db.Where("slug = ? AND is_active = ?", slug, true).First(&hotel).Error
	if err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *HotelRepository) FindByCode(code string) (*models.Hotel, error) {
	var hotel models.Hotel
	err := r.db.Where("code = ? AND is_active = ?", code, true).First(&hotel).Error
	if err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *HotelRepository) FindByID(id uuid.UUID) (*models.Hotel, error) {
	var hotel models.Hotel
	err := r.db.First(&hotel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *HotelRepository) ListProducts(hotelID uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("hotel_id = ? AND is_active = ?", hotelID, true).Find(&products).Error
	return products, err
}

func (r *HotelRepository) ListFeaturedProducts(hotelID uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("hotel_id = ? AND is_active = ? AND is_featured = ?", hotelID, true, true).Find(&products).Error
	return products, err
}

func (r *HotelRepository) FindProductBySlug(hotelID uuid.UUID, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.Where("hotel_id = ? AND slug = ? AND is_active = ?", hotelID, slug, true).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *HotelRepository) CreateReview(review *models.ProductReview) error {
	return r.db.Create(review).Error
}

func (r *HotelRepository) ListApprovedReviews(productID uuid.UUID, limit int) ([]models.ProductReview, error) {
	if limit <= 0 {
		limit = 50
	}
	var reviews []models.ProductReview
	err := r.db.Where("product_id = ? AND is_approved = ?", productID, true).
		Order("created_at DESC").Limit(limit).Find(&reviews).Error
	return reviews, err
}

func (r *HotelRepository) ListAll() ([]models.Hotel, error) {
	var hotels []models.Hotel
	err := r.db.Order("created_at DESC").Find(&hotels).Error
	return hotels, err
}

func (r *HotelRepository) ListActive() ([]models.Hotel, error) {
	var hotels []models.Hotel
	err := r.db.Where("is_active = ?", true).Order("created_at ASC").Find(&hotels).Error
	return hotels, err
}

func (r *HotelRepository) Create(hotel *models.Hotel) error {
	return r.db.Create(hotel).Error
}

func (r *HotelRepository) Update(hotel *models.Hotel) error {
	return r.db.Save(hotel).Error
}

func (r *HotelRepository) CountProducts(hotelID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Where("hotel_id = ?", hotelID).Count(&count).Error
	return count, err
}

func (r *HotelRepository) CountAllProducts() (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Count(&count).Error
	return count, err
}

func (r *HotelRepository) CountHotels() (int64, error) {
	var count int64
	err := r.db.Model(&models.Hotel{}).Count(&count).Error
	return count, err
}

func (r *HotelRepository) ListAllProducts(hotelID uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("hotel_id = ?", hotelID).Order("created_at DESC").Find(&products).Error
	return products, err
}

func (r *HotelRepository) FindProductByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *HotelRepository) FindActiveProductForHotel(hotelID, productID uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.Where("id = ? AND hotel_id = ? AND is_active = ?", productID, hotelID, true).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *HotelRepository) DecrementProductStock(tx *gorm.DB, productID uuid.UUID, quantity int) error {
	result := tx.Model(&models.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *HotelRepository) RestoreProductStock(productID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return nil
	}
	return r.db.Model(&models.Product{}).
		Where("id = ?", productID).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error
}

func (r *HotelRepository) CreateProduct(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *HotelRepository) UpdateProduct(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *HotelRepository) ListProductsBySupplier(supplierID uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Hotel").
		Where("supplier_id = ?", supplierID).
		Order("name ASC, hotel_id ASC").
		Find(&products).Error
	return products, err
}

func (r *HotelRepository) FindProductByIDForSupplier(productID, supplierID uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Hotel").
		Where("id = ? AND supplier_id = ?", productID, supplierID).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// SupplierSalesBySKU aggregates paid order lines for a supplier's products (last N days).
func (r *HotelRepository) SupplierSalesBySKU(supplierID uuid.UUID, since time.Time, limit int) ([]SupplierSKUAgg, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []SupplierSKUAgg
	err := r.db.Raw(`
		SELECT p.id AS product_id, p.name AS product_name, p.slug, p.currency,
			COALESCE(SUM(oi.quantity), 0)::int AS units_sold,
			COALESCE(SUM(oi.total_price), 0)::bigint AS revenue
		FROM products p
		LEFT JOIN order_items oi ON oi.product_id = p.id
		LEFT JOIN orders o ON o.id = oi.order_id
			AND o.status NOT IN ('pending', 'cancelled', 'failed')
			AND o.created_at >= ?
		WHERE p.supplier_id = ?
		GROUP BY p.id, p.name, p.slug, p.currency
		HAVING COALESCE(SUM(oi.quantity), 0) > 0
		ORDER BY revenue DESC
		LIMIT ?
	`, since, supplierID, limit).Scan(&rows).Error
	return rows, err
}

type SupplierSKUAgg struct {
	ProductID   uuid.UUID
	ProductName string
	Slug        string
	Currency    string
	UnitsSold   int
	Revenue     int64
}

// ListSupplierOrderLines returns recent order lines for products owned by a supplier.
func (r *HotelRepository) ListSupplierOrderLines(supplierID uuid.UUID, limit int) ([]SupplierOrderLineRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var rows []SupplierOrderLineRow
	err := r.db.Raw(`
		SELECT o.id AS order_id, o.kkooapp_ref AS order_ref, o.status AS order_status,
			h.name AS hotel_name, o.customer_name, o.customer_phone, o.room_number,
			p.id AS product_id, oi.name AS product_name, oi.quantity, oi.unit_price,
			oi.total_price, o.currency, o.created_at
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN products p ON p.id = oi.product_id
		JOIN hotels h ON h.id = o.hotel_id
		WHERE p.supplier_id = ?
		ORDER BY o.created_at DESC
		LIMIT ?
	`, supplierID, limit).Scan(&rows).Error
	return rows, err
}

type SupplierOrderLineRow struct {
	OrderID       uuid.UUID
	OrderRef      string
	OrderStatus   string
	HotelName     string
	CustomerName  string
	CustomerPhone string
	RoomNumber    string
	ProductID     uuid.UUID
	ProductName   string
	Quantity      int
	UnitPrice     int64
	TotalPrice    int64
	Currency      string
	CreatedAt     time.Time
}

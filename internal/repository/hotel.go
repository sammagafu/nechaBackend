package repository

import (
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

func (r *HotelRepository) CreateProduct(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *HotelRepository) UpdateProduct(product *models.Product) error {
	return r.db.Save(product).Error
}

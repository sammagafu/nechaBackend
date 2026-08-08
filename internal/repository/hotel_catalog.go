package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type HotelCatalogRepository struct {
	db *gorm.DB
}

func NewHotelCatalogRepository(db *gorm.DB) *HotelCatalogRepository {
	return &HotelCatalogRepository{db: db}
}

func (r *HotelCatalogRepository) UpsertRoom(room *models.HotelRoom) (created bool, err error) {
	var existing models.HotelRoom
	err = r.db.Where("hotel_id = ? AND room_number = ?", room.HotelID, room.RoomNumber).First(&existing).Error
	if err == nil {
		room.ID = existing.ID
		return false, r.db.Save(room).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return true, r.db.Create(room).Error
}

func (r *HotelCatalogRepository) UpsertCategory(category *models.HotelCategory) (created bool, err error) {
	var existing models.HotelCategory
	err = r.db.Where("hotel_id = ? AND slug = ? AND kind = ?", category.HotelID, category.Slug, category.Kind).First(&existing).Error
	if err == nil {
		category.ID = existing.ID
		return false, r.db.Save(category).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return true, r.db.Create(category).Error
}

func (r *HotelCatalogRepository) UpsertMenuItem(item *models.HotelMenuItem) (created bool, err error) {
	var existing models.HotelMenuItem
	err = r.db.Where("hotel_id = ? AND slug = ?", item.HotelID, item.Slug).First(&existing).Error
	if err == nil {
		item.ID = existing.ID
		return false, r.db.Save(item).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return true, r.db.Create(item).Error
}

func (r *HotelCatalogRepository) ListRooms(hotelID uuid.UUID, activeOnly bool) ([]models.HotelRoom, error) {
	var rooms []models.HotelRoom
	q := r.db.Where("hotel_id = ?", hotelID).Order("room_number ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	return rooms, q.Find(&rooms).Error
}

func (r *HotelCatalogRepository) ListCategories(hotelID uuid.UUID, kind models.CategoryKind, activeOnly bool) ([]models.HotelCategory, error) {
	var categories []models.HotelCategory
	q := r.db.Where("hotel_id = ? AND kind = ?", hotelID, kind).Order("sort_order ASC, label ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	return categories, q.Find(&categories).Error
}

func (r *HotelCatalogRepository) ListMenuItems(hotelID uuid.UUID, activeOnly bool) ([]models.HotelMenuItem, error) {
	return r.ListMenuItemsByKind(hotelID, "", activeOnly)
}

func (r *HotelCatalogRepository) ListMenuItemsByKind(hotelID uuid.UUID, menuKind string, activeOnly bool) ([]models.HotelMenuItem, error) {
	var items []models.HotelMenuItem
	q := r.db.Where("hotel_id = ?", hotelID).Order("sort_order ASC, name ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	if menuKind != "" {
		q = q.Where("menu_kind = ?", menuKind)
	}
	return items, q.Find(&items).Error
}

// ListMenuItemsByKinds filters by any of the given menu_kind values.
func (r *HotelCatalogRepository) ListMenuItemsByKinds(hotelID uuid.UUID, menuKinds []string, activeOnly bool) ([]models.HotelMenuItem, error) {
	var items []models.HotelMenuItem
	q := r.db.Where("hotel_id = ?", hotelID).Order("sort_order ASC, name ASC")
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	if len(menuKinds) > 0 {
		q = q.Where("menu_kind IN ?", menuKinds)
	}
	return items, q.Find(&items).Error
}

func (r *HotelCatalogRepository) FindMenuItemByID(id uuid.UUID) (*models.HotelMenuItem, error) {
	var item models.HotelMenuItem
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *HotelCatalogRepository) CreateMenuItem(item *models.HotelMenuItem) error {
	return r.db.Create(item).Error
}

func (r *HotelCatalogRepository) UpdateMenuItem(item *models.HotelMenuItem) error {
	return r.db.Save(item).Error
}

func (r *HotelCatalogRepository) DeleteMenuItem(id uuid.UUID) error {
	return r.db.Delete(&models.HotelMenuItem{}, "id = ?", id).Error
}

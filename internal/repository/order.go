package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

var excludedOrderStatuses = []models.OrderStatus{
	models.OrderStatusCancelled,
	models.OrderStatusFailed,
}

type DailyMetricRow struct {
	Date    string
	Count   int64
	Revenue int64
}

type StatusCountRow struct {
	Status string
	Count  int64
}

type HotelPerformanceRow struct {
	HotelID    uuid.UUID
	HotelName  string
	OrderCount int64
	Revenue    int64
}

type ProductPerformanceRow struct {
	ProductID uuid.UUID
	Name      string
	Quantity  int64
	Revenue   int64
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) CreateTx(tx *gorm.DB, order *models.Order) error {
	return tx.Create(order).Error
}

func (r *OrderRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *OrderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items").Preload("Hotel").First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) List(limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Items").Preload("Hotel").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).Count(&count).Error
	return count, err
}

func (r *OrderRepository) CountByStatus(status models.OrderStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *OrderRepository) SumRevenue() (int64, error) {
	var total int64
	err := r.db.Model(&models.Order{}).
		Where("status NOT IN ?", excludedOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *OrderRepository) CountSince(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).
		Where("created_at >= ?", since).
		Where("status NOT IN ?", excludedOrderStatuses).
		Count(&count).Error
	return count, err
}

func (r *OrderRepository) DailyMetricsSince(since time.Time) ([]DailyMetricRow, error) {
	var rows []DailyMetricRow
	err := r.db.Model(&models.Order{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as revenue").
		Where("created_at >= ?", since).
		Where("status NOT IN ?", excludedOrderStatuses).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *OrderRepository) CountByStatusGrouped() ([]StatusCountRow, error) {
	var rows []StatusCountRow
	err := r.db.Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *OrderRepository) TopHotelsByOrders(limit int) ([]HotelPerformanceRow, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	var rows []HotelPerformanceRow
	err := r.db.Model(&models.Order{}).
		Select("orders.hotel_id, hotels.name as hotel_name, COUNT(*) as order_count, COALESCE(SUM(orders.total_amount), 0) as revenue").
		Joins("JOIN hotels ON hotels.id = orders.hotel_id").
		Where("orders.status NOT IN ?", excludedOrderStatuses).
		Group("orders.hotel_id, hotels.name").
		Order("order_count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *OrderRepository) CountByHotel(hotelID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).
		Where("hotel_id = ?", hotelID).
		Where("status NOT IN ?", excludedOrderStatuses).
		Count(&count).Error
	return count, err
}

func (r *OrderRepository) CountByHotelSince(hotelID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).
		Where("hotel_id = ? AND created_at >= ?", hotelID, since).
		Where("status NOT IN ?", excludedOrderStatuses).
		Count(&count).Error
	return count, err
}

func (r *OrderRepository) SumRevenueByHotel(hotelID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&models.Order{}).
		Where("hotel_id = ?", hotelID).
		Where("status NOT IN ?", excludedOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *OrderRepository) SumRevenueByHotelSince(hotelID uuid.UUID, since time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&models.Order{}).
		Where("hotel_id = ? AND created_at >= ?", hotelID, since).
		Where("status NOT IN ?", excludedOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *OrderRepository) DailyMetricsByHotelSince(hotelID uuid.UUID, since time.Time) ([]DailyMetricRow, error) {
	var rows []DailyMetricRow
	err := r.db.Model(&models.Order{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as revenue").
		Where("hotel_id = ? AND created_at >= ?", hotelID, since).
		Where("status NOT IN ?", excludedOrderStatuses).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *OrderRepository) ListByHotel(hotelID uuid.UUID, limit int) ([]models.Order, error) {
	if limit <= 0 || limit > 50 {
		limit = 8
	}
	var orders []models.Order
	err := r.db.Preload("Items").Preload("Hotel").
		Where("hotel_id = ?", hotelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) TopProductsByHotel(hotelID uuid.UUID, limit int) ([]ProductPerformanceRow, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	var rows []ProductPerformanceRow
	err := r.db.Model(&models.OrderItem{}).
		Select("order_items.product_id, order_items.name, COALESCE(SUM(order_items.quantity), 0) as quantity, COALESCE(SUM(order_items.total_price), 0) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.hotel_id = ?", hotelID).
		Where("orders.status NOT IN ?", excludedOrderStatuses).
		Where("order_items.product_id IS NOT NULL").
		Group("order_items.product_id, order_items.name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *OrderRepository) CountByHotelAndStatus(hotelID uuid.UUID, status models.OrderStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).
		Where("hotel_id = ? AND status = ?", hotelID, status).
		Count(&count).Error
	return count, err
}

func (r *OrderRepository) CountByType(orderType models.OrderType) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).Where("type = ?", orderType).Count(&count).Error
	return count, err
}

func (r *OrderRepository) CountInStatuses(statuses []models.OrderStatus) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&models.Order{}).Where("status IN ?", statuses).Count(&count).Error
	return count, err
}

func (r *OrderRepository) CountToday() (int64, error) {
	start := time.Now().Truncate(24 * time.Hour)
	return r.CountSince(start)
}

func (r *OrderRepository) SumRevenueSince(since time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&models.Order{}).
		Where("created_at >= ?", since).
		Where("status NOT IN ?", excludedOrderStatuses).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

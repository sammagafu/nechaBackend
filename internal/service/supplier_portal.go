package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

// SupplierPortalService exposes supplier-scoped catalogue and sales views (brief §11.3).
type SupplierPortalService struct {
	users     *repository.UserRepository
	suppliers *repository.SupplierRepository
	hotels    *repository.HotelRepository
}

func NewSupplierPortalService(
	users *repository.UserRepository,
	suppliers *repository.SupplierRepository,
	hotels *repository.HotelRepository,
) *SupplierPortalService {
	return &SupplierPortalService{users: users, suppliers: suppliers, hotels: hotels}
}

func (s *SupplierPortalService) supplierIDForUser(userID string) (uuid.UUID, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid user id", apperrors.ErrBadRequest.Status)
	}
	user, err := s.users.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperrors.New(apperrors.ErrNotFound.Code, "user not found", apperrors.ErrNotFound.Status)
		}
		return uuid.Nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
	}
	if user.Role != models.UserRoleSupplier && user.Role != models.UserRoleAdmin {
		return uuid.Nil, apperrors.ErrForbidden
	}
	if user.SupplierID == nil {
		return uuid.Nil, apperrors.New(apperrors.ErrForbidden.Code, "supplier account is not linked to a supplier profile", apperrors.ErrForbidden.Status)
	}
	return *user.SupplierID, nil
}

func (s *SupplierPortalService) Dashboard(userID string) (*dto.SupplierDashboardResponse, error) {
	supplierID, err := s.supplierIDForUser(userID)
	if err != nil {
		return nil, err
	}
	supplier, err := s.suppliers.FindByID(supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrNotFound.Code, "supplier not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load supplier", apperrors.ErrInternal.Status)
	}
	products, err := s.hotels.ListProductsBySupplier(supplierID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list products", apperrors.ErrInternal.Status)
	}
	active := 0
	for _, p := range products {
		if p.IsActive {
			active++
		}
	}
	since := time.Now().AddDate(0, 0, -30)
	sales, err := s.hotels.SupplierSalesBySKU(supplierID, since, 10)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load sales", apperrors.ErrInternal.Status)
	}
	var revenue int64
	var units int
	top := make([]dto.SupplierSKUSalesRow, 0, len(sales))
	currency := "TZS"
	for _, row := range sales {
		revenue += row.Revenue
		units += row.UnitsSold
		if row.Currency != "" {
			currency = row.Currency
		}
		top = append(top, dto.SupplierSKUSalesRow{
			ProductID:   row.ProductID.String(),
			ProductName: row.ProductName,
			Slug:        row.Slug,
			UnitsSold:   row.UnitsSold,
			Revenue:     row.Revenue,
			Currency:    row.Currency,
		})
	}
	orders := 0
	lines, _ := s.hotels.ListSupplierOrderLines(supplierID, 200)
	seen := map[string]struct{}{}
	for _, line := range lines {
		if line.CreatedAt.Before(since) {
			continue
		}
		key := line.OrderID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		orders++
	}
	return &dto.SupplierDashboardResponse{
		Supplier: dto.SupplierSummary{
			ID:           supplier.ID.String(),
			Name:         supplier.Name,
			SupplierType: supplier.SupplierType,
			ContactEmail: supplier.ContactEmail,
			ContactPhone: supplier.ContactPhone,
			IsActive:     supplier.IsActive,
		},
		ProductCount:  len(products),
		ActiveSKUs:    active,
		OrdersLast30:  orders,
		UnitsSold30:   units,
		RevenueLast30: revenue,
		Currency:      currency,
		TopSKUs:       top,
	}, nil
}

func (s *SupplierPortalService) ListProducts(userID string) ([]dto.SupplierProductResponse, error) {
	supplierID, err := s.supplierIDForUser(userID)
	if err != nil {
		return nil, err
	}
	products, err := s.hotels.ListProductsBySupplier(supplierID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list products", apperrors.ErrInternal.Status)
	}
	out := make([]dto.SupplierProductResponse, 0, len(products))
	for _, p := range products {
		out = append(out, toSupplierProductResponse(&p))
	}
	return out, nil
}

func (s *SupplierPortalService) UpdateProduct(userID, productID string, req dto.UpdateSupplierProductRequest) (*dto.SupplierProductResponse, error) {
	supplierID, err := s.supplierIDForUser(userID)
	if err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(productID)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid product id", apperrors.ErrBadRequest.Status)
	}
	product, err := s.hotels.FindProductByIDForSupplier(pid, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrNotFound.Code, "product not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load product", apperrors.ErrInternal.Status)
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "price must be >= 0", apperrors.ErrBadRequest.Status)
		}
		product.Price = *req.Price
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "stock must be >= 0", apperrors.ErrBadRequest.Status)
		}
		product.Stock = *req.Stock
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	if err := s.hotels.UpdateProduct(product); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update product", apperrors.ErrInternal.Status)
	}
	resp := toSupplierProductResponse(product)
	return &resp, nil
}

func (s *SupplierPortalService) ListOrders(userID string) ([]dto.SupplierOrderLineResponse, error) {
	supplierID, err := s.supplierIDForUser(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.hotels.ListSupplierOrderLines(supplierID, 100)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list orders", apperrors.ErrInternal.Status)
	}
	out := make([]dto.SupplierOrderLineResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.SupplierOrderLineResponse{
			OrderID:       row.OrderID.String(),
			OrderRef:      row.OrderRef,
			OrderStatus:   row.OrderStatus,
			HotelName:     row.HotelName,
			CustomerName:  row.CustomerName,
			CustomerPhone: row.CustomerPhone,
			RoomNumber:    row.RoomNumber,
			ProductID:     row.ProductID.String(),
			ProductName:   row.ProductName,
			Quantity:      row.Quantity,
			UnitPrice:     row.UnitPrice,
			TotalPrice:    row.TotalPrice,
			Currency:      row.Currency,
			CreatedAt:     row.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *SupplierPortalService) SalesBySKU(userID string) ([]dto.SupplierSKUSalesRow, error) {
	supplierID, err := s.supplierIDForUser(userID)
	if err != nil {
		return nil, err
	}
	since := time.Now().AddDate(0, 0, -90)
	rows, err := s.hotels.SupplierSalesBySKU(supplierID, since, 50)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load sales", apperrors.ErrInternal.Status)
	}
	out := make([]dto.SupplierSKUSalesRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.SupplierSKUSalesRow{
			ProductID:   row.ProductID.String(),
			ProductName: row.ProductName,
			Slug:        row.Slug,
			UnitsSold:   row.UnitsSold,
			Revenue:     row.Revenue,
			Currency:    row.Currency,
		})
	}
	return out, nil
}

func toSupplierProductResponse(p *models.Product) dto.SupplierProductResponse {
	hotelName := ""
	if p.Hotel.ID != uuid.Nil {
		hotelName = p.Hotel.Name
	}
	return dto.SupplierProductResponse{
		ID:          p.ID.String(),
		HotelID:     p.HotelID.String(),
		HotelName:   hotelName,
		Slug:        p.Slug,
		BrandName:   p.BrandName,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Currency:    p.Currency,
		Stock:       p.Stock,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}
}

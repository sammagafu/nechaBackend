package dto

// SupplierDashboardResponse is the supplier portal home summary (brief §11.3).
type SupplierDashboardResponse struct {
	Supplier      SupplierSummary         `json:"supplier"`
	ProductCount  int                     `json:"product_count"`
	ActiveSKUs    int                     `json:"active_skus"`
	OrdersLast30  int                     `json:"orders_last_30_days"`
	UnitsSold30   int                     `json:"units_sold_last_30_days"`
	RevenueLast30 int64                   `json:"revenue_last_30_days"`
	Currency      string                  `json:"currency"`
	TopSKUs       []SupplierSKUSalesRow   `json:"top_skus"`
}

type SupplierSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SupplierType string `json:"supplier_type"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	IsActive     bool   `json:"is_active"`
}

// SupplierProductResponse is a storefront SKU owned by the supplier, across hotels.
type SupplierProductResponse struct {
	ID          string `json:"id"`
	HotelID     string `json:"hotel_id"`
	HotelName   string `json:"hotel_name"`
	Slug        string `json:"slug"`
	BrandName   string `json:"brand_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Price       int64  `json:"price"`
	Currency    string `json:"currency"`
	Stock       int    `json:"stock"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

// UpdateSupplierProductRequest allows suppliers to edit price/stock/active only.
type UpdateSupplierProductRequest struct {
	Price    *int64 `json:"price"`
	Stock    *int   `json:"stock"`
	IsActive *bool  `json:"is_active"`
}

type SupplierSKUSalesRow struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Slug        string `json:"slug"`
	UnitsSold   int    `json:"units_sold"`
	Revenue     int64  `json:"revenue"`
	Currency    string `json:"currency"`
}

// SupplierOrderLineResponse is a fulfilment row for the supplier's SKUs.
type SupplierOrderLineResponse struct {
	OrderID       string `json:"order_id"`
	OrderRef      string `json:"order_ref"`
	OrderStatus   string `json:"order_status"`
	HotelName     string `json:"hotel_name"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	RoomNumber    string `json:"room_number"`
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	Quantity      int    `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"`
	TotalPrice    int64  `json:"total_price"`
	Currency      string `json:"currency"`
	CreatedAt     string `json:"created_at"`
}

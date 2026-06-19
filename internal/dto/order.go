package dto

type FoodOrderItemRequest struct {
	Name      string `json:"name" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
	UnitPrice int64  `json:"unit_price" validate:"required,min=0"`
	Notes     string `json:"notes"`
}

type FoodOrderRequest struct {
	HotelCode     string                 `json:"hotel_code" validate:"required"`
	CustomerName  string                 `json:"customer_name" validate:"required"`
	CustomerPhone string                 `json:"customer_phone" validate:"required"`
	TableNumber   string                 `json:"table_number"`
	RoomNumber    string                 `json:"room_number"`
	Items         []FoodOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Notes         string                 `json:"notes"`
}

type ProductOrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
	UnitPrice int64  `json:"unit_price"`
}

type ProductOrderRequest struct {
	HotelCode      string                    `json:"hotel_code"`
	CustomerName   string                    `json:"customer_name" validate:"required"`
	CustomerPhone  string                    `json:"customer_phone" validate:"required"`
	CustomerEmail  string                    `json:"customer_email"`
	Address        string                    `json:"address"`
	City           string                    `json:"city"`
	Country        string                    `json:"country"`
	RoomNumber     string                    `json:"room_number"`
	PaymentMethod  string                    `json:"payment_method"`
	Currency       string                    `json:"currency"`
	DeliveryFee    int64                     `json:"delivery_fee"`
	ReturnURL      string                    `json:"return_url"`
	CancelURL      string                    `json:"cancel_url"`
	Items          []ProductOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Notes          string                    `json:"notes"`
}

type OrderItemResponse struct {
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	TotalPrice int64  `json:"total_price"`
	Notes      string `json:"notes,omitempty"`
}

type OrderResponse struct {
	ID            string              `json:"id"`
	HotelID       string              `json:"hotel_id"`
	Type          string              `json:"type"`
	Status        string              `json:"status"`
	KkooappRef    string              `json:"kkooapp_ref"`
	CustomerName  string              `json:"customer_name"`
	CustomerPhone string              `json:"customer_phone"`
	TableNumber   string              `json:"table_number,omitempty"`
	RoomNumber    string              `json:"room_number,omitempty"`
	TotalAmount   int64               `json:"total_amount"`
	Currency      string              `json:"currency"`
	Items         []OrderItemResponse `json:"items"`
	Notes            string `json:"notes,omitempty"`
	PaymentProvider  string `json:"payment_provider,omitempty"`
	PaymentStatus    string `json:"payment_status,omitempty"`
	PaymentRef       string `json:"payment_ref,omitempty"`
	PaymentRequired  bool   `json:"payment_required"`
	PaymentURL       string `json:"payment_url,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type OrderTrackResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	KkooappRef string `json:"kkooapp_ref"`
	UpdatedAt  string `json:"updated_at"`
}

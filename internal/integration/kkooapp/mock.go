package kkooapp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MockClient simulates Kkooapp responses for local development without partner credentials.
type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) ReserveHotelRoom(ctx context.Context, req HotelReservationRequest) (*ReservationResponse, error) {
	return &ReservationResponse{
		Reference: fmt.Sprintf("KK-HOTEL-%s", uuid.New().String()[:8]),
		Status:    "confirmed",
		Message:   "Room reservation confirmed",
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (m *MockClient) ReserveTable(ctx context.Context, req TableReservationRequest) (*ReservationResponse, error) {
	return &ReservationResponse{
		Reference: fmt.Sprintf("KK-TABLE-%s", uuid.New().String()[:8]),
		Status:    "confirmed",
		Message:   "Table reservation confirmed",
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (m *MockClient) PlaceFoodOrder(ctx context.Context, req FoodOrderRequest) (*OrderResponse, error) {
	var total int64
	for _, item := range req.Items {
		total += item.UnitPrice * int64(item.Quantity)
	}
	return &OrderResponse{
		Reference:   fmt.Sprintf("KK-FOOD-%s", uuid.New().String()[:8]),
		Status:      "confirmed",
		TotalAmount: total,
		Currency:    "USD",
		Message:     "Food order placed",
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (m *MockClient) GetOrderStatus(ctx context.Context, reference string) (*OrderStatusResponse, error) {
	return &OrderStatusResponse{
		Reference: reference,
		Status:    "preparing",
		UpdatedAt: time.Now().UTC(),
	}, nil
}

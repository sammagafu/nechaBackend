package selcom

import (
	"context"
	"fmt"
)

type MockClient struct {
	AppURL string
}

func NewMockClient(appURL string) *MockClient {
	return &MockClient{AppURL: appURL}
}

func (m *MockClient) CreateOrderMinimal(ctx context.Context, input CreateOrderMinimalInput) (*CheckoutResult, error) {
	ref := input.OrderID
	if len(ref) > 8 {
		ref = ref[:8]
	}
	return &CheckoutResult{
		Reference:         "MOCK-" + ref,
		PaymentGatewayURL: fmt.Sprintf("%s/payment/mock?order_id=%s", trimRightSlash(m.AppURL), input.OrderID),
		PaymentToken:      "MOCKTOKEN",
	}, nil
}

func (m *MockClient) GetOrderStatus(ctx context.Context, orderID string) (*APIResponse, error) {
	ref := orderID
	if len(ref) > 8 {
		ref = ref[:8]
	}
	return &APIResponse{
		Reference:  "MOCK-" + ref,
		ResultCode: "000",
		Result:     "SUCCESS",
		Message:    "Mock payment completed",
	}, nil
}

func trimRightSlash(value string) string {
	if value == "" {
		return "http://localhost:3000"
	}
	for len(value) > 1 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}

package service

import (
	"testing"

	"github.com/nechaafrica/backend/internal/domain/models"
)

func TestWebhookPaymentOutcome(t *testing.T) {
	cases := []struct {
		in             string
		paymentStatus  string
		orderStatus    models.OrderStatus
		releaseStock   bool
	}{
		{"COMPLETED", PaymentStatusCompleted, models.OrderStatusConfirmed, false},
		{"completed", PaymentStatusCompleted, models.OrderStatusConfirmed, false},
		{"CANCELLED", PaymentStatusCancelled, models.OrderStatusCancelled, true},
		{"USERCANCELED", PaymentStatusCancelled, models.OrderStatusCancelled, true},
		{"FAILED", PaymentStatusFailed, models.OrderStatusFailed, true},
		{"REJECTED", PaymentStatusFailed, models.OrderStatusFailed, true},
		{"EXPIRED", PaymentStatusFailed, models.OrderStatusFailed, true},
		{"PENDING", PaymentStatusPending, "", false},
		{"UNKNOWN", PaymentStatusPending, "", false},
	}
	for _, tc := range cases {
		pay, ord, release := webhookPaymentOutcome(tc.in)
		if pay != tc.paymentStatus || ord != tc.orderStatus || release != tc.releaseStock {
			t.Fatalf("webhookPaymentOutcome(%q) = (%s, %s, %v), want (%s, %s, %v)",
				tc.in, pay, ord, release, tc.paymentStatus, tc.orderStatus, tc.releaseStock)
		}
	}
}

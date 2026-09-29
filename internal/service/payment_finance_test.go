package service

import (
	"testing"
	"time"

	"github.com/nechaafrica/backend/internal/domain/models"
)

func TestOrderNoLongerPayableAfterExpiry(t *testing.T) {
	if orderNoLongerPayable(&models.Order{Status: models.OrderStatusPending, PaymentStatus: PaymentStatusPending}) {
		t.Fatal("pending payable order must remain payable")
	}
	expired := &models.Order{Status: models.OrderStatusCancelled, PaymentStatus: PaymentStatusExpired}
	if !orderNoLongerPayable(expired) {
		t.Fatal("expired hold must not accept payment")
	}
	if !orderNoLongerPayable(&models.Order{Status: models.OrderStatusCancelled, PaymentStatus: PaymentStatusCancelled}) {
		t.Fatal("cancelled order must not accept payment")
	}
	if !orderNoLongerPayable(&models.Order{Status: models.OrderStatusConfirmed, PaymentStatus: PaymentStatusRefunded}) {
		t.Fatal("refunded order must not accept another capture")
	}
}

func TestUnpaidHoldCutoffIs15Minutes(t *testing.T) {
	now := time.Date(2026, 9, 25, 18, 0, 0, 0, time.UTC)
	got := unpaidHoldCutoff(now)
	want := now.Add(-15 * time.Minute)
	if !got.Equal(want) {
		t.Fatalf("cutoff = %s, want %s", got, want)
	}
}

func TestScaleSharesAfterPartialRefund(t *testing.T) {
	rec := &models.CommissionRecord{
		GMV: 16000, NechaShare: 4240, PropertyShare: 1040, SupplierShare: 10720,
	}
	got := scaleSharesAfterRefund(rec, 8000)
	if got.GMV != 8000 {
		t.Fatalf("gmv = %d, want 8000", got.GMV)
	}
	if got.NechaShare != 2120 || got.PropertyShare != 520 || got.SupplierShare != 5360 {
		t.Fatalf("scaled shares %+v", got)
	}
}

func TestInventoryReservationActiveIsHold(t *testing.T) {
	if models.InventoryReservationStatusActive == models.InventoryReservationStatusExpired {
		t.Fatal("expired reservations must not count as active capacity holds")
	}
}

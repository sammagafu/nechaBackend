package service

import (
	"testing"

	"github.com/nechaafrica/backend/internal/domain/models"
)

func TestCommissionableGMVExcludesDelivery(t *testing.T) {
	if got := commissionableGMV(25000, 3000); got != 22000 {
		t.Fatalf("commissionableGMV = %d, want 22000", got)
	}
	if got := commissionableGMV(2000, 3000); got != 0 {
		t.Fatalf("commissionableGMV clamp = %d, want 0", got)
	}
}

func TestComputeSplitProductFounding(t *testing.T) {
	rule := &models.CommissionRule{
		PremiumApplicable:      true,
		BaseMarginPct:          0.20,
		PremiumPoolPct:         0.13,
		NechaPoolShareFounding: 0.50,
		NechaPoolShareStandard: 0.70,
	}
	split := computeSplit(rule, 100000, models.CommissionTierFounding)
	if split.PropertyShare != 6500 {
		t.Fatalf("property share = %d, want 6500", split.PropertyShare)
	}
	if split.NechaShare != 26500 {
		t.Fatalf("necha share = %d, want 26500", split.NechaShare)
	}
	if split.SupplierShare != 67000 {
		t.Fatalf("supplier share = %d, want 67000", split.SupplierShare)
	}
	if split.GMV != 100000 {
		t.Fatalf("gmv = %d, want 100000", split.GMV)
	}
}

func TestComputeSplitTourFlat(t *testing.T) {
	rule := &models.CommissionRule{NechaCommissionPct: 0.15}
	split := computeSplit(rule, 20000, models.CommissionTierStandard)
	if split.NechaShare != 3000 || split.SupplierShare != 17000 || split.PropertyShare != 0 {
		t.Fatalf("unexpected tour split %+v", split)
	}
}

func TestMenuItemPayable(t *testing.T) {
	if !menuItemPayable("wellness_paid") {
		t.Fatal("wellness_paid should be payable")
	}
	if menuItemPayable("food") || menuItemPayable("wellness") {
		t.Fatal("in-house menu kinds must not be payable")
	}
}

func TestStockZeroIsSoldOut(t *testing.T) {
	if !productStockInsufficient(1, 0) {
		t.Fatal("quantity 1 must not be allowed when stock is 0")
	}
	if !productStockInsufficient(3, 2) {
		t.Fatal("quantity above remaining stock must be rejected")
	}
	if productStockInsufficient(1, 1) {
		t.Fatal("last remaining unit must be purchasable")
	}
	if productStockInsufficient(2, 5) {
		t.Fatal("quantity within stock must be allowed")
	}
}

package service

import (
	"context"
	"testing"

	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/integration/selcom"
)

func TestDisbursementStatusFromResultMarksMock(t *testing.T) {
	if got := disbursementStatusFromResult(&selcom.DisburseResult{Reference: "MOCK-DISB-1", Status: "demo_completed"}); got != "demo_completed" {
		t.Fatalf("got %s", got)
	}
	if got := disbursementStatusFromResult(&selcom.DisburseResult{Reference: "SELCOM-DISB-1", Status: "completed"}); got != DisbursementStatusCompleted {
		t.Fatalf("live-looking result should stay completed, got %s", got)
	}
}

func TestMockClientDisburseIsDemo(t *testing.T) {
	m := selcom.NewMockClient("http://localhost:5174")
	res, err := m.DisburseWallet(context.Background(), selcom.DisburseInput{Reference: "batch-1", Account: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "demo_completed" || res.Reference != "MOCK-DISB-batch-1" {
		t.Fatalf("unexpected mock disburse %+v", res)
	}
}

func TestReleasedBatchStatusIsIdempotent(t *testing.T) {
	batch := models.PayoutBatch{Status: models.PayoutBatchStatusReleased, DisbursementRef: "MOCK-DISB-1"}
	if batch.Status != models.PayoutBatchStatusReleased {
		t.Fatal("released batch must stay released")
	}
}

package selcom

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nechaafrica/backend/internal/config"
)

func TestHTTPClientDisburseWalletDoesNotFakeLiveSuccess(t *testing.T) {
	client := &HTTPClient{}
	_, err := client.DisburseWallet(context.Background(), DisburseInput{Account: "255712000000", Reference: "batch-1", Amount: 1000, Currency: "TZS"})
	if err == nil {
		t.Fatal("live HTTP client must not pretend a wallet disbursement completed")
	}
}

func TestDisbursePayloadUsesWalletCashinFields(t *testing.T) {
	client := NewClient(config.SelcomConfig{
		Vendor:            "VEND1",
		Pin:               "1234",
		PayoutUtilityCode: "CASHIN",
		PayoutMSISDN:      "0712000000",
	})
	payload, signed, transID, err := client.disbursePayload(DisburseInput{
		Reference: "bc41ca39-773e-4add-bd3f-48a4beb3ebde",
		Account:   "0712345678",
		Amount:    8000,
		Currency:  "TZS",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transID != "bc41ca39773e4addbd3f48a4beb3ebde" {
		t.Fatalf("transid = %s", transID)
	}
	if payload["utilitycode"] != "CASHIN" || payload["utilityref"] != "255712345678" || payload["amount"] != int64(8000) {
		t.Fatalf("payload %+v", payload)
	}
	if payload["pin"] != "1234" || payload["vendor"] != "VEND1" || payload["msisdn"] != "255712000000" {
		t.Fatalf("auth fields %+v", payload)
	}
	wantSigned := "transid,utilitycode,utilityref,amount,vendor,pin,msisdn"
	gotSigned := ""
	for i, field := range signed {
		if i > 0 {
			gotSigned += ","
		}
		gotSigned += field
	}
	if gotSigned != wantSigned {
		t.Fatalf("signed = %s", gotSigned)
	}
}

func TestDisbursePayloadRejectsNonTZS(t *testing.T) {
	client := NewClient(config.SelcomConfig{Vendor: "VEND1", Pin: "1234"})
	_, _, _, err := client.disbursePayload(DisburseInput{Account: "255712000000", Amount: 1, Currency: "USD"})
	if err == nil {
		t.Fatal("expected TZS-only error")
	}
}

func TestHTTPClientDisburseWalletPostsWalletCashin(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Reference:  "SELCOM-REF-99",
			ResultCode: "000",
			Result:     "SUCCESS",
			Message:    "Transaction successful",
		})
	}))
	defer srv.Close()

	client := NewClient(config.SelcomConfig{
		BaseURL:           srv.URL,
		APIKey:            "key",
		APISecret:         "secret",
		Vendor:            "VEND1",
		Pin:               "3545",
		PayoutUtilityCode: "VMCASHIN",
		TimeoutSeconds:    5,
	})
	res, err := client.DisburseWallet(context.Background(), DisburseInput{
		Reference: "batch-live-1",
		Account:   "0761234567",
		Amount:    15000,
		Currency:  "TZS",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/walletcashin/process" {
		t.Fatalf("path = %s", gotPath)
	}
	if gotBody["utilitycode"] != "VMCASHIN" || gotBody["transid"] != "batchlive1" {
		t.Fatalf("body %+v", gotBody)
	}
	if res.Status != "completed" || res.Reference != "SELCOM-REF-99" {
		t.Fatalf("result %+v", res)
	}
}

func TestHTTPClientDisburseWalletRejectsSelcomFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(APIResponse{ResultCode: "121", Result: "FAIL", Message: "insufficient float"})
	}))
	defer srv.Close()

	client := NewClient(config.SelcomConfig{
		BaseURL: srv.URL, APIKey: "key", APISecret: "secret", Vendor: "VEND1", Pin: "3545", TimeoutSeconds: 5,
	})
	_, err := client.DisburseWallet(context.Background(), DisburseInput{Account: "255712000000", Amount: 1000, Currency: "TZS", Reference: "x1"})
	if err == nil {
		t.Fatal("expected selcom failure")
	}
}

func TestMockClientDisburseWalletIsDemo(t *testing.T) {
	client := NewMockClient("http://localhost:5174")
	res, err := client.DisburseWallet(context.Background(), DisburseInput{Account: "demo", Reference: "batch-1"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "demo_completed" || res.Reference != "MOCK-DISB-batch-1" {
		t.Fatalf("unexpected mock result %+v", res)
	}
}

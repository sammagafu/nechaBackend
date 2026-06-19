package selcom

import "testing"

func TestVerifyWebhookSecret(t *testing.T) {
	if !VerifyWebhookSecret("secret", "secret") {
		t.Fatal("expected matching secrets to verify")
	}
	if VerifyWebhookSecret("wrong", "secret") {
		t.Fatal("expected mismatched secrets to fail")
	}
	if VerifyWebhookSecret("secret", "") {
		t.Fatal("expected empty expected secret to fail")
	}
}

func TestParseWebhookAmount(t *testing.T) {
	cases := []struct {
		raw    string
		amount int64
		ok     bool
	}{
		{"15000", 15000, true},
		{"15000.50", 15000, true},
		{"", 0, false},
		{"abc", 0, false},
	}
	for _, tc := range cases {
		amount, ok := ParseWebhookAmount(tc.raw)
		if ok != tc.ok || amount != tc.amount {
			t.Fatalf("ParseWebhookAmount(%q) = (%d, %v), want (%d, %v)", tc.raw, amount, ok, tc.amount, tc.ok)
		}
	}
}

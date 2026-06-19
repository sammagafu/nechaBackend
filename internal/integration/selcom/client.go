package selcom

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/nechaafrica/backend/internal/config"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

type Client interface {
	CreateOrderMinimal(ctx context.Context, input CreateOrderMinimalInput) (*CheckoutResult, error)
	GetOrderStatus(ctx context.Context, orderID string) (*APIResponse, error)
}

type HTTPClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	vendor     string
	httpClient *http.Client
}

func NewClient(cfg config.SelcomConfig) *HTTPClient {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPClient{
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		vendor:    cfg.Vendor,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *HTTPClient) CreateOrderMinimal(ctx context.Context, input CreateOrderMinimalInput) (*CheckoutResult, error) {
	payload := map[string]interface{}{
		"vendor":           c.vendor,
		"order_id":         input.OrderID,
		"buyer_email":      input.BuyerEmail,
		"buyer_name":       input.BuyerName,
		"buyer_phone":      NormalizePhone(input.BuyerPhone),
		"amount":           input.Amount,
		"currency":         input.Currency,
		"redirect_url":     encodeURL(input.RedirectURL),
		"cancel_url":       encodeURL(input.CancelURL),
		"webhook":          encodeURL(input.WebhookURL),
		"buyer_remarks":    input.BuyerRemarks,
		"merchant_remarks": input.MerchantRemarks,
		"no_of_items":      input.NoOfItems,
	}

	signedFields := []string{
		"vendor", "order_id", "buyer_email", "buyer_name", "buyer_phone",
		"amount", "currency", "redirect_url", "cancel_url", "webhook",
		"buyer_remarks", "merchant_remarks", "no_of_items",
	}

	var resp APIResponse
	if err := c.post(ctx, "/v1/checkout/create-order-minimal", payload, signedFields, &resp); err != nil {
		return nil, err
	}
	if resp.ResultCode != "000" {
		return nil, apperrors.New(apperrors.ErrExternalAPI.Code, resp.Message, apperrors.ErrExternalAPI.Status)
	}
	if len(resp.Data) == 0 {
		return nil, apperrors.New(apperrors.ErrExternalAPI.Code, "selcom returned no checkout data", apperrors.ErrExternalAPI.Status)
	}

	data := resp.Data[0]
	result := &CheckoutResult{Reference: resp.Reference}
	if raw, ok := data["payment_gateway_url"].(string); ok {
		result.PaymentGatewayURL = decodeURL(raw)
	}
	if raw, ok := data["payment_token"].(string); ok {
		result.PaymentToken = raw
	}
	if raw, ok := data["qr"].(string); ok {
		result.QR = raw
	}
	if result.PaymentGatewayURL == "" {
		return nil, apperrors.New(apperrors.ErrExternalAPI.Code, "selcom did not return a payment url", apperrors.ErrExternalAPI.Status)
	}
	return result, nil
}

func (c *HTTPClient) GetOrderStatus(ctx context.Context, orderID string) (*APIResponse, error) {
	query := url.Values{"order_id": {orderID}}
	signedFields := []string{"order_id"}
	var resp APIResponse
	if err := c.get(ctx, "/v1/checkout/order-status", query, signedFields, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HTTPClient) post(ctx context.Context, path string, payload map[string]interface{}, signedFields []string, out *APIResponse) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if err := c.setHeaders(req, signedFields, payload); err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *HTTPClient) get(ctx context.Context, path string, query url.Values, signedFields []string, out *APIResponse) error {
	payload := make(map[string]interface{}, len(query))
	for key, values := range query {
		if len(values) > 0 {
			payload[key] = values[0]
		}
	}
	reqURL := c.baseURL + path + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	if err := c.setHeaders(req, signedFields, payload); err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *HTTPClient) setHeaders(req *http.Request, signedFields []string, payload map[string]interface{}) error {
	timestamp := time.Now().Format(time.RFC3339)
	signingString := buildSigningString(timestamp, signedFields, payload)
	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	if _, err := mac.Write([]byte(signingString)); err != nil {
		return err
	}
	digest := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "SELCOM "+base64.StdEncoding.EncodeToString([]byte(c.apiKey)))
	req.Header.Set("Digest-Method", "HS256")
	req.Header.Set("Digest", digest)
	req.Header.Set("Timestamp", timestamp)
	req.Header.Set("Signed-Fields", strings.Join(signedFields, ","))
	return nil
}

func (c *HTTPClient) do(req *http.Request, out *APIResponse) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrExternalAPI.Code, "selcom request failed", apperrors.ErrExternalAPI.Status)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return apperrors.New(apperrors.ErrExternalAPI.Code, string(raw), resp.StatusCode)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return apperrors.Wrap(err, apperrors.ErrExternalAPI.Code, "invalid selcom response", apperrors.ErrExternalAPI.Status)
	}
	return nil
}

func buildSigningString(timestamp string, signedFields []string, payload map[string]interface{}) string {
	parts := []string{"timestamp=" + timestamp}
	for _, field := range signedFields {
		value := payload[field]
		parts = append(parts, fmt.Sprintf("%s=%v", field, formatSigningValue(value)))
	}
	return strings.Join(parts, "&")
}

func formatSigningValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func encodeURL(raw string) string {
	if raw == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeURL(raw string) string {
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return raw
	}
	return string(decoded)
}

func NormalizePhone(phone string) string {
	digits := strings.Builder{}
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	value := digits.String()
	switch {
	case strings.HasPrefix(value, "255") && len(value) >= 12:
		return value
	case strings.HasPrefix(value, "0") && len(value) >= 10:
		return "255" + value[1:]
	case len(value) == 9:
		return "255" + value
	default:
		return value
	}
}

// SortedFieldNames returns map keys in stable order (used by tests).
func SortedFieldNames(payload map[string]interface{}) []string {
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

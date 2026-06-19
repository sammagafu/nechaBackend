package kkooapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nechaafrica/backend/internal/config"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

type Client interface {
	ReserveHotelRoom(ctx context.Context, req HotelReservationRequest) (*ReservationResponse, error)
	ReserveTable(ctx context.Context, req TableReservationRequest) (*ReservationResponse, error)
	PlaceFoodOrder(ctx context.Context, req FoodOrderRequest) (*OrderResponse, error)
	GetOrderStatus(ctx context.Context, reference string) (*OrderStatusResponse, error)
}

type HTTPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	endpoints  config.KkooappEndpoints
}

func NewClient(cfg config.KkooappConfig) *HTTPClient {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoints: cfg.Endpoints,
	}
}

func (c *HTTPClient) ReserveHotelRoom(ctx context.Context, req HotelReservationRequest) (*ReservationResponse, error) {
	var resp ReservationResponse
	if err := c.do(ctx, http.MethodPost, c.endpoints.HotelReservation, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HTTPClient) ReserveTable(ctx context.Context, req TableReservationRequest) (*ReservationResponse, error) {
	var resp ReservationResponse
	if err := c.do(ctx, http.MethodPost, c.endpoints.TableReservation, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HTTPClient) PlaceFoodOrder(ctx context.Context, req FoodOrderRequest) (*OrderResponse, error) {
	var resp OrderResponse
	if err := c.do(ctx, http.MethodPost, c.endpoints.FoodOrder, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HTTPClient) GetOrderStatus(ctx context.Context, reference string) (*OrderStatusResponse, error) {
	path := strings.ReplaceAll(c.endpoints.OrderStatus, "{id}", reference)
	var resp OrderStatusResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to encode request", apperrors.ErrInternal.Status)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create request", apperrors.ErrInternal.Status)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrExternalAPI.Code, "kkooapp request failed", apperrors.ErrExternalAPI.Status)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrExternalAPI.Code, "failed to read kkooapp response", apperrors.ErrExternalAPI.Status)
	}

	if res.StatusCode >= 400 {
		var apiErr APIErrorResponse
		msg := fmt.Sprintf("kkooapp returned status %d", res.StatusCode)
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Error.Message != "" {
			msg = apiErr.Error.Message
		}
		return apperrors.New(apperrors.ErrExternalAPI.Code, msg, apperrors.ErrExternalAPI.Status)
	}

	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return apperrors.Wrap(err, apperrors.ErrExternalAPI.Code, "invalid kkooapp response", apperrors.ErrExternalAPI.Status)
		}
	}
	return nil
}

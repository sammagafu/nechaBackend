package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type PaymentHandler struct {
	payments    *service.PaymentService
	commissions *service.CommissionService
	selcom      config.SelcomConfig
}

func NewPaymentHandler(payments *service.PaymentService, selcomCfg config.SelcomConfig, commissions *service.CommissionService) *PaymentHandler {
	return &PaymentHandler{payments: payments, selcom: selcomCfg, commissions: commissions}
}

func (h *PaymentHandler) SelcomWebhook(c *fiber.Ctx) error {
	if !h.selcom.MockMode && !selcom.VerifyWebhookSecret(c.Get("X-Webhook-Secret"), h.selcom.WebhookSecret) {
		return response.Fail(c, apperrors.New(apperrors.ErrUnauthorized.Code, "invalid webhook secret", apperrors.ErrUnauthorized.Status))
	}

	var payload selcom.WebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid webhook payload", apperrors.ErrBadRequest.Status))
	}
	if err := h.payments.HandleWebhook(c.Context(), payload); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"received": true})
}

func (h *PaymentHandler) PaymentStatus(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Query("order_id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status))
	}
	paymentStatus, orderStatus, isDemo, err := h.payments.GetPaymentStatus(c.Context(), orderID)
	if err != nil {
		return response.Fail(c, err)
	}
	out := fiber.Map{
		"order_id":        orderID.String(),
		"payment_status":  paymentStatus,
		"order_status":    orderStatus,
		"payment_is_demo": isDemo,
	}
	if isDemo {
		out["payment_disclaimer"] = service.DemoPaymentDisclaimer
	}
	return response.OK(c, out)
}

func (h *PaymentHandler) MockComplete(c *fiber.Ctx) error {
	if !h.selcom.MockMode {
		return response.Fail(c, apperrors.New(apperrors.ErrNotFound.Code, "not found", apperrors.ErrNotFound.Status))
	}

	orderID, err := uuid.Parse(c.Query("order_id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status))
	}
	order, err := h.payments.CompleteMockPayment(c.Context(), orderID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{
		"order_id":           order.ID.String(),
		"payment_status":     order.PaymentStatus,
		"order_status":       order.Status,
		"payment_is_demo":    true,
		"payment_disclaimer": service.DemoPaymentDisclaimer,
		"payment_provider":   order.PaymentProvider,
	})
}

func (h *PaymentHandler) MockRefund(c *fiber.Ctx) error {
	if !h.selcom.MockMode {
		return response.Fail(c, apperrors.New(apperrors.ErrNotFound.Code, "not found", apperrors.ErrNotFound.Status))
	}
	orderID, err := uuid.Parse(c.Query("order_id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status))
	}
	amount, _ := strconv.ParseInt(c.Query("amount"), 10, 64)
	result, err := h.payments.ApplyRefund(orderID, amount, c.Query("key"), c.Query("reason"))
	if err != nil {
		return response.Fail(c, err)
	}
	if h.commissions != nil && !result.Idempotent && result.Order != nil {
		_ = h.commissions.ReverseForRefund(result.Order, result.Amount)
	}
	return response.OK(c, fiber.Map{
		"order_id":           result.OrderID,
		"amount":             result.Amount,
		"refunded_total":     result.RefundedTotal,
		"payment_status":     result.PaymentStatus,
		"idempotent":         result.Idempotent,
		"payment_is_demo":    result.IsDemo,
		"payment_disclaimer": result.Disclaimer,
	})
}

func (h *PaymentHandler) MockReserveCapacity(c *fiber.Ctx) error {
	if !h.selcom.MockMode {
		return response.Fail(c, apperrors.New(apperrors.ErrNotFound.Code, "not found", apperrors.ErrNotFound.Status))
	}
	itemID, err := uuid.Parse(c.Query("item_id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid item id", apperrors.ErrBadRequest.Status))
	}
	qty, _ := strconv.Atoi(c.Query("qty"))
	res, err := h.payments.CreateDemoReservation(itemID, qty, c.Query("expired") == "1")
	if err != nil {
		return response.Fail(c, err)
	}
	active, _ := h.payments.CountActiveReservations(itemID)
	return response.OK(c, fiber.Map{
		"reservation_id":  res.ID.String(),
		"status":          res.Status,
		"expires_at":      res.ExpiresAt,
		"active_holds":    active,
		"payment_is_demo": true,
	})
}

func (h *PaymentHandler) MockExpireReservations(c *fiber.Ctx) error {
	if !h.selcom.MockMode {
		return response.Fail(c, apperrors.New(apperrors.ErrNotFound.Code, "not found", apperrors.ErrNotFound.Status))
	}
	n, err := h.payments.ExpireStaleReservations()
	if err != nil {
		return response.Fail(c, err)
	}
	out := fiber.Map{
		"expired_count":      n,
		"payment_is_demo":    true,
		"payment_disclaimer": service.DemoPaymentDisclaimer,
	}
	if itemID, err := uuid.Parse(c.Query("item_id")); err == nil {
		active, _ := h.payments.CountActiveReservations(itemID)
		out["active_holds"] = active
	}
	return response.OK(c, out)
}

func (h *PaymentHandler) MockExpireHold(c *fiber.Ctx) error {
	if !h.selcom.MockMode {
		return response.Fail(c, apperrors.New(apperrors.ErrNotFound.Code, "not found", apperrors.ErrNotFound.Status))
	}
	orderID, err := uuid.Parse(c.Query("order_id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status))
	}
	order, err := h.payments.ExpireHoldNow(orderID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{
		"order_id":        order.ID.String(),
		"order_status":    order.Status,
		"payment_status":  order.PaymentStatus,
		"stock_held":      order.StockHeld,
		"payment_is_demo": true,
	})
}

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type PaymentHandler struct {
	payments *service.PaymentService
	selcom   config.SelcomConfig
}

func NewPaymentHandler(payments *service.PaymentService, selcomCfg config.SelcomConfig) *PaymentHandler {
	return &PaymentHandler{payments: payments, selcom: selcomCfg}
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
	paymentStatus, orderStatus, err := h.payments.GetPaymentStatus(c.Context(), orderID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{
		"order_id":       orderID.String(),
		"payment_status": paymentStatus,
		"order_status":   orderStatus,
	})
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
		"order_id":       order.ID.String(),
		"payment_status": order.PaymentStatus,
		"order_status":   order.Status,
	})
}

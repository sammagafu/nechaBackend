package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type OrderHandler struct {
	orders *service.OrderService
}

func NewOrderHandler(orders *service.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

func (h *OrderHandler) CreateProduct(c *fiber.Ctx) error {
	var req dto.ProductOrderRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.orders.CreateProductOrder(c.Context(), req, userIDFromCtx(c))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *OrderHandler) CreateFood(c *fiber.Ctx) error {
	var req dto.FoodOrderRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.orders.CreateFoodOrder(c.Context(), req, userIDFromCtx(c))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *OrderHandler) Track(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status))
	}
	result, err := h.orders.Track(c.Context(), id)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

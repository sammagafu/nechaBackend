package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type SupplierPortalHandler struct {
	portal *service.SupplierPortalService
}

func NewSupplierPortalHandler(portal *service.SupplierPortalService) *SupplierPortalHandler {
	return &SupplierPortalHandler{portal: portal}
}

func (h *SupplierPortalHandler) Dashboard(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.portal.Dashboard(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *SupplierPortalHandler) ListProducts(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.portal.ListProducts(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *SupplierPortalHandler) UpdateProduct(c *fiber.Ctx) error {
	var req dto.UpdateSupplierProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid request body", apperrors.ErrBadRequest.Status))
	}
	userID, _ := c.Locals("user_id").(string)
	out, err := h.portal.UpdateProduct(userID, c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *SupplierPortalHandler) ListOrders(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.portal.ListOrders(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *SupplierPortalHandler) Sales(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.portal.SalesBySKU(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

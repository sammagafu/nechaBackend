package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	"github.com/nechaafrica/backend/pkg/response"
)

type DiscoveryHandler struct {
	discovery *service.DiscoveryService
}

func NewDiscoveryHandler(discovery *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{discovery: discovery}
}

func (h *DiscoveryHandler) PortalBySlug(c *fiber.Ctx) error {
	result, err := h.discovery.PortalByHotelSlug(c.Params("slug"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *DiscoveryHandler) SubmitEvent(c *fiber.Ctx) error {
	var req dto.SubmitDiscoveryEventRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.discovery.SubmitEvent(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *DiscoveryHandler) AdminList(c *fiber.Ctx) error {
	result, err := h.discovery.AdminList(c.Query("section"), c.Query("status"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *DiscoveryHandler) AdminGet(c *fiber.Ctx) error {
	result, err := h.discovery.AdminGet(c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *DiscoveryHandler) AdminCreate(c *fiber.Ctx) error {
	var req dto.CreateDiscoveryItemRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.discovery.AdminCreate(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *DiscoveryHandler) AdminUpdate(c *fiber.Ctx) error {
	var req dto.UpdateDiscoveryItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.discovery.AdminUpdate(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	"github.com/nechaafrica/backend/pkg/response"
)

type HotelHandler struct {
	hotels     *service.HotelService
	guestStays *service.GuestStayService
}

func NewHotelHandler(hotels *service.HotelService, guestStays *service.GuestStayService) *HotelHandler {
	return &HotelHandler{hotels: hotels, guestStays: guestStays}
}

func (h *HotelHandler) PartnersLanding(c *fiber.Ctx) error {
	catalogSlug := c.Query("catalog_slug", "sea-cliff")
	result, err := h.hotels.PartnersLanding(catalogSlug)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) GetByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	result, err := h.hotels.GetByCode(code)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) ListProductsByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	featured := c.Query("featured") == "true"
	result, err := h.hotels.ListProductsByCode(code, featured)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	ref := c.Query("ref")
	result, err := h.hotels.GetBySlug(slug, ref)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) ListProducts(c *fiber.Ctx) error {
	slug := c.Params("slug")
	featured := c.Query("featured") == "true"
	result, err := h.hotels.ListProducts(slug, featured)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) GetProduct(c *fiber.Ctx) error {
	slug := c.Params("slug")
	productSlug := c.Params("productSlug")
	result, err := h.hotels.GetProduct(slug, productSlug)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) ListRooms(c *fiber.Ctx) error {
	result, err := h.hotels.ListRooms(c.Params("slug"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) ListMenu(c *fiber.Ctx) error {
	result, err := h.hotels.ListMenu(c.Params("slug"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *HotelHandler) RecordScan(c *fiber.Ctx) error {
	var req dto.HotelScanRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	if h.guestStays == nil {
		return response.OK(c, fiber.Map{"recorded": false})
	}
	if err := h.guestStays.RecordScan(c.Params("slug"), req.Ref, req.Channel, req.ScannedAt, userIDFromCtx(c)); err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, fiber.Map{"recorded": true})
}

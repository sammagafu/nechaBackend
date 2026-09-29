package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	"github.com/nechaafrica/backend/pkg/response"
)

type PlatformHandler struct {
	platform      *service.PlatformService
	guestRequests *service.GuestRequestService
}

func NewPlatformHandler(platform *service.PlatformService, guestRequests *service.GuestRequestService) *PlatformHandler {
	return &PlatformHandler{platform: platform, guestRequests: guestRequests}
}

func (h *PlatformHandler) Settings(c *fiber.Ctx) error {
	result, err := h.platform.Settings()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *PlatformHandler) AdminSettings(c *fiber.Ctx) error {
	result, err := h.platform.Settings()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *PlatformHandler) AdminUpdateSettings(c *fiber.Ctx) error {
	var req dto.UpdatePlatformSettingsRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.platform.UpdateSettings(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *PlatformHandler) SubmitGuestRequest(c *fiber.Ctx) error {
	var req dto.SubmitGuestRequestRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.guestRequests.Submit(c.Params("slug"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *PlatformHandler) AdminListGuestRequests(c *fiber.Ctx) error {
	status := c.Query("status")
	result, err := h.guestRequests.List(status, 200)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

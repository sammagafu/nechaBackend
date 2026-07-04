package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	"github.com/nechaafrica/backend/pkg/response"
)

type InquiryHandler struct {
	inquiries *service.InquiryService
}

func NewInquiryHandler(inquiries *service.InquiryService) *InquiryHandler {
	return &InquiryHandler{inquiries: inquiries}
}

func (h *InquiryHandler) Submit(c *fiber.Ctx) error {
	var req dto.SubmitInquiryRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.inquiries.Submit(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *InquiryHandler) AdminList(c *fiber.Ctx) error {
	result, err := h.inquiries.AdminList(c.Query("status"), c.Query("type"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *InquiryHandler) AdminUpdateStatus(c *fiber.Ctx) error {
	var req struct {
		Status string `json:"status" validate:"required"`
	}
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	if err := h.inquiries.AdminUpdateStatus(c.Params("id"), req.Status); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

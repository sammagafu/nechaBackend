package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type ReservationHandler struct {
	reservations *service.ReservationService
}

func NewReservationHandler(reservations *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservations: reservations}
}

func (h *ReservationHandler) CreateHotel(c *fiber.Ctx) error {
	var req dto.HotelReservationRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.reservations.CreateHotelReservation(c.Context(), req, userIDFromCtx(c))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *ReservationHandler) CreateTable(c *fiber.Ctx) error {
	var req dto.TableReservationRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.reservations.CreateTableReservation(c.Context(), req, userIDFromCtx(c))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *ReservationHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid reservation id", apperrors.ErrBadRequest.Status))
	}
	role, _ := c.Locals("user_role").(string)
	result, err := h.reservations.GetByID(id, userIDFromCtx(c), role == "admin")
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func userIDFromCtx(c *fiber.Ctx) *uuid.UUID {
	raw, ok := c.Locals("user_id").(string)
	if !ok || raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

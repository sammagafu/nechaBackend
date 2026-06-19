package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type AdminHandler struct {
	admin  *service.AdminService
	auth   *service.AuthService
	importSvc *service.ImportService
}

func NewAdminHandler(admin *service.AdminService, auth *service.AuthService, importSvc *service.ImportService) *AdminHandler {
	return &AdminHandler{admin: admin, auth: auth, importSvc: importSvc}
}

func (h *AdminHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	result, err := h.auth.Me(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	result, err := h.admin.Dashboard()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) Analytics(c *fiber.Ctx) error {
	result, err := h.admin.Analytics()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) StoreDashboard(c *fiber.Ctx) error {
	result, err := h.admin.StoreDashboard(c.Params("hotelId"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) ListHotels(c *fiber.Ctx) error {
	result, err := h.admin.ListHotels()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) GetHotel(c *fiber.Ctx) error {
	result, err := h.admin.GetHotel(c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) CreateHotel(c *fiber.Ctx) error {
	var req dto.CreateHotelRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.admin.CreateHotel(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *AdminHandler) UpdateHotel(c *fiber.Ctx) error {
	var req dto.UpdateHotelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.admin.UpdateHotel(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) ListProducts(c *fiber.Ctx) error {
	result, err := h.admin.ListProducts(c.Params("hotelId"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) CreateProduct(c *fiber.Ctx) error {
	var req dto.CreateProductRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.admin.CreateProduct(c.Params("hotelId"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *AdminHandler) UpdateProduct(c *fiber.Ctx) error {
	var req dto.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.admin.UpdateProduct(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) OrderSummary(c *fiber.Ctx) error {
	result, err := h.admin.OrderSummary()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) GetOrder(c *fiber.Ctx) error {
	result, err := h.admin.GetOrder(c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) ListOrders(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result, err := h.admin.ListOrders(limit, offset)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) ListGuestStays(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result, err := h.admin.ListGuestStays(limit, offset)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	var req dto.UpdateStatusRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.admin.UpdateOrderStatus(c.Params("id"), req.Status)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) GetReservation(c *fiber.Ctx) error {
	result, err := h.admin.GetReservation(c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) ListReservations(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result, err := h.admin.ListReservations(limit, offset)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) UpdateReservationStatus(c *fiber.Ctx) error {
	var req dto.UpdateStatusRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.admin.UpdateReservationStatus(c.Params("id"), req.Status)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AdminHandler) ImportCSV(c *fiber.Ctx) error {
	if h.importSvc == nil {
		return response.Fail(c, apperrors.New(apperrors.ErrInternal.Code, "import service unavailable", apperrors.ErrInternal.Status))
	}
	hotelID, err := uuid.Parse(c.Params("hotelId"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status))
	}
	kind := strings.ToLower(strings.TrimSpace(c.Params("kind")))
	file, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "csv file is required", apperrors.ErrBadRequest.Status))
	}
	f, err := file.Open()
	if err != nil {
		return response.Fail(c, err)
	}
	defer f.Close()

	var result *dto.ImportResult
	switch kind {
	case "rooms":
		result, err = h.importSvc.ImportRooms(hotelID, f)
	case "categories":
		result, err = h.importSvc.ImportCategories(hotelID, f)
	case "menu":
		result, err = h.importSvc.ImportMenu(hotelID, f)
	default:
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "kind must be rooms, categories, or menu", apperrors.ErrBadRequest.Status))
	}
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

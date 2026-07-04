package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type CommerceHandler struct {
	commerce *service.CommerceService
	payouts  *service.PayoutService
}

func NewCommerceHandler(commerce *service.CommerceService, payouts *service.PayoutService) *CommerceHandler {
	return &CommerceHandler{commerce: commerce, payouts: payouts}
}

// Influencers
func (h *CommerceHandler) AdminListInfluencers(c *fiber.Ctx) error {
	out, err := h.commerce.ListInfluencers()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminCreateInfluencer(c *fiber.Ctx) error {
	var req dto.CreateInfluencerRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.CreateInfluencer(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, out)
}

func (h *CommerceHandler) AdminUpdateInfluencer(c *fiber.Ctx) error {
	var req dto.UpdateInfluencerRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.UpdateInfluencer(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

// Booking referrals
func (h *CommerceHandler) SubmitBookingReferral(c *fiber.Ctx) error {
	var req dto.CreateBookingReferralRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.CreateBookingReferral(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, out)
}

func (h *CommerceHandler) AdminListBookingReferrals(c *fiber.Ctx) error {
	out, err := h.commerce.ListBookingReferrals()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

// Suppliers
func (h *CommerceHandler) AdminListSuppliers(c *fiber.Ctx) error {
	out, err := h.commerce.ListSuppliers()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminCreateSupplier(c *fiber.Ctx) error {
	var req dto.CreateSupplierRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.CreateSupplier(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, out)
}

func (h *CommerceHandler) AdminUpdateSupplier(c *fiber.Ctx) error {
	var req dto.UpdateSupplierRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.UpdateSupplier(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

// Catalogue
func (h *CommerceHandler) AdminListCatalogue(c *fiber.Ctx) error {
	out, err := h.commerce.ListCatalogue()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminCreateCatalogueItem(c *fiber.Ctx) error {
	var req dto.CreateCatalogueItemRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.CreateCatalogueItem(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, out)
}

func (h *CommerceHandler) AdminUpdateCatalogueItem(c *fiber.Ctx) error {
	var req dto.UpdateCatalogueItemRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.UpdateCatalogueItem(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminSetCatalogueVisibility(c *fiber.Ctx) error {
	var req dto.SetVisibilityRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	if err := h.commerce.SetCatalogueVisibility(c.Params("id"), req); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

// Commission
func (h *CommerceHandler) AdminListCommissionRules(c *fiber.Ctx) error {
	out, err := h.commerce.ListCommissionRules()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminUpsertCommissionRule(c *fiber.Ctx) error {
	var req dto.UpsertCommissionRuleRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.UpsertCommissionRule(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminListCommissionRecords(c *fiber.Ctx) error {
	out, err := h.commerce.ListCommissionRecords()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

// Payouts
func (h *CommerceHandler) AdminListPayoutBatches(c *fiber.Ctx) error {
	out, err := h.payouts.List()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) AdminGeneratePayoutBatches(c *fiber.Ctx) error {
	var req dto.GenerateBatchesRequest
	_ = c.BodyParser(&req)
	until := time.Now()
	if req.Until != "" {
		if t, err := time.Parse(time.RFC3339, req.Until); err == nil {
			until = t
		}
	}
	out, err := h.payouts.GenerateBatches(until)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, out)
}

func (h *CommerceHandler) AdminReleasePayoutBatch(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid batch id", apperrors.ErrBadRequest.Status))
	}
	out, err := h.payouts.Release(id)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

// Rewards
func (h *CommerceHandler) AdminUpsertRewardRule(c *fiber.Ctx) error {
	var req dto.UpsertRewardRuleRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	out, err := h.commerce.UpsertRewardRule(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

// Event log
func (h *CommerceHandler) AdminListEventLog(c *fiber.Ctx) error {
	limit := 200
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	out, err := h.commerce.ListEventLog(limit)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *CommerceHandler) UserRewardBalance(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid user", apperrors.ErrBadRequest.Status))
	}
	balance, ledger, err := h.commerce.RewardBalance(uid)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"balance": balance, "ledger": ledger})
}

type PartnerHandler struct {
	partner *service.PartnerService
}

func NewPartnerHandler(partner *service.PartnerService) *PartnerHandler {
	return &PartnerHandler{partner: partner}
}

func (h *PartnerHandler) Dashboard(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.partner.Dashboard(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *PartnerHandler) ListOrders(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.partner.ListOrders(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *PartnerHandler) ListProducts(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	out, err := h.partner.ListProducts(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

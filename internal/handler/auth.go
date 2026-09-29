package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	"github.com/nechaafrica/backend/pkg/response"
)

type AuthHandler struct {
	auth *service.AuthService
	oauth config.OAuthConfig
}

func NewAuthHandler(auth *service.AuthService, oauth config.OAuthConfig) *AuthHandler {
	return &AuthHandler{auth: auth, oauth: oauth}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.auth.Register(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.auth.Login(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	result, err := h.auth.Me(userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	var req dto.SocialLoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.auth.LoginWithGoogle(c.Context(), h.oauth, req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *AuthHandler) AppleLogin(c *fiber.Ctx) error {
	var req dto.SocialLoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.auth.LoginWithApple(c.Context(), h.oauth, req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

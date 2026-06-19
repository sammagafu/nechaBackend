package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	jwtmanager "github.com/nechaafrica/backend/pkg/jwt"
	"github.com/nechaafrica/backend/pkg/response"
)

func Auth(jwt *jwtmanager.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return response.Fail(c, apperrors.ErrUnauthorized)
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Fail(c, apperrors.ErrUnauthorized)
		}
		claims, err := jwt.Parse(parts[1])
		if err != nil {
			return response.Fail(c, apperrors.ErrUnauthorized)
		}
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", string(claims.Role))
		return c.Next()
	}
}

func OptionalAuth(jwt *jwtmanager.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return c.Next()
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if claims, err := jwt.Parse(parts[1]); err == nil {
				c.Locals("user_id", claims.UserID)
				c.Locals("user_email", claims.Email)
				c.Locals("user_role", string(claims.Role))
			}
		}
		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("user_role").(string)
		if role != string(jwtmanager.RoleAdmin) {
			return response.Fail(c, apperrors.ErrForbidden)
		}
		return c.Next()
	}
}

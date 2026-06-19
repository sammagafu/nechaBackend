package response

import (
	"github.com/gofiber/fiber/v2"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
	Meta    *Meta     `json:"meta,omitempty"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    metaFromCtx(c),
	})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    metaFromCtx(c),
	})
}

func Fail(c *fiber.Ctx, err error) error {
	if appErr, ok := apperrors.IsAppError(err); ok {
		return c.Status(appErr.Status).JSON(ErrorResponse{
			Success: false,
			Error: ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
			Meta: metaFromCtx(c),
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    apperrors.ErrInternal.Code,
			Message: apperrors.ErrInternal.Message,
		},
		Meta: metaFromCtx(c),
	})
}

func metaFromCtx(c *fiber.Ctx) *Meta {
	if id, ok := c.Locals("request_id").(string); ok && id != "" {
		return &Meta{RequestID: id}
	}
	return nil
}

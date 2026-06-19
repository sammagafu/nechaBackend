package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/service"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"github.com/nechaafrica/backend/pkg/response"
)

type MessagingHandler struct {
	notifications *service.NotificationService
	alerts        *service.AlertService
	chat          *service.ChatService
	webhooks      *service.WebhookService
}

func NewMessagingHandler(
	notifications *service.NotificationService,
	alerts *service.AlertService,
	chat *service.ChatService,
	webhooks *service.WebhookService,
) *MessagingHandler {
	return &MessagingHandler{
		notifications: notifications,
		alerts:        alerts,
		chat:          chat,
		webhooks:      webhooks,
	}
}

func (h *MessagingHandler) ListNotifications(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	result, err := h.notifications.ListForUser(userID, 30)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) MarkNotificationsRead(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	var req dto.MarkNotificationsReadRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	if err := h.notifications.MarkRead(userID, req.IDs); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"ok": true})
}

func (h *MessagingHandler) ListActiveAlerts(c *fiber.Ctx) error {
	result, err := h.alerts.ListActive()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminListAlerts(c *fiber.Ctx) error {
	result, err := h.alerts.ListAll()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminCreateAlert(c *fiber.Ctx) error {
	var req dto.CreateSystemAlertRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.alerts.Create(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *MessagingHandler) AdminUpdateAlert(c *fiber.Ctx) error {
	var req dto.CreateSystemAlertRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.alerts.Update(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) StartChat(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	var req dto.StartConversationRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.chat.StartConversation(userID, req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *MessagingHandler) ListMyChats(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	result, err := h.chat.ListUserConversations(userID, 20)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) GetMyChat(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	result, err := h.chat.GetUserConversation(userID, c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) SendMyMessage(c *fiber.Ctx) error {
	userID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	var req dto.SendMessageRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.chat.SendUserMessage(userID, c.Params("id"), req.Body)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *MessagingHandler) AdminListChats(c *fiber.Ctx) error {
	result, err := h.chat.ListConversations(50)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminGetChat(c *fiber.Ctx) error {
	result, err := h.chat.GetConversation(c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminSendChatMessage(c *fiber.Ctx) error {
	adminID, err := requireUserID(c)
	if err != nil {
		return response.Fail(c, err)
	}
	var req dto.SendMessageRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	result, err := h.chat.SendAdminMessage(c.Params("id"), adminID, req.Body)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, result)
}

func (h *MessagingHandler) AdminCloseChat(c *fiber.Ctx) error {
	result, err := h.chat.CloseConversation(c.Params("id"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminListWebhooks(c *fiber.Ctx) error {
	result, err := h.webhooks.ListEndpoints()
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminCreateWebhook(c *fiber.Ctx) error {
	var req dto.CreateWebhookRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	endpoint, secret, err := h.webhooks.CreateEndpoint(req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.Created(c, fiber.Map{
		"endpoint": endpoint,
		"secret":   secret,
	})
}

func (h *MessagingHandler) AdminUpdateWebhook(c *fiber.Ctx) error {
	var req dto.UpdateWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.New(apperrors.ErrBadRequest.Code, "invalid JSON body", apperrors.ErrBadRequest.Status))
	}
	result, err := h.webhooks.UpdateEndpoint(c.Params("id"), req)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) AdminListWebhookDeliveries(c *fiber.Ctx) error {
	result, err := h.webhooks.ListDeliveries(50)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *MessagingHandler) InboundWebhook(c *fiber.Ctx) error {
	secret := c.Get("X-Webhook-Secret")
	if !h.webhooks.VerifyInbound(secret) {
		return response.Fail(c, apperrors.New(apperrors.ErrUnauthorized.Code, "invalid webhook secret", apperrors.ErrUnauthorized.Status))
	}
	var req dto.InboundWebhookRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.Fail(c, err)
	}
	if err := h.webhooks.HandleInbound(req); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"received": true})
}

func requireUserID(c *fiber.Ctx) (uuid.UUID, error) {
	raw, ok := c.Locals("user_id").(string)
	if !ok || raw == "" {
		return uuid.Nil, apperrors.New(apperrors.ErrUnauthorized.Code, "authentication required", apperrors.ErrUnauthorized.Status)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.New(apperrors.ErrUnauthorized.Code, "invalid user", apperrors.ErrUnauthorized.Status)
	}
	return id, nil
}

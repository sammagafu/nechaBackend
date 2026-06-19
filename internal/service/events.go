package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/repository"
)

const (
	EventOrderCreated            = "order.created"
	EventOrderStatusUpdated      = "order.status_updated"
	EventReservationCreated      = "reservation.created"
	EventReservationStatusUpdated = "reservation.status_updated"
	EventChatStarted             = "chat.started"
	EventChatMessage             = "chat.message"
	EventChatReply               = "chat.reply"
)

type EventService struct {
	notifications *NotificationService
	webhooks      *WebhookService
	users         *repository.UserRepository
}

func NewEventService(
	notifications *NotificationService,
	webhooks *WebhookService,
	users *repository.UserRepository,
) *EventService {
	return &EventService{
		notifications: notifications,
		webhooks:      webhooks,
		users:         users,
	}
}

func (s *EventService) OrderCreated(order *models.Order, hotelName string) {
	payload := map[string]interface{}{
		"order_id":   order.ID.String(),
		"hotel_id":   order.HotelID.String(),
		"hotel_name": hotelName,
		"type":       string(order.Type),
		"status":     string(order.Status),
		"total":      order.TotalAmount,
		"currency":   order.Currency,
	}
	s.notifyAdmins("order.created", "New order", fmt.Sprintf("Order from %s — %s %d", hotelName, order.Currency, order.TotalAmount), "/admin/orders", models.NotificationSeverityInfo)
	if order.UserID != nil {
		_ = s.notifications.NotifyUser(*order.UserID, "order.created", "Order placed", fmt.Sprintf("Your order at %s was received.", hotelName), "/orders/"+order.ID.String()+"/track", models.NotificationSeveritySuccess)
	}
	s.webhooks.DispatchEvent(context.Background(), EventOrderCreated, payload)
}

func (s *EventService) OrderStatusUpdated(order *models.Order, hotelName, previousStatus string) {
	payload := map[string]interface{}{
		"order_id":        order.ID.String(),
		"hotel_name":      hotelName,
		"status":          string(order.Status),
		"previous_status": previousStatus,
	}
	s.notifyAdmins("order.status_updated", "Order updated", fmt.Sprintf("Order %s is now %s", order.ID.String()[:8], order.Status), "/admin/orders", models.NotificationSeverityInfo)
	if order.UserID != nil {
		_ = s.notifications.NotifyUser(*order.UserID, "order.status_updated", "Order status updated", fmt.Sprintf("Your order is now %s.", order.Status), "/orders/"+order.ID.String()+"/track", models.NotificationSeverityInfo)
	}
	s.webhooks.DispatchEvent(context.Background(), EventOrderStatusUpdated, payload)
}

func (s *EventService) ReservationCreated(reservation *models.Reservation, hotelName string) {
	payload := map[string]interface{}{
		"reservation_id": reservation.ID.String(),
		"hotel_id":       reservation.HotelID.String(),
		"hotel_name":     hotelName,
		"type":           string(reservation.Type),
		"status":         string(reservation.Status),
	}
	s.notifyAdmins("reservation.created", "New reservation", fmt.Sprintf("Reservation at %s", hotelName), "/admin/reservations", models.NotificationSeverityInfo)
	if reservation.UserID != nil {
		_ = s.notifications.NotifyUser(*reservation.UserID, "reservation.created", "Reservation confirmed", fmt.Sprintf("Your reservation at %s was received.", hotelName), "/reservations/"+reservation.ID.String(), models.NotificationSeveritySuccess)
	}
	s.webhooks.DispatchEvent(context.Background(), EventReservationCreated, payload)
}

func (s *EventService) ReservationStatusUpdated(reservation *models.Reservation, hotelName, previousStatus string) {
	payload := map[string]interface{}{
		"reservation_id":  reservation.ID.String(),
		"hotel_name":      hotelName,
		"status":          string(reservation.Status),
		"previous_status": previousStatus,
	}
	s.notifyAdmins("reservation.status_updated", "Reservation updated", fmt.Sprintf("Reservation is now %s", reservation.Status), "/admin/reservations", models.NotificationSeverityInfo)
	if reservation.UserID != nil {
		_ = s.notifications.NotifyUser(*reservation.UserID, "reservation.status_updated", "Reservation updated", fmt.Sprintf("Your reservation is now %s.", reservation.Status), "/reservations/"+reservation.ID.String(), models.NotificationSeverityInfo)
	}
	s.webhooks.DispatchEvent(context.Background(), EventReservationStatusUpdated, payload)
}

func (s *EventService) ChatStarted(conv *models.Conversation, message string) {
	category := conv.Category
	if category == "" {
		category = conv.Subject
	}
	payload := map[string]interface{}{
		"conversation_id": conv.ID.String(),
		"guest_name":      conv.GuestName,
		"guest_email":     conv.GuestEmail,
		"category":        category,
		"message":         message,
	}
	s.notifyAdmins("chat.started", "New message", fmt.Sprintf("%s · %s", conv.GuestName, category), "/admin/chat", models.NotificationSeverityWarning)
	s.webhooks.DispatchEvent(context.Background(), EventChatStarted, payload)
}

func (s *EventService) ChatMessage(conv *models.Conversation, msg *models.Message) {
	if msg.SenderRole != models.MessageSenderGuest && msg.SenderRole != models.MessageSenderCustomer {
		return
	}
	payload := map[string]interface{}{
		"conversation_id": conv.ID.String(),
		"guest_name":      conv.GuestName,
		"message":         msg.Body,
	}
	s.notifyAdmins("chat.message", "Chat message", fmt.Sprintf("%s sent a message", conv.GuestName), "/admin/chat/"+conv.ID.String(), models.NotificationSeverityInfo)
	s.webhooks.DispatchEvent(context.Background(), EventChatMessage, payload)
}

func (s *EventService) ChatAdminReply(conv *models.Conversation, msg *models.Message) {
	if conv.UserID == nil {
		return
	}
	body := msg.Body
	if len(body) > 120 {
		body = body[:117] + "..."
	}
	link := "/"
	_ = s.notifications.NotifyUser(*conv.UserID, EventChatReply, "New reply from Necha", body, link, models.NotificationSeverityInfo)
	payload := map[string]interface{}{
		"conversation_id": conv.ID.String(),
		"message":         msg.Body,
	}
	s.webhooks.DispatchEvent(context.Background(), EventChatReply, payload)
}

func (s *EventService) notifyAdmins(nType, title, body, link string, severity models.NotificationSeverity) {
	admins, err := s.users.ListByRole(models.UserRoleAdmin)
	if err != nil || len(admins) == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, len(admins))
	for _, a := range admins {
		ids = append(ids, a.ID)
	}
	_ = s.notifications.NotifyUsers(ids, nType, title, body, link, severity)
}

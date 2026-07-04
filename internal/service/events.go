package service

import (
	"context"
	"fmt"
	"strings"

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
	EventInquiryCreated          = "inquiry.created"
	EventGuestRequestCreated     = "guest_request.created"
)

type EventService struct {
	notifications *NotificationService
	webhooks      *WebhookService
	users         *repository.UserRepository
	email         EmailNotifier
	sms           SMSNotifier
	commissions   *CommissionService
	rewards       *repository.RewardRepository
}

// SetCommissionService wires the commission engine after construction (avoids an
// initialisation-order cycle in main.go). When set, order confirmation generates commission
// records and delivery confirmation marks them eligible for payout.
func (s *EventService) SetCommissionService(c *CommissionService) {
	s.commissions = c
}

func (s *EventService) SetRewardRepository(r *repository.RewardRepository) {
	s.rewards = r
}

type EmailNotifier interface {
	NotifyAdmin(subject, body string) error
	Send(to, subject, body string) error
}

type SMSNotifier interface {
	NotifyAdmin(body string) error
	Send(to, body string) error
}

func NewEventService(
	notifications *NotificationService,
	webhooks *WebhookService,
	users *repository.UserRepository,
	email EmailNotifier,
	sms SMSNotifier,
) *EventService {
	return &EventService{
		notifications: notifications,
		webhooks:      webhooks,
		users:         users,
		email:         email,
		sms:           sms,
	}
}

func (s *EventService) OrderCreated(order *models.Order, hotel *models.Hotel) {
	hotelName := ""
	hotelEmail := ""
	hotelPhone := ""
	if hotel != nil {
		hotelName = hotel.Name
		hotelEmail = hotel.Email
		hotelPhone = hotel.Phone
	}
	payload := map[string]interface{}{
		"order_id":   order.ID.String(),
		"hotel_id":   order.HotelID.String(),
		"hotel_name": hotelName,
		"type":       string(order.Type),
		"status":     string(order.Status),
		"total":      order.TotalAmount,
		"currency":   order.Currency,
	}
	title := orderTitle(order, hotelName)
	body := orderEmailBody(order, hotelName)
	smsBody := orderSMSBody(order, hotelName)

	s.notifyAdmins("order.created", title, body, "/admin/orders", models.NotificationSeverityInfo)

	// Email: Necha admin + the hotel (when a contact address is on file).
	if s.email != nil {
		_ = s.email.NotifyAdmin(title, body)
		if hotelEmail != "" {
			_ = s.email.Send(hotelEmail, title, body)
		}
	}
	// SMS: Necha admin + the hotel (when a contact number is on file).
	if s.sms != nil {
		_ = s.sms.NotifyAdmin(smsBody)
		if hotelPhone != "" {
			_ = s.sms.Send(hotelPhone, smsBody)
		}
	}

	if order.UserID != nil {
		_ = s.notifications.NotifyUser(*order.UserID, "order.created", "Order placed", fmt.Sprintf("Your order at %s was received.", hotelName), "/orders/"+order.ID.String()+"/track", models.NotificationSeveritySuccess)
	}
	// Commission record creation is a side effect of order confirmation (brief §8.2).
	// Generate() itself skips free-utility categories (in-house dining, room service).
	if s.commissions != nil {
		_ = s.commissions.Generate(order, hotel)
	}
	s.earnRewards(order)
	s.webhooks.DispatchEvent(context.Background(), EventOrderCreated, payload)
}

func orderTitle(order *models.Order, hotelName string) string {
	switch order.Type {
	case models.OrderTypeFood:
		return "New food/bar order — " + hotelName
	case models.OrderTypeProduct:
		return "New product order — " + hotelName
	default:
		return "New order — " + hotelName
	}
}

func orderEmailBody(order *models.Order, hotelName string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Order at %s\n", hotelName)
	if order.RoomNumber != "" {
		fmt.Fprintf(&b, "Room: %s\n", order.RoomNumber)
	}
	if order.CustomerName != "" {
		fmt.Fprintf(&b, "Guest: %s", order.CustomerName)
		if order.CustomerPhone != "" {
			fmt.Fprintf(&b, " (%s)", order.CustomerPhone)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nItems:\n")
	for _, item := range order.Items {
		fmt.Fprintf(&b, "  %d× %s — %s %d\n", item.Quantity, item.Name, order.Currency, item.TotalPrice)
	}
	fmt.Fprintf(&b, "\nTotal: %s %d\n", order.Currency, order.TotalAmount)
	if order.Notes != "" {
		fmt.Fprintf(&b, "Notes: %s\n", order.Notes)
	}
	return b.String()
}

func orderSMSBody(order *models.Order, hotelName string) string {
	itemCount := 0
	for _, item := range order.Items {
		itemCount += item.Quantity
	}
	room := order.RoomNumber
	if room == "" {
		room = "—"
	}
	return fmt.Sprintf("Necha: new order at %s. Room %s, %d item(s), total %s %d.", hotelName, room, itemCount, order.Currency, order.TotalAmount)
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
	// Delivery/fulfilment confirmation makes commission eligible for payout (brief §2.3).
	if s.commissions != nil && order.Status == models.OrderStatusDelivered {
		_ = s.commissions.MarkFulfilled(order)
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

func (s *EventService) InquiryCreated(inquiry *models.Inquiry) {
	title := "New inquiry: " + inquiry.Type
	body := fmt.Sprintf("%s — %s (%s)", inquiry.Name, inquiry.Email, inquiry.Type)
	if inquiry.Company != "" {
		body = fmt.Sprintf("%s — %s at %s", inquiry.Name, inquiry.Type, inquiry.Company)
	}
	s.notifyAdmins("inquiry.created", title, body, "/admin/inquiries", models.NotificationSeverityInfo)
	if s.email != nil {
		_ = s.email.NotifyAdmin(title, body+"\n\n"+inquiry.Message)
	}
	payload := map[string]interface{}{
		"inquiry_id": inquiry.ID.String(),
		"type":       inquiry.Type,
		"email":      inquiry.Email,
	}
	s.webhooks.DispatchEvent(context.Background(), EventInquiryCreated, payload)
}

func (s *EventService) GuestRequestCreated(req *models.GuestRequest, hotelName string) {
	title := "Guest request at " + hotelName
	body := fmt.Sprintf("%s · room %s · %s\n\n%s", req.GuestName, req.RoomNumber, req.GuestPhone, req.Body)
	s.notifyAdmins("guest_request.created", title, body, "/admin/reservations", models.NotificationSeverityWarning)
	if s.email != nil {
		_ = s.email.NotifyAdmin(title, body)
	}
	payload := map[string]interface{}{
		"guest_request_id": req.ID.String(),
		"hotel_name":       hotelName,
		"category":         req.Category,
	}
	s.webhooks.DispatchEvent(context.Background(), EventGuestRequestCreated, payload)
}

func (s *EventService) earnRewards(order *models.Order) {
	if s.rewards == nil || order.UserID == nil {
		return
	}
	rule, err := s.rewards.ActiveRule()
	if err != nil || rule.PointsPerCurrencyUnit <= 0 {
		return
	}
	points := int64(float64(order.TotalAmount) * rule.PointsPerCurrencyUnit)
	if points <= 0 {
		return
	}
	orderID := order.ID
	_ = s.rewards.AppendEntry(&models.RewardLedgerEntry{
		UserID:    order.UserID,
		OrderID:   &orderID,
		EntryType: models.RewardEntryTypeEarn,
		Points:    points,
		Note:      "order " + order.ID.String()[:8],
	})
}

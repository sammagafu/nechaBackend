package dto

import "time"

type NotificationResponse struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Link      string     `json:"link"`
	Severity  string     `json:"severity"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type NotificationListResponse struct {
	Items       []NotificationResponse `json:"items"`
	UnreadCount int                    `json:"unread_count"`
}

type MarkNotificationsReadRequest struct {
	IDs []string `json:"ids"`
}

type SystemAlertResponse struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Severity string     `json:"severity"`
	Link     string     `json:"link"`
	IsActive bool       `json:"is_active"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

type CreateSystemAlertRequest struct {
	Title    string     `json:"title" validate:"required"`
	Body     string     `json:"body" validate:"required"`
	Severity string     `json:"severity"`
	Link     string     `json:"link"`
	IsActive *bool      `json:"is_active"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

type StartConversationRequest struct {
	Category  string `json:"category" validate:"required"`
	Message   string `json:"message" validate:"required"`
	HotelSlug string `json:"hotel_slug"`
}

type StartConversationResponse struct {
	Conversation ConversationResponse `json:"conversation"`
}

type SendMessageRequest struct {
	Body string `json:"body" validate:"required"`
}

type ConversationResponse struct {
	ID         string    `json:"id"`
	Category   string    `json:"category"`
	Status     string    `json:"status"`
	GuestName  string    `json:"guest_name"`
	GuestEmail string    `json:"guest_email"`
	HotelID    *string   `json:"hotel_id"`
	UserID     *string   `json:"user_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MessageResponse struct {
	ID         string    `json:"id"`
	SenderRole string    `json:"sender_role"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConversationDetailResponse struct {
	Conversation ConversationResponse `json:"conversation"`
	Messages     []MessageResponse    `json:"messages"`
}

type CreateWebhookRequest struct {
	URL         string   `json:"url" validate:"required"`
	Description string   `json:"description"`
	Events      []string `json:"events" validate:"required"`
	IsActive    *bool    `json:"is_active"`
}

type UpdateWebhookRequest struct {
	URL         *string  `json:"url"`
	Description *string  `json:"description"`
	Events      []string `json:"events"`
	IsActive    *bool    `json:"is_active"`
}

type WebhookEndpointResponse struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Events      []string  `json:"events"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type WebhookDeliveryResponse struct {
	ID           string    `json:"id"`
	EndpointID   string    `json:"endpoint_id"`
	EventType    string    `json:"event_type"`
	Status       string    `json:"status"`
	ResponseCode int       `json:"response_code"`
	Attempts     int       `json:"attempts"`
	CreatedAt    time.Time `json:"created_at"`
}

type InboundWebhookRequest struct {
	Event     string                 `json:"event" validate:"required"`
	Reference string                 `json:"reference"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
}

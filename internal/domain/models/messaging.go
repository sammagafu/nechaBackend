package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationSeverity string

const (
	NotificationSeverityInfo    NotificationSeverity = "info"
	NotificationSeveritySuccess NotificationSeverity = "success"
	NotificationSeverityWarning NotificationSeverity = "warning"
	NotificationSeverityAlert   NotificationSeverity = "alert"
)

type Notification struct {
	ID        uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID            `gorm:"type:uuid;index;not null" json:"user_id"`
	Type      string               `gorm:"not null;index" json:"type"`
	Title     string               `gorm:"not null" json:"title"`
	Body      string               `json:"body"`
	Link      string               `json:"link"`
	Severity  NotificationSeverity `gorm:"not null;default:info" json:"severity"`
	ReadAt    *time.Time           `json:"read_at"`
	CreatedAt time.Time            `json:"created_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

type SystemAlert struct {
	ID        uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	Title     string               `gorm:"not null" json:"title"`
	Body      string               `gorm:"not null" json:"body"`
	Severity  NotificationSeverity `gorm:"not null;default:info" json:"severity"`
	Link      string               `json:"link"`
	IsActive  bool                 `gorm:"not null;default:true;index" json:"is_active"`
	StartsAt  *time.Time           `json:"starts_at"`
	EndsAt    *time.Time           `json:"ends_at"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

func (a *SystemAlert) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type ConversationStatus string

const (
	ConversationStatusOpen   ConversationStatus = "open"
	ConversationStatusClosed ConversationStatus = "closed"
)

type Conversation struct {
	ID         uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	Category   string             `gorm:"not null;default:general;index" json:"category"`
	Subject    string             `json:"-"`
	Status     ConversationStatus `gorm:"not null;default:open;index" json:"status"`
	GuestName  string             `json:"guest_name"`
	GuestEmail string             `gorm:"index" json:"guest_email"`
	GuestToken string             `gorm:"index" json:"-"`
	HotelID    *uuid.UUID         `gorm:"type:uuid;index" json:"hotel_id"`
	UserID     *uuid.UUID         `gorm:"type:uuid;index" json:"user_id"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	Messages   []Message          `gorm:"foreignKey:ConversationID" json:"-"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type MessageSenderRole string

const (
	MessageSenderGuest    MessageSenderRole = "guest"
	MessageSenderCustomer MessageSenderRole = "customer"
	MessageSenderAdmin    MessageSenderRole = "admin"
	MessageSenderSystem   MessageSenderRole = "system"
)

type Message struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID uuid.UUID         `gorm:"type:uuid;index;not null" json:"conversation_id"`
	SenderRole     MessageSenderRole `gorm:"not null" json:"sender_role"`
	SenderID       *uuid.UUID        `gorm:"type:uuid" json:"sender_id"`
	Body           string            `gorm:"not null" json:"body"`
	CreatedAt      time.Time         `json:"created_at"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type WebhookEndpoint struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	URL         string      `gorm:"not null" json:"url"`
	Secret      string      `gorm:"not null" json:"-"`
	Description string      `json:"description"`
	Events      StringSlice `gorm:"type:jsonb;not null;default:'[]'" json:"events"`
	IsActive    bool        `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (w *WebhookEndpoint) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

type WebhookDeliveryStatus string

const (
	WebhookDeliveryPending WebhookDeliveryStatus = "pending"
	WebhookDeliverySuccess WebhookDeliveryStatus = "success"
	WebhookDeliveryFailed  WebhookDeliveryStatus = "failed"
)

type WebhookDelivery struct {
	ID           uuid.UUID             `gorm:"type:uuid;primaryKey" json:"id"`
	EndpointID   uuid.UUID             `gorm:"type:uuid;index;not null" json:"endpoint_id"`
	EventType    string                `gorm:"not null;index" json:"event_type"`
	Payload      string                `gorm:"type:text;not null" json:"payload"`
	Status       WebhookDeliveryStatus `gorm:"not null;default:pending;index" json:"status"`
	ResponseCode int                   `json:"response_code"`
	ResponseBody string                `gorm:"type:text" json:"response_body"`
	Attempts     int                   `gorm:"not null;default:0" json:"attempts"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

func (d *WebhookDelivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

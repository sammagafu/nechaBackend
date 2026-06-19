package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type ChatService struct {
	chat   *repository.ChatRepository
	hotels *repository.HotelRepository
	users  *repository.UserRepository
	events *EventService
}

func NewChatService(
	chat *repository.ChatRepository,
	hotels *repository.HotelRepository,
	users *repository.UserRepository,
	events *EventService,
) *ChatService {
	return &ChatService{chat: chat, hotels: hotels, users: users, events: events}
}

func (s *ChatService) StartConversation(userID uuid.UUID, req dto.StartConversationRequest) (*dto.StartConversationResponse, error) {
	category, err := normalizeChatCategory(req.Category)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
	}

	var hotelID *uuid.UUID
	if req.HotelSlug != "" {
		hotel, err := s.hotels.FindBySlug(req.HotelSlug)
		if err == nil {
			hotelID = &hotel.ID
		}
	}

	conv := &models.Conversation{
		Category:   category,
		Status:     models.ConversationStatusOpen,
		GuestName:  user.FullName,
		GuestEmail: user.Email,
		HotelID:    hotelID,
		UserID:     &userID,
	}
	if err := s.chat.CreateConversation(conv); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to start conversation", apperrors.ErrInternal.Status)
	}
	msg := &models.Message{
		ConversationID: conv.ID,
		SenderRole:     models.MessageSenderCustomer,
		SenderID:       &userID,
		Body:           req.Message,
	}
	if err := s.chat.CreateMessage(msg); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to send message", apperrors.ErrInternal.Status)
	}
	if s.events != nil {
		s.events.ChatStarted(conv, req.Message)
	}
	return &dto.StartConversationResponse{
		Conversation: toConversationResponse(conv),
	}, nil
}

func (s *ChatService) GetUserConversation(userID uuid.UUID, id string) (*dto.ConversationDetailResponse, error) {
	conv, err := s.loadOwnedConversation(userID, id)
	if err != nil {
		return nil, err
	}
	return s.conversationDetail(conv)
}

func (s *ChatService) SendUserMessage(userID uuid.UUID, id, body string) (*dto.MessageResponse, error) {
	conv, err := s.loadOwnedConversation(userID, id)
	if err != nil {
		return nil, err
	}
	if conv.Status == models.ConversationStatusClosed {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "conversation is closed", apperrors.ErrBadRequest.Status)
	}
	msg := &models.Message{
		ConversationID: conv.ID,
		SenderRole:     models.MessageSenderCustomer,
		SenderID:       &userID,
		Body:           body,
	}
	if err := s.chat.CreateMessage(msg); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to send message", apperrors.ErrInternal.Status)
	}
	conv.UpdatedAt = msg.CreatedAt
	_ = s.chat.UpdateConversation(conv)
	if s.events != nil {
		s.events.ChatMessage(conv, msg)
	}
	resp := toMessageResponse(msg)
	return &resp, nil
}

func (s *ChatService) ListUserConversations(userID uuid.UUID, limit int) ([]dto.ConversationResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := s.chat.ListConversationsByUser(userID, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list conversations", apperrors.ErrInternal.Status)
	}
	result := make([]dto.ConversationResponse, 0, len(items))
	for _, c := range items {
		result = append(result, toConversationResponse(&c))
	}
	return result, nil
}

func (s *ChatService) ListConversations(limit int) ([]dto.ConversationResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	items, err := s.chat.ListConversations(limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list conversations", apperrors.ErrInternal.Status)
	}
	result := make([]dto.ConversationResponse, 0, len(items))
	for _, c := range items {
		result = append(result, toConversationResponse(&c))
	}
	return result, nil
}

func (s *ChatService) GetConversation(id string) (*dto.ConversationDetailResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid conversation id", apperrors.ErrBadRequest.Status)
	}
	conv, err := s.chat.FindConversation(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "conversation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load conversation", apperrors.ErrInternal.Status)
	}
	return s.conversationDetail(conv)
}

func (s *ChatService) SendAdminMessage(id string, adminID uuid.UUID, body string) (*dto.MessageResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid conversation id", apperrors.ErrBadRequest.Status)
	}
	conv, err := s.chat.FindConversation(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "conversation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load conversation", apperrors.ErrInternal.Status)
	}
	msg := &models.Message{
		ConversationID: conv.ID,
		SenderRole:     models.MessageSenderAdmin,
		SenderID:       &adminID,
		Body:           body,
	}
	if err := s.chat.CreateMessage(msg); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to send message", apperrors.ErrInternal.Status)
	}
	conv.UpdatedAt = msg.CreatedAt
	_ = s.chat.UpdateConversation(conv)
	if s.events != nil {
		s.events.ChatAdminReply(conv, msg)
	}
	resp := toMessageResponse(msg)
	return &resp, nil
}

func (s *ChatService) CloseConversation(id string) (*dto.ConversationResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid conversation id", apperrors.ErrBadRequest.Status)
	}
	conv, err := s.chat.FindConversation(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "conversation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load conversation", apperrors.ErrInternal.Status)
	}
	conv.Status = models.ConversationStatusClosed
	if err := s.chat.UpdateConversation(conv); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to close conversation", apperrors.ErrInternal.Status)
	}
	resp := toConversationResponse(conv)
	return &resp, nil
}

func (s *ChatService) loadOwnedConversation(userID uuid.UUID, id string) (*models.Conversation, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid conversation id", apperrors.ErrBadRequest.Status)
	}
	conv, err := s.chat.FindConversationForUser(uid, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "conversation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load conversation", apperrors.ErrInternal.Status)
	}
	return conv, nil
}

func (s *ChatService) conversationDetail(conv *models.Conversation) (*dto.ConversationDetailResponse, error) {
	messages, err := s.chat.ListMessages(conv.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load messages", apperrors.ErrInternal.Status)
	}
	msgResp := make([]dto.MessageResponse, 0, len(messages))
	for _, m := range messages {
		msgResp = append(msgResp, toMessageResponse(&m))
	}
	return &dto.ConversationDetailResponse{
		Conversation: toConversationResponse(conv),
		Messages:     msgResp,
	}, nil
}

func toConversationResponse(c *models.Conversation) dto.ConversationResponse {
	var hotelID *string
	if c.HotelID != nil {
		s := c.HotelID.String()
		hotelID = &s
	}
	var userID *string
	if c.UserID != nil {
		s := c.UserID.String()
		userID = &s
	}
	category := c.Category
	if category == "" {
		category = c.Subject
	}
	return dto.ConversationResponse{
		ID:         c.ID.String(),
		Category:   category,
		Status:     string(c.Status),
		GuestName:  c.GuestName,
		GuestEmail: c.GuestEmail,
		HotelID:    hotelID,
		UserID:     userID,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func toMessageResponse(m *models.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:         m.ID.String(),
		SenderRole: string(m.SenderRole),
		Body:       m.Body,
		CreatedAt:  m.CreatedAt,
	}
}

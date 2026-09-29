package repository

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateConversation(c *models.Conversation) error {
	return r.db.Create(c).Error
}

func (r *ChatRepository) FindConversation(id uuid.UUID) (*models.Conversation, error) {
	var c models.Conversation
	err := r.db.First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ChatRepository) FindConversationForUser(id, userID uuid.UUID) (*models.Conversation, error) {
	var c models.Conversation
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ChatRepository) ListConversationsByUser(userID uuid.UUID, limit int) ([]models.Conversation, error) {
	var items []models.Conversation
	q := r.db.Where("user_id = ?", userID).Order("updated_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&items).Error
	return items, err
}

func (r *ChatRepository) ListConversations(limit int) ([]models.Conversation, error) {
	var items []models.Conversation
	q := r.db.Order("updated_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&items).Error
	return items, err
}

func (r *ChatRepository) UpdateConversation(c *models.Conversation) error {
	return r.db.Save(c).Error
}

func (r *ChatRepository) CreateMessage(m *models.Message) error {
	return r.db.Create(m).Error
}

func (r *ChatRepository) ListMessages(conversationID uuid.UUID) ([]models.Message, error) {
	var items []models.Message
	err := r.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&items).Error
	return items, err
}

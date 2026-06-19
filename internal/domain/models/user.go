package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleAdmin    UserRole = "admin"
)

type AuthProvider string

const (
	AuthProviderEmail  AuthProvider = "email"
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderApple  AuthProvider = "apple"
)

type User struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string       `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string       `json:"-"`
	FullName     string       `gorm:"not null" json:"full_name"`
	Phone        string       `json:"phone"`
	Role         UserRole     `gorm:"not null;default:customer" json:"role"`
	AuthProvider AuthProvider `gorm:"not null;default:email" json:"auth_provider"`
	ProviderID   string       `gorm:"index" json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

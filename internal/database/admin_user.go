package database

import (
	"errors"
	"strings"

	"github.com/nechaafrica/backend/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UpsertSuperUser creates an admin user or updates an existing account when force is true.
func UpsertSuperUser(db *gorm.DB, email, password, fullName string, force bool) (created bool, err error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return false, errors.New("email is required")
	}
	if password == "" {
		return false, errors.New("password is required")
	}
	if fullName == "" {
		fullName = "Necha Admin"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	var user models.User
	findErr := db.Where("email = ?", email).First(&user).Error
	if findErr == nil {
		if !force {
			return false, nil
		}
		user.PasswordHash = string(hash)
		user.FullName = fullName
		user.Role = models.UserRoleAdmin
		if user.AuthProvider == "" {
			user.AuthProvider = models.AuthProviderEmail
		}
		return false, db.Save(&user).Error
	}
	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return false, findErr
	}

	user = models.User{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         models.UserRoleAdmin,
		AuthProvider: models.AuthProviderEmail,
	}
	return true, db.Create(&user).Error
}

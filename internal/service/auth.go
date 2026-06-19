package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	jwtmanager "github.com/nechaafrica/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	users      *repository.UserRepository
	jwt        *jwtmanager.Manager
	guestStays *GuestStayService
}

func NewAuthService(users *repository.UserRepository, jwt *jwtmanager.Manager, guestStays *GuestStayService) *AuthService {
	return &AuthService{users: users, jwt: jwt, guestStays: guestStays}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	_, err := s.users.FindByEmail(req.Email)
	if err == nil {
		return nil, apperrors.New(apperrors.ErrConflict.Code, "email already registered", apperrors.ErrConflict.Status)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to check email", apperrors.ErrInternal.Status)
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to hash password", apperrors.ErrInternal.Status)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Role:         models.UserRoleCustomer,
		AuthProvider: models.AuthProviderEmail,
	}
	if err := s.users.Create(user); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create user", apperrors.ErrInternal.Status)
	}

	if s.guestStays != nil {
		s.guestStays.RecordFromContext(req.HotelContext, user.ID, models.GuestStaySourceRegister)
	}

	token, err := s.jwt.Generate(user.ID.String(), user.Email, jwtmanager.Role(user.Role))
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to generate token", apperrors.ErrInternal.Status)
	}

	return &dto.AuthResponse{
		AccessToken: token,
		User:        toUserResponse(user),
	}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrUnauthorized.Code, "invalid credentials", apperrors.ErrUnauthorized.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
	}

	if err := s.ensurePasswordLoginAllowed(user); err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.New(apperrors.ErrUnauthorized.Code, "invalid credentials", apperrors.ErrUnauthorized.Status)
	}

	token, err := s.jwt.Generate(user.ID.String(), user.Email, jwtmanager.Role(user.Role))
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to generate token", apperrors.ErrInternal.Status)
	}

	return &dto.AuthResponse{
		AccessToken: token,
		User:        toUserResponse(user),
	}, nil
}

func (s *AuthService) Me(userID string) (*dto.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrUnauthorized.Code, "invalid session", apperrors.ErrUnauthorized.Status)
	}
	user, err := s.users.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "user not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func toUserResponse(u *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:           u.ID.String(),
		Email:        u.Email,
		FullName:     u.FullName,
		Phone:        u.Phone,
		Role:         string(u.Role),
		AuthProvider: string(u.AuthProvider),
	}
}

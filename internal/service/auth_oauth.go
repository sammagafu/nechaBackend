package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/integration/oauth"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	jwtmanager "github.com/nechaafrica/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *AuthService) LoginWithGoogle(ctx context.Context, cfg config.OAuthConfig, req dto.SocialLoginRequest) (*dto.AuthResponse, error) {
	profile, err := oauth.VerifyGoogleIDToken(ctx, req.IDToken, cfg.GoogleClientID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrUnauthorized.Code, "invalid google token", apperrors.ErrUnauthorized.Status)
	}
	if profile.Email == "" {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "google account email is required", apperrors.ErrBadRequest.Status)
	}
	if !profile.Verified {
		return nil, apperrors.New(apperrors.ErrUnauthorized.Code, "google email is not verified", apperrors.ErrUnauthorized.Status)
	}
	name := strings.TrimSpace(profile.FullName)
	if name == "" {
		name = strings.Split(profile.Email, "@")[0]
	}
	return s.loginOrCreateSocialUser(models.AuthProviderGoogle, profile.Subject, profile.Email, name, req.HotelContext)
}

func (s *AuthService) LoginWithApple(ctx context.Context, cfg config.OAuthConfig, req dto.SocialLoginRequest) (*dto.AuthResponse, error) {
	profile, err := oauth.VerifyAppleIDToken(ctx, req.IDToken, cfg.AppleClientID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrUnauthorized.Code, "invalid apple token", apperrors.ErrUnauthorized.Status)
	}
	email := oauth.NormalizeAppleEmail(profile.Email, profile.Subject)
	name := strings.TrimSpace(req.FullName)
	if name == "" {
		name = strings.Split(email, "@")[0]
	}
	return s.loginOrCreateSocialUser(models.AuthProviderApple, profile.Subject, email, name, req.HotelContext)
}

func (s *AuthService) loginOrCreateSocialUser(provider models.AuthProvider, providerID, email, fullName string, hotelContext *dto.HotelContextInput) (*dto.AuthResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.users.FindByProvider(string(provider), providerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
	}
	created := false
	if user == nil {
		existing, emailErr := s.users.FindByEmail(email)
		if emailErr == nil {
			if existing.AuthProvider == models.AuthProviderEmail && existing.PasswordHash != "" {
				return nil, apperrors.New(apperrors.ErrConflict.Code, "sign in with email and password first", apperrors.ErrConflict.Status)
			}
			user = existing
			user.AuthProvider = provider
			user.ProviderID = providerID
			if user.FullName == "" {
				user.FullName = fullName
			}
			if err := s.users.Update(user); err != nil {
				return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to link social account", apperrors.ErrInternal.Status)
			}
		} else if errors.Is(emailErr, gorm.ErrRecordNotFound) {
			user = &models.User{
				Email:        email,
				FullName:     fullName,
				Role:         models.UserRoleCustomer,
				AuthProvider: provider,
				ProviderID:   providerID,
			}
			if err := s.users.Create(user); err != nil {
				return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create user", apperrors.ErrInternal.Status)
			}
			created = true
		} else {
			return nil, apperrors.Wrap(emailErr, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
		}
	}

	if created && s.guestStays != nil {
		s.guestStays.RecordFromContext(hotelContext, user.ID, models.GuestStaySourceRegister)
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

func (s *AuthService) ensurePasswordLoginAllowed(user *models.User) error {
	if user.PasswordHash == "" && user.AuthProvider != models.AuthProviderEmail {
		return apperrors.New(apperrors.ErrUnauthorized.Code, fmt.Sprintf("sign in with %s instead", user.AuthProvider), apperrors.ErrUnauthorized.Status)
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

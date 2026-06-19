package dto

type RegisterRequest struct {
	Email        string             `json:"email" validate:"required,email"`
	Password     string             `json:"password" validate:"required,min=8"`
	FullName     string             `json:"full_name" validate:"required"`
	Phone        string             `json:"phone"`
	HotelContext *HotelContextInput `json:"hotel_context"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	Role         string `json:"role"`
	AuthProvider string `json:"auth_provider"`
}

type SocialLoginRequest struct {
	IDToken      string             `json:"id_token" validate:"required"`
	FullName     string             `json:"full_name"`
	HotelContext *HotelContextInput `json:"hotel_context"`
}

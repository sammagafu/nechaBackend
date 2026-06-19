package service

import (
	"strings"

	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

var chatCategories = map[string]string{
	"orders":         "Orders",
	"delivery":       "Delivery",
	"products":       "Products",
	"hotel-partners": "Hotel partners",
	"account":        "Account",
	"other":          "Other",
}

func normalizeChatCategory(raw string) (string, error) {
	category := strings.TrimSpace(strings.ToLower(raw))
	if category == "" {
		return "", apperrors.New(apperrors.ErrBadRequest.Code, "category is required", apperrors.ErrBadRequest.Status)
	}
	if _, ok := chatCategories[category]; !ok {
		return "", apperrors.New(apperrors.ErrBadRequest.Code, "invalid category", apperrors.ErrBadRequest.Status)
	}
	return category, nil
}

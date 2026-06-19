package oauth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleProfile struct {
	Subject   string
	Email     string
	FullName  string
	Verified  bool
}

func VerifyGoogleIDToken(ctx context.Context, idToken, clientID string) (*GoogleProfile, error) {
	if clientID == "" {
		return nil, fmt.Errorf("google client id not configured")
	}
	payload, err := idtoken.Validate(ctx, idToken, clientID)
	if err != nil {
		return nil, err
	}
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	return &GoogleProfile{
		Subject:  payload.Subject,
		Email:    email,
		FullName: name,
		Verified: verified,
	}, nil
}

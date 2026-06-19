package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AppleProfile struct {
	Subject string
	Email   string
}

type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var (
	appleKeysMu  sync.Mutex
	appleKeys    map[string]*rsa.PublicKey
	appleKeysAt  time.Time
	appleKeysTTL = 6 * time.Hour
)

func VerifyAppleIDToken(ctx context.Context, idToken, clientID string) (*AppleProfile, error) {
	if clientID == "" {
		return nil, fmt.Errorf("apple client id not configured")
	}
	token, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		kid, _ := t.Header["kid"].(string)
		return applePublicKey(ctx, kid)
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid apple identity token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid apple claims")
	}
	if iss, _ := claims["iss"].(string); iss != "https://appleid.apple.com" {
		return nil, fmt.Errorf("invalid apple issuer")
	}
	aud, _ := claims["aud"].(string)
	if aud != clientID {
		return nil, fmt.Errorf("invalid apple audience")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, fmt.Errorf("missing apple subject")
	}
	email, _ := claims["email"].(string)
	return &AppleProfile{Subject: sub, Email: email}, nil
}

func applePublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	appleKeysMu.Lock()
	defer appleKeysMu.Unlock()
	if appleKeys == nil || time.Since(appleKeysAt) > appleKeysTTL {
		keys, err := fetchAppleKeys(ctx)
		if err != nil {
			return nil, err
		}
		appleKeys = keys
		appleKeysAt = time.Now()
	}
	key := appleKeys[kid]
	if key == nil {
		keys, err := fetchAppleKeys(ctx)
		if err != nil {
			return nil, err
		}
		appleKeys = keys
		appleKeysAt = time.Now()
		key = appleKeys[kid]
	}
	if key == nil {
		return nil, fmt.Errorf("apple public key not found")
	}
	return key, nil
}

func fetchAppleKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var jwks appleJWKS
	if err := json.Unmarshal(raw, &jwks); err != nil {
		return nil, err
	}
	out := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Kid == "" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(key.N, key.E)
		if err != nil {
			continue
		}
		out[key.Kid] = pub
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no apple keys loaded")
	}
	return out, nil
}

func rsaPublicKeyFromJWK(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		e = 65537
	}
	return &rsa.PublicKey{N: n, E: e}, nil
}

func NormalizeAppleEmail(email, subject string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email != "" {
		return email
	}
	return fmt.Sprintf("%s@privaterelay.appleid.com", subject)
}

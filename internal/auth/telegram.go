package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type TelegramClaims struct {
	jwt.RegisteredClaims
	ID                string `json:"id"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
}

type TelegramUser struct {
	TelegramID int64
	Name       string
	Username   string
	PhotoURL   string
}

type Verifier struct {
	kf       keyfunc.Keyfunc
	clientID string
}

func NewVerifier(ctx context.Context, clientID string) (*Verifier, error) {
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{
		"https://oauth.telegram.org/.well-known/jwks.json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	return &Verifier{kf: kf, clientID: clientID}, nil
}

func (v *Verifier) VerifyIDToken(ctx context.Context, idToken string) (*TelegramUser, error) {
	token, err := jwt.ParseWithClaims(idToken, &TelegramClaims{}, v.kf.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("invalid id_token: %w", err)
	}

	claims, ok := token.Claims.(*TelegramClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.Issuer != "https://oauth.telegram.org" {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}
	if len(claims.Audience) == 0 || claims.Audience[0] != v.clientID {
		return nil, fmt.Errorf("invalid audience")
	}

	tgID, _ := strconv.ParseInt(claims.ID, 10, 64)

	return &TelegramUser{
		TelegramID: tgID,
		Name:       claims.Name,
		Username:   claims.PreferredUsername,
		PhotoURL:   claims.Picture,
	}, nil
}

func SplitName(full string) (first, last string) {
	parts := strings.SplitN(full, " ", 2)
	first = parts[0]
	if len(parts) > 1 {
		last = parts[1]
	}
	return
}

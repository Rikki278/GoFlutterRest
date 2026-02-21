package jwt

import (
	"errors"
	"time"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims embedded in the JWT access token.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	gojwt.RegisteredClaims
}

// Service handles JWT generation and validation.
type Service struct {
	cfg *config.JWTConfig
}

func NewService(cfg *config.JWTConfig) *Service {
	return &Service{cfg: cfg}
}

// GenerateAccessToken creates a short-lived access token signed with the access secret.
func (s *Service) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(s.cfg.AccessExpiresDuration)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.AccessSecret))
}

// GenerateRefreshToken creates a long-lived opaque token (random UUID string).
// The actual refresh token stored in DB is just a UUID — simpler and revocable.
func (s *Service) GenerateRefreshToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// ValidateAccessToken parses and validates an access token, returning claims.
func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, apperror.Unauthorized("unexpected signing method")
		}
		return []byte(s.cfg.AccessSecret), nil
	})

	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, apperror.TokenExpired()
		}
		return nil, apperror.Unauthorized("invalid or malformed token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token claims")
	}

	return claims, nil
}

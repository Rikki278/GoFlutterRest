package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/internal/jwt"
	"github.com/acidsoft/gorestteach/internal/repository"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ─── DTOs ────────────────────────────────────────────────────────────────────

type RegisterInput struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// ─── Use Case ────────────────────────────────────────────────────────────────

type AuthUseCase struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.RefreshTokenRepository
	jwtService *jwt.Service
	jwtCfg     *config.JWTConfig
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	jwtService *jwt.Service,
	jwtCfg *config.JWTConfig,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
		jwtCfg:     jwtCfg,
	}
}

// Register creates a new user account.
func (uc *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*domain.UserPublic, error) {
	// Check email uniqueness
	_, err := uc.userRepo.GetByEmail(ctx, strings.ToLower(input.Email))
	if err == nil {
		// user found → conflict
		return nil, apperror.Conflict("Email is already registered")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	user := &domain.User{
		Name:     input.Name,
		Email:    strings.ToLower(input.Email),
		Password: string(hash),
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	pub := user.ToPublic()
	return &pub, nil
}

// Login verifies credentials and returns an access + refresh token pair.
func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	user, err := uc.userRepo.GetByEmail(ctx, strings.ToLower(input.Email))
	if err != nil {
		// Return generic message to prevent email enumeration
		return nil, apperror.Unauthorized("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, apperror.Unauthorized("Invalid email or password")
	}

	return uc.issueTokenPair(ctx, user)
}

// Refresh exchanges a valid refresh token for a new access + refresh token pair.
// Old refresh token is deleted (rotation pattern).
func (uc *AuthUseCase) Refresh(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	storedToken, err := uc.tokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	if storedToken.IsExpired() {
		// Clean up expired token
		_ = uc.tokenRepo.DeleteByToken(ctx, refreshTokenStr)
		return nil, apperror.Unauthorized("Refresh token has expired, please login again")
	}

	user, err := uc.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token (rotation)
	_ = uc.tokenRepo.DeleteByToken(ctx, refreshTokenStr)

	return uc.issueTokenPair(ctx, user)
}

// Logout invalidates the given refresh token.
func (uc *AuthUseCase) Logout(ctx context.Context, refreshTokenStr string) error {
	return uc.tokenRepo.DeleteByToken(ctx, refreshTokenStr)
}

// issueTokenPair is an internal helper that generates both tokens and persists the refresh token.
func (uc *AuthUseCase) issueTokenPair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	refreshTokenStr, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal(err)
	}

	refreshRecord := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(uc.jwtCfg.RefreshExpiresDuration),
	}
	if err := uc.tokenRepo.Save(ctx, refreshRecord); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
	}, nil
}

// GetUserIDFromContext is a helper used by handlers to extract userID safely.
func GetUserID(ctx map[string]any) (uuid.UUID, error) {
	id, ok := ctx["user_id"].(uuid.UUID)
	if !ok {
		return uuid.Nil, apperror.Unauthorized("user identity not found in context")
	}
	return id, nil
}

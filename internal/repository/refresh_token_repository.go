package repository

import (
	"context"
	"errors"

	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, tokenStr string) (*domain.RefreshToken, error)
	DeleteByToken(ctx context.Context, tokenStr string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Save(ctx context.Context, token *domain.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.WithContext(ctx).First(&token, "token = ?", tokenStr).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Unauthorized("refresh token not found or already used")
		}
		return nil, apperror.Internal(err)
	}
	return &token, nil
}

func (r *refreshTokenRepository) DeleteByToken(ctx context.Context, tokenStr string) error {
	return r.db.WithContext(ctx).
		Where("token = ?", tokenStr).
		Delete(&domain.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.RefreshToken{}).Error
}

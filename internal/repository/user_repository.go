package repository

import (
	"context"
	"errors"

	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the persistence contract for users.
// Keeping it as an interface allows easy mocking in tests.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateAvatar(ctx context.Context, userID, avatarID uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User")
		}
		return nil, apperror.Internal(err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User")
		}
		return nil, apperror.Internal(err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *userRepository) UpdateAvatar(ctx context.Context, userID, avatarID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("avatar_id", avatarID).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

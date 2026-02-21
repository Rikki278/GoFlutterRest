package repository

import (
	"context"

	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageRepository interface {
	Save(ctx context.Context, image *domain.Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error)
}

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) Save(ctx context.Context, image *domain.Image) error {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	var img domain.Image
	if err := r.db.WithContext(ctx).First(&img, "id = ?", id).Error; err != nil {
		return nil, apperror.NotFound("Image")
	}
	return &img, nil
}

package repository

import (
	"context"
	"errors"

	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	List(ctx context.Context, page, perPage int, search string) ([]domain.Post, int64, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateImage(ctx context.Context, postID, imageID uuid.UUID) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *postRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	var post domain.Post
	err := r.db.WithContext(ctx).
		Preload("User"). // eager-load author info
		First(&post, "posts.id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Post")
		}
		return nil, apperror.Internal(err)
	}
	return &post, nil
}

// List supports pagination and optional full-text search on title/body.
func (r *postRepository) List(ctx context.Context, page, perPage int, search string) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Post{}).Preload("User")
	if search != "" {
		pattern := "%" + search + "%"
		q = q.Where("title ILIKE ? OR body ILIKE ?", pattern, pattern)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperror.Internal(err)
	}

	offset := (page - 1) * perPage
	if err := q.Offset(offset).Limit(perPage).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, apperror.Internal(err)
	}

	return posts, total, nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	if err := r.db.WithContext(ctx).Save(post).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Post{}, "id = ?", id).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (r *postRepository) UpdateImage(ctx context.Context, postID, imageID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		Update("image_id", imageID).Error; err != nil {
		return apperror.Internal(err)
	}
	return nil
}

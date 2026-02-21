package usecase

import (
	"context"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/internal/repository"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/google/uuid"
)

// ─── DTOs ────────────────────────────────────────────────────────────────────

type CreatePostInput struct {
	Title string `json:"title" validate:"required,min=3,max=255"`
	Body  string `json:"body"  validate:"required,min=10"`
}

type UpdatePostInput struct {
	Title string `json:"title" validate:"omitempty,min=3,max=255"`
	Body  string `json:"body"  validate:"omitempty,min=10"`
}

type ListPostsInput struct {
	Page    int    `form:"page"     validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Search  string `form:"search"`
}

// ─── Use Case ────────────────────────────────────────────────────────────────

type PostUseCase struct {
	postRepo  repository.PostRepository
	imageRepo repository.ImageRepository
	uploadCfg *config.UploadConfig
}

func NewPostUseCase(
	postRepo repository.PostRepository,
	imageRepo repository.ImageRepository,
	uploadCfg *config.UploadConfig,
) *PostUseCase {
	return &PostUseCase{postRepo: postRepo, imageRepo: imageRepo, uploadCfg: uploadCfg}
}

// Create creates a new post owned by userID.
func (uc *PostUseCase) Create(ctx context.Context, userID uuid.UUID, input CreatePostInput) (*domain.Post, error) {
	post := &domain.Post{
		UserID: userID,
		Title:  input.Title,
		Body:   input.Body,
	}
	if err := uc.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

// GetByID returns a single post with author info.
func (uc *PostUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	return uc.postRepo.GetByID(ctx, id)
}

// List returns a paginated list of posts with optional search.
func (uc *PostUseCase) List(ctx context.Context, input ListPostsInput) ([]domain.Post, int64, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = 10
	}
	return uc.postRepo.List(ctx, page, perPage, input.Search)
}

// Update updates a post, enforcing that only the owner can edit it.
func (uc *PostUseCase) Update(ctx context.Context, postID, userID uuid.UUID, input UpdatePostInput) (*domain.Post, error) {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, apperror.Forbidden()
	}

	if input.Title != "" {
		post.Title = input.Title
	}
	if input.Body != "" {
		post.Body = input.Body
	}

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

// Delete deletes a post, enforcing ownership.
func (uc *PostUseCase) Delete(ctx context.Context, postID, userID uuid.UUID) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return apperror.Forbidden()
	}

	return uc.postRepo.Delete(ctx, postID)
}

// AttachImage validates and stores an image blob, then links it to the post.
func (uc *PostUseCase) AttachImage(ctx context.Context, postID, userID uuid.UUID, data []byte, contentType string) (*domain.Post, error) {
	// Verify post exists and caller is the owner
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, apperror.Forbidden()
	}

	if err := validateImageUpload(data, contentType, uc.uploadCfg.MaxSizeMB); err != nil {
		return nil, err
	}

	img := &domain.Image{
		Data:        data,
		ContentType: contentType,
		Size:        int64(len(data)),
	}
	if err := uc.imageRepo.Save(ctx, img); err != nil {
		return nil, err
	}

	if err := uc.postRepo.UpdateImage(ctx, postID, img.ID); err != nil {
		return nil, err
	}

	return uc.postRepo.GetByID(ctx, postID)
}

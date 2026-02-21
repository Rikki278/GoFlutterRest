package usecase

import (
	"context"
	"errors"
	"net/http"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/internal/domain"
	"github.com/acidsoft/gorestteach/internal/repository"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/google/uuid"
)

// ─── DTOs ────────────────────────────────────────────────────────────────────

type UpdateUserInput struct {
	Name string `json:"name" validate:"omitempty,min=2,max=100"`
	Bio  string `json:"bio"  validate:"omitempty,max=500"`
}

// ─── Use Case ────────────────────────────────────────────────────────────────

type UserUseCase struct {
	userRepo  repository.UserRepository
	imageRepo repository.ImageRepository
	uploadCfg *config.UploadConfig
}

func NewUserUseCase(
	userRepo repository.UserRepository,
	imageRepo repository.ImageRepository,
	uploadCfg *config.UploadConfig,
) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, imageRepo: imageRepo, uploadCfg: uploadCfg}
}

// GetProfile returns the full profile of any user by ID.
func (uc *UserUseCase) GetProfile(ctx context.Context, id uuid.UUID) (*domain.UserPublic, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	pub := user.ToPublic()
	return &pub, nil
}

// UpdateProfile updates the authenticated user's name and/or bio.
func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateUserInput) (*domain.UserPublic, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	user.Bio = input.Bio

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	pub := user.ToPublic()
	return &pub, nil
}

// UploadAvatar validates and stores avatar image data as a blob in the DB.
func (uc *UserUseCase) UploadAvatar(ctx context.Context, userID uuid.UUID, data []byte, contentType string) (*domain.UserPublic, error) {
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

	if err := uc.userRepo.UpdateAvatar(ctx, userID, img.ID); err != nil {
		return nil, err
	}

	return uc.GetProfile(ctx, userID)
}

// ─── Shared validation ────────────────────────────────────────────────────────

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func validateImageUpload(data []byte, contentType string, maxMB int64) error {
	_ = errors.New("") // keep import used

	// Validate content type
	if !allowedImageTypes[contentType] {
		return apperror.UnsupportedMedia("Only JPEG, PNG, WebP and GIF images are allowed")
	}

	// Validate size
	maxBytes := maxMB * 1024 * 1024
	if int64(len(data)) > maxBytes {
		return apperror.FileTooLarge(maxMB)
	}

	// Additional sniff-check: validate actual bytes match claimed MIME
	sniffed := http.DetectContentType(data)
	if !allowedImageTypes[sniffed] {
		return &apperror.AppError{
			HTTPStatus: http.StatusUnsupportedMediaType,
			Code:       apperror.ErrUnsupportedMedia,
			Message:    "File content does not match an allowed image type",
		}
	}

	return nil
}

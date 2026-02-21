package handler

import (
	"github.com/acidsoft/gorestteach/internal/repository"
	"github.com/acidsoft/gorestteach/pkg/response"
	"github.com/gin-gonic/gin"
)

// ImageHandler serves image blobs from the database.
type ImageHandler struct {
	imageRepo repository.ImageRepository
}

func NewImageHandler(imageRepo repository.ImageRepository) *ImageHandler {
	return &ImageHandler{imageRepo: imageRepo}
}

// GetImage godoc
// @Summary      Get image by ID
// @Description  Returns the raw image bytes with the correct Content-Type header.
//
//	Use the image_id from user.avatar_id or post.image_id to build this URL.
//
// @Tags         images
// @Produce      image/jpeg
// @Param        id   path  string  true  "Image UUID"
// @Success      200  {file}  binary
// @Failure      404  {object}  map[string]any
// @Router       /images/{id} [get]
func (h *ImageHandler) GetImage(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	img, ucErr := h.imageRepo.GetByID(c.Request.Context(), id)
	if ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	// Serve raw bytes with proper Content-Type — no JSON wrapper needed here
	c.Data(200, img.ContentType, img.Data)
}

// ─── Health check handler ─────────────────────────────────────────────────────

// HealthCheck is a simple liveness probe endpoint.
func HealthCheck(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "ok",
		"version": "1.0.0",
		"service": "gorestteach",
	})
}

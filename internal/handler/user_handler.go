package handler

import (
	"io"
	"net/http"

	"github.com/acidsoft/gorestteach/internal/usecase"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/acidsoft/gorestteach/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile, decoded from the JWT token.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	profile, err := h.userUC.GetProfile(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.OK(c, profile)
}

// GetUser godoc
// @Summary      Get public user profile
// @Description  Returns the public profile of any user by their UUID.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User UUID"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	profile, ucErr := h.userUC.GetProfile(c.Request.Context(), id)
	if ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	response.OK(c, profile)
}

// UpdateMe godoc
// @Summary      Update current user profile
// @Description  Updates the authenticated user's name and/or bio.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      usecase.UpdateUserInput  true  "Profile update payload"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	var input usecase.UpdateUserInput
	if err := bindAndValidate(c, &input); err != nil {
		_ = c.Error(err)
		return
	}

	profile, err := h.userUC.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.OK(c, profile)
}

// UploadAvatar godoc
// @Summary      Upload avatar
// @Description  Uploads a JPEG/PNG/WebP avatar image. Stored as blob in PostgreSQL. Max 5MB.
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar  formData  file  true  "Avatar image file"
// @Success      200     {object}  map[string]any
// @Failure      400     {object}  map[string]any
// @Failure      413     {object}  map[string]any
// @Failure      415     {object}  map[string]any
// @Router       /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		_ = c.Error(apperror.New(http.StatusBadRequest, apperror.ErrBadRequest,
			"Field 'avatar' with image file is required"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		_ = c.Error(apperror.Internal(err))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		_ = c.Error(apperror.Internal(err))
		return
	}

	contentType := http.DetectContentType(data)

	profile, ucErr := h.userUC.UploadAvatar(c.Request.Context(), userID, data, contentType)
	if ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	response.OK(c, profile)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// parseUUID parses a UUID from a Gin path parameter.
func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.New(http.StatusBadRequest, apperror.ErrBadRequest,
			"Invalid UUID format for parameter: "+param)
	}
	return id, nil
}

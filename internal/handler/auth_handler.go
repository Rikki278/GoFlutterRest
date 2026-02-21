package handler

import (
	"io"
	"net/http"

	"github.com/acidsoft/gorestteach/internal/middleware"
	"github.com/acidsoft/gorestteach/internal/usecase"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/acidsoft/gorestteach/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// AuthHandler handles auth-related HTTP requests.
type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account. Returns the created user profile.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      usecase.RegisterInput  true  "Registration payload"
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      409   {object}  map[string]any
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input usecase.RegisterInput
	if err := bindAndValidate(c, &input); err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.authUC.Register(c.Request.Context(), input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Created(c, user)
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and returns access + refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      usecase.LoginInput  true  "Login credentials"
// @Success      200   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input usecase.LoginInput
	if err := bindAndValidate(c, &input); err != nil {
		_ = c.Error(err)
		return
	}

	tokens, err := h.authUC.Login(c.Request.Context(), input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.OK(c, tokens)
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchanges a refresh token for a new access + refresh token pair (rotation).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{refresh_token=string}  true  "Refresh token"
// @Success      200   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := bindAndValidate(c, &body); err != nil {
		_ = c.Error(err)
		return
	}

	tokens, err := h.authUC.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.OK(c, tokens)
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidates the given refresh token. Access token cannot be revoked (use short expiry).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  object{refresh_token=string}  true  "Refresh token to invalidate"
// @Success      204
// @Failure      401  {object}  map[string]any
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := bindAndValidate(c, &body); err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.authUC.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		_ = c.Error(err)
		return
	}

	response.NoContent(c)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// bindAndValidate binds JSON body and runs struct-level validation.
// Returns a typed *apperror.AppError on failure.
func bindAndValidate(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		if err == io.EOF {
			return apperror.New(http.StatusBadRequest, apperror.ErrBadRequest, "Request body is required")
		}
		return apperror.New(http.StatusBadRequest, apperror.ErrBadRequest, "Invalid JSON: "+err.Error())
	}

	if err := validate.Struct(dst); err != nil {
		var fieldErrors []apperror.FieldError
		for _, fe := range err.(validator.ValidationErrors) {
			fieldErrors = append(fieldErrors, apperror.FieldError{
				Field:   fe.Field(),
				Message: validationMessage(fe),
			})
		}
		return apperror.ValidationError(fieldErrors)
	}

	return nil
}

// bindQueryAndValidate binds query parameters and validates.
func bindQueryAndValidate(c *gin.Context, dst any) error {
	if err := c.ShouldBindQuery(dst); err != nil {
		return apperror.New(http.StatusBadRequest, apperror.ErrBadRequest, "Invalid query parameters: "+err.Error())
	}
	if err := validate.Struct(dst); err != nil {
		var fieldErrors []apperror.FieldError
		for _, fe := range err.(validator.ValidationErrors) {
			fieldErrors = append(fieldErrors, apperror.FieldError{
				Field:   fe.Field(),
				Message: validationMessage(fe),
			})
		}
		return apperror.ValidationError(fieldErrors)
	}
	return nil
}

// validationMessage produces a human-readable message for each validator tag.
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + fe.Param() + " characters long"
	case "max":
		return "Must be at most " + fe.Param() + " characters long"
	default:
		return "Invalid value"
	}
}

// mustGetUserID extracts the authenticated user's UUID from context.
// Panics if middleware was skipped (programming error, not user error).
func mustGetUserID(c *gin.Context) interface{} {
	id, exists := c.Get(middleware.ContextUserID)
	if !exists {
		panic("auth middleware not applied: user_id not in context")
	}
	return id
}

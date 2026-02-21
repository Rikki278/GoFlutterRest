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

// PostHandler handles post CRUD and image attachment endpoints.
type PostHandler struct {
	postUC *usecase.PostUseCase
}

func NewPostHandler(postUC *usecase.PostUseCase) *PostHandler {
	return &PostHandler{postUC: postUC}
}

// Create godoc
// @Summary      Create a post
// @Description  Creates a new post owned by the authenticated user.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      usecase.CreatePostInput  true  "Post payload"
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Router       /posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	var input usecase.CreatePostInput
	if err := bindAndValidate(c, &input); err != nil {
		_ = c.Error(err)
		return
	}

	post, err := h.postUC.Create(c.Request.Context(), userID, input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Created(c, post)
}

// List godoc
// @Summary      List posts
// @Description  Returns a paginated list of posts. Supports ?page=1&per_page=10&search=keyword
// @Tags         posts
// @Produce      json
// @Security     BearerAuth
// @Param        page      query   int     false  "Page number (default: 1)"
// @Param        per_page  query   int     false  "Items per page (default: 10, max: 100)"
// @Param        search    query   string  false  "Search keyword in title/body"
// @Success      200  {object}  map[string]any
// @Router       /posts [get]
func (h *PostHandler) List(c *gin.Context) {
	var input usecase.ListPostsInput
	if err := bindQueryAndValidate(c, &input); err != nil {
		_ = c.Error(err)
		return
	}

	posts, total, err := h.postUC.List(c.Request.Context(), input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = 10
	}

	response.OKWithMeta(c, posts, response.PaginationMeta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GetByID godoc
// @Summary      Get post by ID
// @Description  Returns a single post with the author's info.
// @Tags         posts
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Post UUID"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /posts/{id} [get]
func (h *PostHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	post, ucErr := h.postUC.GetByID(c.Request.Context(), id)
	if ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	response.OK(c, post)
}

// Update godoc
// @Summary      Update post
// @Description  Updates a post. Only the post owner can update it.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Post UUID"
// @Param        body  body      usecase.UpdatePostInput  true  "Update payload"
// @Success      200   {object}  map[string]any
// @Failure      403   {object}  map[string]any
// @Failure      404   {object}  map[string]any
// @Router       /posts/{id} [put]
func (h *PostHandler) Update(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	id, err := parseUUID(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	var input usecase.UpdatePostInput
	if err := bindAndValidate(c, &input); err != nil {
		_ = c.Error(err)
		return
	}

	post, ucErr := h.postUC.Update(c.Request.Context(), id, userID, input)
	if ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	response.OK(c, post)
}

// Delete godoc
// @Summary      Delete post
// @Description  Deletes a post. Only the post owner can delete it.
// @Tags         posts
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Post UUID"
// @Success      204
// @Failure      403  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /posts/{id} [delete]
func (h *PostHandler) Delete(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	id, err := parseUUID(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	if ucErr := h.postUC.Delete(c.Request.Context(), id, userID); ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	response.NoContent(c)
}

// AttachImage godoc
// @Summary      Attach image to post
// @Description  Uploads an image and attaches it to the post. Only the owner can attach images.
// @Tags         posts
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string  true  "Post UUID"
// @Param        image  formData  file    true  "Image file (JPEG, PNG, WebP, GIF)"
// @Success      200    {object}  map[string]any
// @Failure      403    {object}  map[string]any
// @Failure      413    {object}  map[string]any
// @Failure      415    {object}  map[string]any
// @Router       /posts/{id}/image [post]
func (h *PostHandler) AttachImage(c *gin.Context) {
	userID := mustGetUserID(c).(uuid.UUID)

	id, err := parseUUID(c, "id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		_ = c.Error(apperror.New(http.StatusBadRequest, apperror.ErrBadRequest,
			"Field 'image' with image file is required"))
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

	post, ucErr := h.postUC.AttachImage(c.Request.Context(), id, userID, data, contentType)
	if ucErr != nil {
		_ = c.Error(ucErr)
		return
	}

	response.OK(c, post)
}

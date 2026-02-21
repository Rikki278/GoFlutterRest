package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─── Response envelope ────────────────────────────────────────────────────────

// Success wraps any payload in a standard envelope.
type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Meta    any  `json:"meta,omitempty"`
}

// errorEnvelope wraps error information in a standard envelope.
type errorEnvelope struct {
	Success bool      `json:"success"`
	Error   errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Pagination metadata for list responses.
type PaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

// ─── Success responses ────────────────────────────────────────────────────────

// OK sends HTTP 200 with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successEnvelope{Success: true, Data: data})
}

// Created sends HTTP 201 with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, successEnvelope{Success: true, Data: data})
}

// OKWithMeta sends HTTP 200 with data and pagination meta.
func OKWithMeta(c *gin.Context, data any, meta PaginationMeta) {
	c.JSON(http.StatusOK, successEnvelope{Success: true, Data: data, Meta: meta})
}

// NoContent sends HTTP 204 (no body).
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ─── Error responses ──────────────────────────────────────────────────────────

// Error sends a structured error JSON response.
func Error(c *gin.Context, status int, code string, message string, details any) {
	c.JSON(status, errorEnvelope{
		Success: false,
		Error: errorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

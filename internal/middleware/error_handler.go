package middleware

import (
	"errors"
	"net/http"

	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/acidsoft/gorestteach/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ErrorHandler is a Gin middleware that catches panics and *apperror.AppError values
// set via c.Error(), translating them into structured JSON responses.
//
// Usage pattern in handlers:
//
//	if err != nil {
//	    _ = c.Error(err)
//	    c.Abort()
//	    return
//	}
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Run handler
		c.Next()

		// Check if any errors were accumulated
		if len(c.Errors) == 0 {
			return
		}

		// Process the last error (most specific)
		err := c.Errors.Last().Err

		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			// Known application error — return its code and message
			if appErr.Cause != nil {
				log.Error().Err(appErr.Cause).Str("code", string(appErr.Code)).Msg("application error")
			}
			response.Error(c, appErr.HTTPStatus, string(appErr.Code), appErr.Message, appErr.Details)
			return
		}

		// Unknown error — log it, return generic 500 (never leak internals)
		log.Error().Err(err).Str("path", c.Request.URL.Path).Msg("unhandled error")
		response.Error(c,
			http.StatusInternalServerError,
			string(apperror.ErrInternal),
			"An unexpected error occurred. Please try again later.",
			nil,
		)
	}
}

// Recovery is a Gin middleware that recovers from panics and converts them to
// structured error responses rather than crashing the server.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Str("path", c.Request.URL.Path).Msg("panic recovered")
				response.Error(c,
					http.StatusInternalServerError,
					string(apperror.ErrInternal),
					"An unexpected error occurred. Please try again later.",
					nil,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}

package middleware

import (
	"strings"

	"github.com/acidsoft/gorestteach/internal/jwt"
	"github.com/acidsoft/gorestteach/pkg/apperror"
	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID is the key used to store the authenticated user's ID in the Gin context.
	ContextUserID = "user_id"
	// ContextUserEmail is the key used to store the authenticated user's email.
	ContextUserEmail = "user_email"
)

// Auth verifies the Bearer JWT access token in the Authorization header.
// On success, it stores user_id and user_email into the Gin context.
func Auth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			_ = c.Error(apperror.Unauthorized("Authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			_ = c.Error(apperror.Unauthorized("Authorization header must be in format: Bearer <token>"))
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateAccessToken(parts[1])
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Store user info into context for downstream handlers
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)

		c.Next()
	}
}

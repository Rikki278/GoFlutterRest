package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken stores issued refresh tokens in the database.
// This allows server-side revocation (logout, password change, etc.).
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:varchar(512);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

// IsExpired returns true if the token is past its expiry time.
func (r *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(r.ExpiresAt)
}

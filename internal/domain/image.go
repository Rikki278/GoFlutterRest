package domain

import (
	"time"

	"github.com/google/uuid"
)

// Image stores binary file data directly in PostgreSQL as bytea.
// Using a separate table keeps the User/Post rows lean.
type Image struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Data        []byte    `gorm:"type:bytea;not null"                            json:"-"`
	ContentType string    `gorm:"type:varchar(50);not null"                      json:"content_type"`
	Size        int64     `gorm:"not null"                                       json:"size"`
	CreatedAt   time.Time `                                                      json:"created_at"`
}

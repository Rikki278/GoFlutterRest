package domain

import (
	"time"

	"github.com/google/uuid"
)

// Post is an article/post entity owned by a User.
type Post struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Title     string     `gorm:"type:varchar(255);not null"                     json:"title"`
	Body      string     `gorm:"type:text;not null"                             json:"body"`
	ImageID   *uuid.UUID `gorm:"type:uuid"                                      json:"image_id,omitempty"`
	User      *User      `gorm:"foreignKey:UserID"                              json:"author,omitempty"`
	CreatedAt time.Time  `                                                      json:"created_at"`
	UpdatedAt time.Time  `                                                      json:"updated_at"`
}

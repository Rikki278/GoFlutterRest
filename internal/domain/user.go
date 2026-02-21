package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the core user entity stored in the database.
type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null"                     json:"name"`
	Email     string     `gorm:"type:varchar(255);uniqueIndex;not null"          json:"email"`
	Password  string     `gorm:"type:varchar(255);not null"                     json:"-"` // never serialized
	Bio       string     `gorm:"type:text"                                       json:"bio"`
	AvatarID  *uuid.UUID `gorm:"type:uuid"                                       json:"avatar_id,omitempty"`
	CreatedAt time.Time  `                                                       json:"created_at"`
	UpdatedAt time.Time  `                                                       json:"updated_at"`
}

// UserPublic is the safe public representation of a user (no password).
type UserPublic struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Bio       string     `json:"bio"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Bio:       u.Bio,
		AvatarID:  u.AvatarID,
		CreatedAt: u.CreatedAt,
	}
}

package auth

import (
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type User struct {
	ID           id.ID     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Class        string    `json:"class,omitempty"`
	Jurusan      string    `json:"jurusan,omitempty"`
	Position     string    `json:"position,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    id.ID     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

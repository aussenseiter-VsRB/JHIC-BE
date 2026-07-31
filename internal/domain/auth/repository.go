package auth

import (
	"context"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	ByID(ctx context.Context, id id.ID) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, token string, userID id.ID, expiresAt int64) error
	ByToken(ctx context.Context, token string) (*Session, error)
	DeleteByUserID(ctx context.Context, userID id.ID) error
}

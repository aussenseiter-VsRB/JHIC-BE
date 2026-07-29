package auth

import "context"

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	ByID(ctx context.Context, id string) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, token, userID string, expiresAt int64) error
	ByToken(ctx context.Context, token string) (*Session, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

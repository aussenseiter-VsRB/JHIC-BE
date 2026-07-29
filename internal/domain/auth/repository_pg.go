package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) *UsersRepositoryPG {
	return &UsersRepositoryPG{pool: pool}
}

func (r *UsersRepositoryPG) Create(ctx context.Context, u *User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, role, avatar_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Role, u.AvatarURL, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UsersRepositoryPG) ByID(ctx context.Context, id string) (*User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users WHERE id = $1`, id,
	)
	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}
	return u, nil
}

func (r *UsersRepositoryPG) ByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users WHERE email = $1`, email,
	)
	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return u, nil
}

type SessionsRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewSessionsRepository(pool *pgxpool.Pool) *SessionsRepositoryPG {
	return &SessionsRepositoryPG{pool: pool}
}

func (r *SessionsRepositoryPG) Create(ctx context.Context, token, userID string, expiresAt int64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, created_at, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		token, userID, time.Now().UTC(), time.Unix(expiresAt, 0).UTC(),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionsRepositoryPG) ByToken(ctx context.Context, token string) (*Session, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT token, user_id, created_at, expires_at
		 FROM sessions WHERE token = $1 AND expires_at > NOW()`, token,
	)
	s := &Session{}
	err := row.Scan(&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session by token: %w", err)
	}
	return s, nil
}

func (r *SessionsRepositoryPG) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}
	return nil
}

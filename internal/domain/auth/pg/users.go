package pg

import (
	"context"
	"fmt"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{pool: pool}
}

func (r *UsersRepository) Create(ctx context.Context, u *auth.User) error {
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

func (r *UsersRepository) ByID(ctx context.Context, id string) (*auth.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users WHERE id = $1`, id,
	)
	u := &auth.User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}
	return u, nil
}

func (r *UsersRepository) ByEmail(ctx context.Context, email string) (*auth.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users WHERE email = $1`, email,
	)
	u := &auth.User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return u, nil
}

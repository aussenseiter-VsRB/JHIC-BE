package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
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
		`INSERT INTO users (id, email, password_hash, name, role, avatar_url, class, jurusan, position, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Role, u.AvatarURL, u.Class, u.Jurusan, u.Position, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UsersRepository) ByID(ctx context.Context, id id.ID) (*auth.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, COALESCE(avatar_url, ''), COALESCE(class, ''), COALESCE(jurusan, ''), COALESCE(position, ''), created_at, updated_at
		 FROM users WHERE id = $1`, id,
	)
	u := &auth.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UsersRepository) ByEmail(ctx context.Context, email string) (*auth.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, COALESCE(avatar_url, ''), COALESCE(class, ''), COALESCE(jurusan, ''), COALESCE(position, ''), created_at, updated_at
		 FROM users WHERE email = $1`, email,
	)
	u := &auth.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner, u *auth.User) error {
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.AvatarURL, &u.Class, &u.Jurusan, &u.Position, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("scan user: %w", err)
	}
	return nil
}

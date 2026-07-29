package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryPG struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryPG {
	return &RepositoryPG{pool: pool}
}

func (r *RepositoryPG) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *RepositoryPG) ByID(ctx context.Context, id string) (*User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users WHERE id = $1`, id,
	)
	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by id: %w", err)
	}
	return u, nil
}

func (r *RepositoryPG) ByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, name, role, COALESCE(avatar_url, ''), created_at, updated_at
		 FROM users WHERE email = $1`, email,
	)
	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return u, nil
}

func (r *RepositoryPG) Update(ctx context.Context, u *User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $2, name = $3, avatar_url = $4, updated_at = $5
		 WHERE id = $1`,
		u.ID, u.Email, u.Name, u.AvatarURL, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *RepositoryPG) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`,
		id, role,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *RepositoryPG) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

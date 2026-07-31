package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const userColumns = `id, email, name, role, COALESCE(avatar_url, ''), COALESCE(class, ''), COALESCE(jurusan, ''), COALESCE(position, ''), created_at, updated_at`

func (r *Repository) List(ctx context.Context) ([]user.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+userColumns+`
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		if err := scanUser(rows, &u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) ByID(ctx context.Context, id string) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id,
	)
	u := &user.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user by id: %w", err)
	}
	return u, nil
}

func (r *Repository) ByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email,
	)
	u := &user.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user by email: %w", err)
	}
	return u, nil
}

func (r *Repository) Create(ctx context.Context, u *user.User, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, role, avatar_url, class, jurusan, position, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		u.ID, u.Email, passwordHash, u.Name, u.Role, u.AvatarURL, u.Class, u.Jurusan, u.Position, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *Repository) FindByPosition(ctx context.Context, position, class, jurusan string) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users
		 WHERE position = $1 AND ($2 = '' OR class = $2) AND ($3 = '' OR jurusan = $3)
		 LIMIT 1`,
		position, class, jurusan,
	)
	u := &user.User{}
	if err := scanUser(row, u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user by position: %w", err)
	}
	return u, nil
}

func (r *Repository) Update(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $2, name = $3, avatar_url = $4, class = $5, jurusan = $6, position = $7, updated_at = $8
		 WHERE id = $1`,
		u.ID, u.Email, u.Name, u.AvatarURL, u.Class, u.Jurusan, u.Position, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *Repository) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`,
		id, role,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner, u *user.User) error {
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AvatarURL, &u.Class, &u.Jurusan, &u.Position, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("scan user: %w", err)
	}
	return nil
}

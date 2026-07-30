package berita

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

func (r *RepositoryPG) List(ctx context.Context) ([]Berita, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, author_id, title, content, COALESCE(image_url, ''), created_at, updated_at
		 FROM berita ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list berita: %w", err)
	}
	defer rows.Close()

	var list []Berita
	for rows.Next() {
		var b Berita
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.Title, &b.Content, &b.ImageURL, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan berita: %w", err)
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (r *RepositoryPG) ByID(ctx context.Context, id string) (*Berita, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, author_id, title, content, COALESCE(image_url, ''), created_at, updated_at
		 FROM berita WHERE id = $1`, id,
	)
	b := &Berita{}
	err := row.Scan(&b.ID, &b.AuthorID, &b.Title, &b.Content, &b.ImageURL, &b.CreatedAt, &b.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("berita by id: %w", err)
	}
	return b, nil
}

func (r *RepositoryPG) Create(ctx context.Context, b *Berita) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO berita (id, author_id, title, content, image_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		b.ID, b.AuthorID, b.Title, b.Content, b.ImageURL, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create berita: %w", err)
	}
	return nil
}

func (r *RepositoryPG) Update(ctx context.Context, b *Berita) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE berita SET title = $2, content = $3, image_url = $4, updated_at = $5
		 WHERE id = $1`,
		b.ID, b.Title, b.Content, b.ImageURL, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update berita: %w", err)
	}
	return nil
}

func (r *RepositoryPG) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM berita WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete berita: %w", err)
	}
	return nil
}

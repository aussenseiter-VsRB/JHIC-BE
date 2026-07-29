package workspace

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

func (r *RepositoryPG) Create(ctx context.Context, w *Workspace) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO workspaces (id, name, owner_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		w.ID, w.Name, w.OwnerID, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	return nil
}

func (r *RepositoryPG) ByID(ctx context.Context, id string) (*Workspace, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, owner_id, created_at, updated_at
		 FROM workspaces WHERE id = $1`, id,
	)
	w := &Workspace{}
	err := row.Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("workspace by id: %w", err)
	}
	return w, nil
}

func (r *RepositoryPG) ListByOwner(ctx context.Context, ownerID string) ([]Workspace, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, owner_id, created_at, updated_at
		 FROM workspaces WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan workspace: %w", err)
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, rows.Err()
}

func (r *RepositoryPG) Update(ctx context.Context, w *Workspace) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE workspaces SET name = $2, updated_at = $3 WHERE id = $1`,
		w.ID, w.Name, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}
	return nil
}

func (r *RepositoryPG) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	return nil
}

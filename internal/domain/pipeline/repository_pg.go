package pipeline

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

func (r *RepositoryPG) Create(ctx context.Context, p *Pipeline) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO pipelines (id, workspace_id, name, description, status, config, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.WorkspaceID, p.Name, p.Description, p.Status, p.Config, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create pipeline: %w", err)
	}
	return nil
}

func (r *RepositoryPG) ByID(ctx context.Context, id string) (*Pipeline, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, COALESCE(description, ''), status, config, created_at, updated_at
		 FROM pipelines WHERE id = $1`, id,
	)
	p := &Pipeline{}
	err := row.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.Config, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pipeline by id: %w", err)
	}
	return p, nil
}

func (r *RepositoryPG) ListByWorkspace(ctx context.Context, workspaceID string) ([]Pipeline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, workspace_id, name, COALESCE(description, ''), status, config, created_at, updated_at
		 FROM pipelines WHERE workspace_id = $1 ORDER BY created_at DESC`, workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pipelines: %w", err)
	}
	defer rows.Close()

	var pipelines []Pipeline
	for rows.Next() {
		var p Pipeline
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.Config, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pipeline: %w", err)
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, rows.Err()
}

func (r *RepositoryPG) Update(ctx context.Context, p *Pipeline) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE pipelines SET name = $2, description = $3, status = $4, config = $5, updated_at = $6
		 WHERE id = $1`,
		p.ID, p.Name, p.Description, p.Status, p.Config, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update pipeline: %w", err)
	}
	return nil
}

func (r *RepositoryPG) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM pipelines WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete pipeline: %w", err)
	}
	return nil
}

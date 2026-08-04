package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/analytics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Record(ctx context.Context, e analytics.Event) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO analytics_events (event_name, session_hash, user_id, properties, created_at) VALUES ($1,$2,$3,$4,$5)`, e.Name, e.SessionID, e.UserID, analytics.Properties(e.Properties), e.CreatedAt)
	return err
}

func (r *Repository) Summary(ctx context.Context, since time.Time, prefix string) ([]analytics.Summary, error) {
	query := `SELECT event_name, COUNT(*) FROM analytics_events WHERE created_at >= $1`
	if prefix == "" {
		query += ` AND event_name NOT LIKE 'berita.%'`
	} else {
		query += ` AND event_name LIKE $2`
	}
	query += ` GROUP BY event_name ORDER BY COUNT(*) DESC`
	args := []any{since}
	if prefix != "" {
		args = append(args, prefix+"%")
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("analytics summary: %w", err)
	}
	defer rows.Close()
	var out []analytics.Summary
	for rows.Next() {
		var s analytics.Summary
		if err := rows.Scan(&s.Name, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

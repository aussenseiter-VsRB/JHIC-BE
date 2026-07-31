package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRepository struct {
	pool *pgxpool.Pool
}

func NewSessionsRepository(pool *pgxpool.Pool) *SessionsRepository {
	return &SessionsRepository{pool: pool}
}

func (r *SessionsRepository) Create(ctx context.Context, token string, userID id.ID, expiresAt int64) error {
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

func (r *SessionsRepository) ByToken(ctx context.Context, token string) (*auth.Session, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT token, user_id, created_at, expires_at
		 FROM sessions WHERE token = $1 AND expires_at > NOW()`, token,
	)
	s := &auth.Session{}
	err := row.Scan(&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session by token: %w", err)
	}
	return s, nil
}

func (r *SessionsRepository) DeleteByUserID(ctx context.Context, userID id.ID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}
	return nil
}

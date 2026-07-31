package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const requestColumns = `id, requester_id, company, location, start_date, end_date, description, status, COALESCE(cancel_reason, ''), current_step, created_at, updated_at`

const requestColumnsPrefixed = `r.id, r.requester_id, r.company, r.location, r.start_date, r.end_date, r.description, r.status, COALESCE(r.cancel_reason, ''), r.current_step, r.created_at, r.updated_at`

const stepColumns = `id, pkl_request_id, position, approver_id, status, COALESCE(note, ''), sequence, decided_at, created_at, updated_at`

func (r *Repository) CreateRequest(ctx context.Context, req *pkl.PklRequest, steps []pkl.Step) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create request: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO pkl_requests (id, requester_id, company, location, start_date, end_date, description, status, cancel_reason, current_step, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		req.ID, req.RequesterID, req.Company, req.Location, req.StartDate, req.EndDate, req.Description, req.Status, req.CancelReason, req.CurrentStep, req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert pkl request: %w", err)
	}

	for _, s := range steps {
		_, err = tx.Exec(ctx,
			`INSERT INTO pkl_approval_steps (id, pkl_request_id, position, approver_id, status, note, sequence, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			s.ID, s.RequestID, s.Position, s.ApproverID, s.Status, s.Note, s.Sequence, s.CreatedAt, s.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert pkl approval step: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create request: %w", err)
	}
	return nil
}

func (r *Repository) ByID(ctx context.Context, id string) (*pkl.PklRequest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+requestColumns+` FROM pkl_requests WHERE id = $1`, id,
	)
	req := &pkl.PklRequest{}
	if err := scanRequest(row, req); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return req, nil
}

func (r *Repository) ListByRequester(ctx context.Context, requesterID string) ([]pkl.PklRequest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+requestColumns+` FROM pkl_requests WHERE requester_id = $1 ORDER BY created_at DESC`, requesterID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pkl requests by requester: %w", err)
	}
	defer rows.Close()

	var list []pkl.PklRequest
	for rows.Next() {
		var req pkl.PklRequest
		if err := scanRequest(rows, &req); err != nil {
			return nil, err
		}
		list = append(list, req)
	}
	return list, rows.Err()
}

func (r *Repository) ListForApprover(ctx context.Context, approverID string) ([]pkl.PklRequest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT `+requestColumnsPrefixed+`
		 FROM pkl_requests r
		 JOIN pkl_approval_steps s ON s.pkl_request_id = r.id
		 WHERE s.approver_id = $1
		 ORDER BY r.created_at DESC`, approverID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pkl requests for approver: %w", err)
	}
	defer rows.Close()

	var list []pkl.PklRequest
	for rows.Next() {
		var req pkl.PklRequest
		if err := scanRequest(rows, &req); err != nil {
			return nil, err
		}
		list = append(list, req)
	}
	return list, rows.Err()
}

func (r *Repository) ListAll(ctx context.Context) ([]pkl.PklRequest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+requestColumns+` FROM pkl_requests ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all pkl requests: %w", err)
	}
	defer rows.Close()

	var list []pkl.PklRequest
	for rows.Next() {
		var req pkl.PklRequest
		if err := scanRequest(rows, &req); err != nil {
			return nil, err
		}
		list = append(list, req)
	}
	return list, rows.Err()
}

func (r *Repository) StepsByRequest(ctx context.Context, requestID string) ([]pkl.Step, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+stepColumns+` FROM pkl_approval_steps WHERE pkl_request_id = $1 ORDER BY sequence ASC`, requestID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pkl approval steps: %w", err)
	}
	defer rows.Close()

	var list []pkl.Step
	for rows.Next() {
		var s pkl.Step
		if err := scanStep(rows, &s); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *Repository) Decide(ctx context.Context, req *pkl.PklRequest, step *pkl.Step, expectedReqStatus, expectedStepStatus string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin decide: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE pkl_approval_steps SET status = $2, note = $3, decided_at = $4, updated_at = NOW()
		 WHERE id = $1 AND pkl_request_id = $5 AND status = $6`,
		step.ID, step.Status, step.Note, step.DecidedAt, step.RequestID, expectedStepStatus,
	)
	if err != nil {
		return fmt.Errorf("update pkl approval step: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("approval step already decided")
	}

	tag, err = tx.Exec(ctx,
		`UPDATE pkl_requests SET status = $2, current_step = $3, updated_at = NOW()
		 WHERE id = $1 AND status = $4`,
		req.ID, req.Status, req.CurrentStep, expectedReqStatus,
	)
	if err != nil {
		return fmt.Errorf("update pkl request: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("pkl request status changed")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit decide: %w", err)
	}
	return nil
}

func (r *Repository) Cancel(ctx context.Context, req *pkl.PklRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin cancel: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE pkl_requests SET status = $2, cancel_reason = $3, updated_at = NOW()
		 WHERE id = $1 AND status IN ('pending', 'needs_further_action')`,
		req.ID, req.Status, req.CancelReason,
	)
	if err != nil {
		return fmt.Errorf("cancel pkl request: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("request cannot be cancelled in its current state")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit cancel: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRequest(row rowScanner, req *pkl.PklRequest) error {
	err := row.Scan(&req.ID, &req.RequesterID, &req.Company, &req.Location, &req.StartDate, &req.EndDate, &req.Description, &req.Status, &req.CancelReason, &req.CurrentStep, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("scan pkl request: %w", err)
	}
	return nil
}

func scanStep(row rowScanner, s *pkl.Step) error {
	err := row.Scan(&s.ID, &s.RequestID, &s.Position, &s.ApproverID, &s.Status, &s.Note, &s.Sequence, &s.DecidedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("scan pkl approval step: %w", err)
	}
	return nil
}

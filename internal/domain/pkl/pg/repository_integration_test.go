//go:build integration

package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/require"
)

func seedRequest(t *testing.T, requesterID id.ID) (*pkl.PklRequest, []pkl.Step) {
	t.Helper()
	now := time.Now().UTC()

	for i, pos := range positions {
		seedGuru(t, id.ID(100+i), pos+"@example.com", pos)
	}

	req := &pkl.PklRequest{
		ID:          id.New(),
		RequesterID: requesterID,
		Company:     "PT Maju",
		Location:    "Jl. Merdeka 1",
		StartDate:   time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
		Description: "PKL di perusahaan",
		Status:      pkl.StatusPending,
		CurrentStep: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	steps := make([]pkl.Step, 0, len(positions))
	for i, pos := range positions {
		steps = append(steps, pkl.Step{
			ID:         id.New(),
			RequestID:  req.ID,
			Position:   pos,
			ApproverID: id.ID(100 + i),
			Status:     pkl.StepPending,
			Sequence:   i + 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	return req, steps
}

func TestRepository_CreateRequest(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")

	req, steps := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req, steps))

	got, err := repo.ByID(ctx, req.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "PT Maju", got.Company)
	require.Equal(t, pkl.StatusPending, got.Status)
	require.Equal(t, 1, got.CurrentStep)

	storedSteps, err := repo.StepsByRequest(ctx, req.ID)
	require.NoError(t, err)
	require.Len(t, storedSteps, 4)
	for i, s := range storedSteps {
		require.Equal(t, i+1, s.Sequence)
		require.Equal(t, positions[i], s.Position)
		require.Equal(t, id.ID(100+i), s.ApproverID)
		require.Equal(t, pkl.StepPending, s.Status)
	}
}

func TestRepository_ListByRequester(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")
	req1, steps1 := seedRequest(t, id.ID(10))
	req2, steps2 := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req1, steps1))
	require.NoError(t, repo.CreateRequest(ctx, req2, steps2))

	list, err := repo.ListByRequester(ctx, id.ID(10))
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestRepository_ListForApprover(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")

	req1, steps1 := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req1, steps1))
	req2, steps2 := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req2, steps2))

	list, err := repo.ListForApprover(ctx, id.ID(101))
	require.NoError(t, err)
	require.Len(t, list, 2)

	unrelated, err := repo.ListForApprover(ctx, id.ID(199))
	require.NoError(t, err)
	require.Len(t, unrelated, 0)
}

func TestRepository_Decide(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")
	req, steps := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req, steps))

	now := time.Now().UTC()
	req.CurrentStep = 2
	req.UpdatedAt = now
	steps[0].Status = pkl.StepApproved
	steps[0].DecidedAt = &now
	steps[0].UpdatedAt = now

	require.NoError(t, repo.Decide(ctx, req, &steps[0], pkl.StatusPending, pkl.StepPending))

	gotReq, err := repo.ByID(ctx, req.ID)
	require.NoError(t, err)
	require.Equal(t, 2, gotReq.CurrentStep)
	require.Equal(t, pkl.StatusPending, gotReq.Status)

	gotSteps, err := repo.StepsByRequest(ctx, req.ID)
	require.NoError(t, err)
	require.Equal(t, pkl.StepApproved, gotSteps[0].Status)
	require.NotNil(t, gotSteps[0].DecidedAt)
}

func TestRepository_Decide_Conflict(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")
	req, steps := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req, steps))

	now := time.Now().UTC()
	steps[0].Status = pkl.StepApproved
	steps[0].DecidedAt = &now
	steps[0].UpdatedAt = now
	require.NoError(t, repo.Decide(ctx, req, &steps[0], pkl.StatusPending, pkl.StepPending))

	second := *req
	second.UpdatedAt = time.Now().UTC()
	err := repo.Decide(ctx, &second, &steps[0], pkl.StatusPending, pkl.StepPending)
	require.EqualError(t, err, "approval step already decided")
}

func TestRepository_Cancel(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")
	req, steps := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req, steps))

	req.Status = pkl.StatusCancelled
	req.CancelReason = "salah perusahaan"
	req.UpdatedAt = time.Now().UTC()
	require.NoError(t, repo.Cancel(ctx, req))

	got, err := repo.ByID(ctx, req.ID)
	require.NoError(t, err)
	require.Equal(t, pkl.StatusCancelled, got.Status)
	require.Equal(t, "salah perusahaan", got.CancelReason)
}

func TestRepository_Cancel_Conflict(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")
	req, steps := seedRequest(t, id.ID(10))
	require.NoError(t, repo.CreateRequest(ctx, req, steps))

	req.Status = pkl.StatusCancelled
	req.CancelReason = "alasan"
	req.UpdatedAt = time.Now().UTC()
	require.NoError(t, repo.Cancel(ctx, req))

	err := repo.Cancel(ctx, req)
	require.EqualError(t, err, "request cannot be cancelled in its current state")
}

func TestRepository_NoConnectionLeak(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	before := pool.Stat().AcquiredConns()

	seedUser(t, id.ID(10), "student@example.com", "PPLG 1", "PPLG")
	repo := pg.NewRepository(pool)
	for i := 0; i < 5; i++ {
		req, steps := seedRequest(t, id.ID(10))
		require.NoError(t, repo.CreateRequest(ctx, req, steps))
		_, err := repo.ByID(ctx, req.ID)
		require.NoError(t, err)
		_, err = repo.StepsByRequest(ctx, req.ID)
		require.NoError(t, err)
	}

	require.Equal(t, before, pool.Stat().AcquiredConns())
}

var positions = []string{"wali_kelas", "bk", "kesiswaan", "kaprog"}

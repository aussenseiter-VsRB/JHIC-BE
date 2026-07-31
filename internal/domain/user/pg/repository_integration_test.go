//go:build integration

package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user/pg"
	"github.com/stretchr/testify/require"
)

func TestRepository_List(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, "user-1", "a@example.com", "A", "user")
	seedUser(t, "user-2", "b@example.com", "B", "admin")

	list, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 2)

	var ids []string
	for _, u := range list {
		ids = append(ids, u.ID)
	}
	require.Contains(t, ids, "user-1")
	require.Contains(t, ids, "user-2")
}

func TestRepository_ByID(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, "user-1", "a@example.com", "A", "user")

	got, err := repo.ByID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "a@example.com", got.Email)
	require.Equal(t, "A", got.Name)
	require.Equal(t, "user", got.Role)

	missing, err := repo.ByID(ctx, "missing")
	require.NoError(t, err)
	require.Nil(t, missing)
}

func TestRepository_ByEmail(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, "user-1", "a@example.com", "A", "user")

	got, err := repo.ByEmail(ctx, "a@example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "user-1", got.ID)

	missing, err := repo.ByEmail(ctx, "nobody@example.com")
	require.NoError(t, err)
	require.Nil(t, missing)
}

func TestRepository_Update(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, "user-1", "a@example.com", "A", "user")

	u, err := repo.ByID(ctx, "user-1")
	require.NoError(t, err)
	u.Name = "Renamed"
	u.AvatarURL = "https://cdn.example.com/new.png"
	u.UpdatedAt = time.Now().UTC()
	require.NoError(t, repo.Update(ctx, u))

	got, err := repo.ByID(ctx, "user-1")
	require.NoError(t, err)
	require.Equal(t, "Renamed", got.Name)
	require.Equal(t, "https://cdn.example.com/new.png", got.AvatarURL)
	require.False(t, got.UpdatedAt.IsZero())
}

func TestRepository_UpdateRole(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, "user-1", "a@example.com", "A", "user")

	require.NoError(t, repo.UpdateRole(ctx, "user-1", "jurnal"))

	got, err := repo.ByID(ctx, "user-1")
	require.NoError(t, err)
	require.Equal(t, "jurnal", got.Role)
}

func TestRepository_Delete(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, "user-1", "a@example.com", "A", "user")

	require.NoError(t, repo.Delete(ctx, "user-1"))

	got, err := repo.ByID(ctx, "user-1")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestRepository_NoConnectionLeak(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	before := pool.Stat().AcquiredConns()

	repo := pg.NewRepository(pool)
	for i := 0; i < 10; i++ {
		_, err := repo.List(ctx)
		require.NoError(t, err)
		_, err = repo.ByID(ctx, "missing")
		require.NoError(t, err)
	}

	require.Equal(t, before, pool.Stat().AcquiredConns())
}

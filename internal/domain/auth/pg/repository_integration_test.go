//go:build integration

package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/require"
)

func seedUser(t *testing.T, id id.ID, email string) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		`INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		int64(id), email, "hash", "Seed", "user",
	)
	require.NoError(t, err)
}

func TestUsersRepository_CreateAndByID(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewUsersRepository(pool)

	now := time.Now().UTC()
	u := &auth.User{
		ID:           id.ID(1),
		Email:        "a@example.com",
		PasswordHash: "hash",
		Name:         "A",
		Role:         "user",
		AvatarURL:    "https://cdn.example.com/a.png",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.ByID(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, u.Email, got.Email)
	require.Equal(t, u.PasswordHash, got.PasswordHash)
	require.Equal(t, u.Name, got.Name)
	require.Equal(t, u.Role, got.Role)
	require.Equal(t, u.AvatarURL, got.AvatarURL)
	require.False(t, got.CreatedAt.IsZero())
	require.False(t, got.UpdatedAt.IsZero())
}

func TestUsersRepository_ByID_NotFound(t *testing.T) {
	pool := startPostgres(t)
	repo := pg.NewUsersRepository(pool)

	got, err := repo.ByID(context.Background(), id.ID(99))
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestUsersRepository_ByEmail(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewUsersRepository(pool)

	seedUser(t, id.ID(1), "a@example.com")

	got, err := repo.ByEmail(ctx, "a@example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, id.ID(1), got.ID)

	missing, err := repo.ByEmail(ctx, "nobody@example.com")
	require.NoError(t, err)
	require.Nil(t, missing)
}

func TestSessionsRepository_CreateAndByToken(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewSessionsRepository(pool)

	seedUser(t, id.ID(1), "a@example.com")

	require.NoError(t, repo.Create(ctx, "token-1", id.ID(1), time.Now().Add(time.Hour).Unix()))

	s, err := repo.ByToken(ctx, "token-1")
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, id.ID(1), s.UserID)
	require.False(t, s.CreatedAt.IsZero())
	require.False(t, s.ExpiresAt.IsZero())
}

func TestSessionsRepository_ByToken_Expired(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewSessionsRepository(pool)

	seedUser(t, id.ID(1), "a@example.com")

	require.NoError(t, repo.Create(ctx, "expired-token", id.ID(1), -1))

	s, err := repo.ByToken(ctx, "expired-token")
	require.NoError(t, err)
	require.Nil(t, s)
}

func TestSessionsRepository_DeleteByUserID(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewSessionsRepository(pool)

	seedUser(t, id.ID(1), "a@example.com")
	require.NoError(t, repo.Create(ctx, "token-1", id.ID(1), time.Now().Add(time.Hour).Unix()))
	require.NoError(t, repo.Create(ctx, "token-2", id.ID(1), time.Now().Add(time.Hour).Unix()))

	require.NoError(t, repo.DeleteByUserID(ctx, id.ID(1)))

	for _, tok := range []string{"token-1", "token-2"} {
		s, err := repo.ByToken(ctx, tok)
		require.NoError(t, err)
		require.Nil(t, s)
	}
}

func TestAuthRepositories_NoConnectionLeak(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	before := pool.Stat().AcquiredConns()

	users := pg.NewUsersRepository(pool)
	sessions := pg.NewSessionsRepository(pool)
	for i := 0; i < 10; i++ {
		_, err := users.ByID(ctx, id.ID(99))
		require.NoError(t, err)
		_, err = users.ByEmail(ctx, "nobody@example.com")
		require.NoError(t, err)
		_, err = sessions.ByToken(ctx, "missing")
		require.NoError(t, err)
	}

	require.Equal(t, before, pool.Stat().AcquiredConns())
}

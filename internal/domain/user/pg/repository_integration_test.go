//go:build integration

package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/require"
)

func TestRepository_List(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(1), "a@example.com", "A", "user")
	seedUser(t, id.ID(2), "b@example.com", "B", "admin")

	list, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 2)

	var ids []id.ID
	for _, u := range list {
		ids = append(ids, u.ID)
	}
	require.Contains(t, ids, id.ID(1))
	require.Contains(t, ids, id.ID(2))
}

func TestRepository_ByID(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(1), "a@example.com", "A", "user")

	got, err := repo.ByID(ctx, id.ID(1))
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "a@example.com", got.Email)
	require.Equal(t, "A", got.Name)
	require.Equal(t, "user", got.Role)

	missing, err := repo.ByID(ctx, id.ID(99))
	require.NoError(t, err)
	require.Nil(t, missing)
}

func TestRepository_ByEmail(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(1), "a@example.com", "A", "user")

	got, err := repo.ByEmail(ctx, "a@example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, id.ID(1), got.ID)

	missing, err := repo.ByEmail(ctx, "nobody@example.com")
	require.NoError(t, err)
	require.Nil(t, missing)
}

func TestRepository_Update(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(1), "a@example.com", "A", "user")

	u, err := repo.ByID(ctx, id.ID(1))
	require.NoError(t, err)
	u.Name = "Renamed"
	u.AvatarURL = "https://cdn.example.com/new.png"
	u.Class = "PPLG 1"
	u.Jurusan = "PPLG"
	u.UpdatedAt = time.Now().UTC()
	require.NoError(t, repo.Update(ctx, u))

	got, err := repo.ByID(ctx, id.ID(1))
	require.NoError(t, err)
	require.Equal(t, "Renamed", got.Name)
	require.Equal(t, "https://cdn.example.com/new.png", got.AvatarURL)
	require.Equal(t, "PPLG 1", got.Class)
	require.Equal(t, "PPLG", got.Jurusan)
	require.False(t, got.UpdatedAt.IsZero())
}

func TestRepository_Create(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	u := &user.User{
		ID: id.ID(1), Email: "a@example.com", Name: "A", Role: "guru",
		Class: "PPLG 1", Jurusan: "PPLG", Position: "wali_kelas",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, u, "hashed-password"))

	got, err := repo.ByID(ctx, id.ID(1))
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "wali_kelas", got.Position)
	require.Equal(t, "PPLG 1", got.Class)
	require.Equal(t, "PPLG", got.Jurusan)

	var hash string
	err = pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, id.ID(1)).Scan(&hash)
	require.NoError(t, err)
	require.Equal(t, "hashed-password", hash)
}

func TestRepository_FindByPosition(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedGuru(t, id.ID(10), "wk1@example.com", "wali_kelas", "PPLG 1", "PPLG")
	seedGuru(t, id.ID(11), "wk2@example.com", "wali_kelas", "PPLG 2", "PPLG")
	seedGuru(t, id.ID(12), "bk@example.com", "bk", "", "")
	seedGuru(t, id.ID(13), "kesiswaan@example.com", "kesiswaan", "", "")
	seedGuru(t, id.ID(14), "kaprog@example.com", "kaprog", "", "PPLG")
	seedGuru(t, id.ID(15), "kaprog-ak@example.com", "kaprog", "", "AK")

	t.Run("wali kelas by class", func(t *testing.T) {
		got, err := repo.FindByPosition(ctx, "wali_kelas", "PPLG 1", "")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, id.ID(10), got.ID)
	})

	t.Run("kaprog by jurusan", func(t *testing.T) {
		got, err := repo.FindByPosition(ctx, "kaprog", "", "AK")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, id.ID(15), got.ID)
	})

	t.Run("bk school wide", func(t *testing.T) {
		got, err := repo.FindByPosition(ctx, "bk", "", "")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, id.ID(12), got.ID)
	})

	t.Run("kesiswaan school wide", func(t *testing.T) {
		got, err := repo.FindByPosition(ctx, "kesiswaan", "", "")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, id.ID(13), got.ID)
	})

	t.Run("no match", func(t *testing.T) {
		got, err := repo.FindByPosition(ctx, "kaprog", "", "HTL")
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

func TestRepository_UpdateRole(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(1), "a@example.com", "A", "user")

	require.NoError(t, repo.UpdateRole(ctx, id.ID(1), "jurnal"))

	got, err := repo.ByID(ctx, id.ID(1))
	require.NoError(t, err)
	require.Equal(t, "jurnal", got.Role)
}

func TestRepository_Delete(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	repo := pg.NewRepository(pool)

	seedUser(t, id.ID(1), "a@example.com", "A", "user")

	require.NoError(t, repo.Delete(ctx, id.ID(1)))

	got, err := repo.ByID(ctx, id.ID(1))
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
		_, err = repo.ByID(ctx, id.ID(99))
		require.NoError(t, err)
	}

	require.Equal(t, before, pool.Stat().AcquiredConns())
}

//go:build integration

package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb/pg"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateAndByID(t *testing.T) {
	pool := startPostgres(t)
	repo := pg.NewRepository(pool)

	reg := newRegistration()
	require.NoError(t, repo.Create(context.Background(), reg))

	got, err := repo.ByID(context.Background(), reg.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, reg.ID, got.ID)
	require.Equal(t, "Budi Santoso", got.Nama)
	require.Equal(t, spmb.StatusProses, got.Status)
}

func TestRepository_ByIDNotFound(t *testing.T) {
	pool := startPostgres(t)
	repo := pg.NewRepository(pool)

	got, err := repo.ByID(context.Background(), id.New())
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestRepository_ListAllOrderByCreatedDesc(t *testing.T) {
	pool := startPostgres(t)
	repo := pg.NewRepository(pool)

	first := newRegistration()
	first.CreatedAt = time.Now().UTC().Add(-time.Hour)
	second := newRegistration()
	second.Nama = "Citra Lestari"
	second.CreatedAt = time.Now().UTC()

	require.NoError(t, repo.Create(context.Background(), first))
	require.NoError(t, repo.Create(context.Background(), second))

	list, err := repo.ListAll(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, "Citra Lestari", list[0].Nama)
	require.Equal(t, "Budi Santoso", list[1].Nama)
}

func TestRepository_UpdateStatusApprove(t *testing.T) {
	pool := startPostgres(t)
	repo := pg.NewRepository(pool)

	reg := newRegistration()
	require.NoError(t, repo.Create(context.Background(), reg))

	reg.Status = spmb.StatusApprove
	reg.UpdatedAt = time.Now().UTC()
	require.NoError(t, repo.UpdateStatus(context.Background(), reg, spmb.StatusProses))

	got, err := repo.ByID(context.Background(), reg.ID)
	require.NoError(t, err)
	require.Equal(t, spmb.StatusApprove, got.Status)
}

func TestRepository_UpdateStatusStaleExpectedFails(t *testing.T) {
	pool := startPostgres(t)
	repo := pg.NewRepository(pool)

	reg := newRegistration()
	require.NoError(t, repo.Create(context.Background(), reg))

	reg.Status = spmb.StatusApprove
	require.Error(t, repo.UpdateStatus(context.Background(), reg, "wrong-status"))
}

func newRegistration() *spmb.SpmbRegistration {
	now := time.Now().UTC()
	return &spmb.SpmbRegistration{
		ID:           id.New(),
		Nama:         "Budi Santoso",
		Nik:          "3204123456789012",
		Nisn:         "0123456789",
		KkNo:         "3204012312310001",
		TempatLahir:  "Bandung",
		TanggalLahir: "2010-05-10",
		JenisKelamin: "Laki-laki",
		Agama:        "Islam",
		Alamat:       "Jl. Raya Soreang No. 1",
		AsalSekolah:  "SMPN 1 Soreang",
		NoHP:         "081234567890",
		NamaAyah:     "Ahmad Santoso",
		NamaIbu:      "Siti Aminah",
		Jurusan:      "PPLG",
		Status:       spmb.StatusProses,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

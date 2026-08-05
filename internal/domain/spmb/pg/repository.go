package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const columns = `id, nama, nik, COALESCE(nisn, ''), COALESCE(kk_no, ''), COALESCE(tempat_lahir, ''), COALESCE(tanggal_lahir, ''), COALESCE(jenis_kelamin, ''), COALESCE(agama, ''), COALESCE(alamat, ''), COALESCE(asal_sekolah, ''), COALESCE(no_hp, ''), COALESCE(nama_ayah, ''), COALESCE(nama_ibu, ''), COALESCE(jurusan, ''), status, COALESCE(cancel_reason, ''), created_at, updated_at`

func (r *Repository) Create(ctx context.Context, reg *spmb.SpmbRegistration) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO spmb_registrations (id, nama, nik, nisn, kk_no, tempat_lahir, tanggal_lahir, jenis_kelamin, agama, alamat, asal_sekolah, no_hp, nama_ayah, nama_ibu, jurusan, status, cancel_reason, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		reg.ID, reg.Nama, reg.Nik, reg.Nisn, reg.KkNo, reg.TempatLahir, reg.TanggalLahir, reg.JenisKelamin, reg.Agama, reg.Alamat, reg.AsalSekolah, reg.NoHP, reg.NamaAyah, reg.NamaIbu, reg.Jurusan, reg.Status, reg.CancelReason, reg.CreatedAt, reg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert spmb registration: %w", err)
	}
	return nil
}

func (r *Repository) ByID(ctx context.Context, id id.ID) (*spmb.SpmbRegistration, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+columns+` FROM spmb_registrations WHERE id = $1`, id,
	)
	reg := &spmb.SpmbRegistration{}
	if err := scan(row, reg); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return reg, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]spmb.SpmbRegistration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+` FROM spmb_registrations ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list spmb registrations: %w", err)
	}
	defer rows.Close()

	var list []spmb.SpmbRegistration
	for rows.Next() {
		var reg spmb.SpmbRegistration
		if err := scan(rows, &reg); err != nil {
			return nil, err
		}
		list = append(list, reg)
	}
	return list, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, reg *spmb.SpmbRegistration, expectedStatus string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE spmb_registrations SET status = $2, cancel_reason = $3, updated_at = NOW()
		 WHERE id = $1 AND status = $4`,
		reg.ID, reg.Status, reg.CancelReason, expectedStatus,
	)
	if err != nil {
		return fmt.Errorf("update spmb registration status: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("registration status changed")
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(row rowScanner, reg *spmb.SpmbRegistration) error {
	err := row.Scan(&reg.ID, &reg.Nama, &reg.Nik, &reg.Nisn, &reg.KkNo, &reg.TempatLahir, &reg.TanggalLahir, &reg.JenisKelamin, &reg.Agama, &reg.Alamat, &reg.AsalSekolah, &reg.NoHP, &reg.NamaAyah, &reg.NamaIbu, &reg.Jurusan, &reg.Status, &reg.CancelReason, &reg.CreatedAt, &reg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("scan spmb registration: %w", err)
	}
	return nil
}

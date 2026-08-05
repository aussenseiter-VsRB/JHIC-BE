package spmb

import (
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

const (
	StatusProses  = "proses"
	StatusApprove = "approve"
	StatusCancel  = "cancel"
)

type SpmbRegistration struct {
	ID           id.ID     `json:"id"`
	Nama         string    `json:"nama"`
	Nik          string    `json:"nik"`
	Nisn         string    `json:"nisn,omitempty"`
	KkNo         string    `json:"kk_no,omitempty"`
	TempatLahir  string    `json:"tempat_lahir,omitempty"`
	TanggalLahir string    `json:"tanggal_lahir,omitempty"`
	JenisKelamin string    `json:"jenis_kelamin"`
	Agama        string    `json:"agama,omitempty"`
	Alamat       string    `json:"alamat"`
	AsalSekolah  string    `json:"asal_sekolah,omitempty"`
	NoHP         string    `json:"no_hp,omitempty"`
	NamaAyah     string    `json:"nama_ayah,omitempty"`
	NamaIbu      string    `json:"nama_ibu,omitempty"`
	Jurusan      string    `json:"jurusan"`
	Status       string    `json:"status"`
	CancelReason string    `json:"cancel_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

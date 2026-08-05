package spmb_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/mocks"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/spmb"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/spmb/content"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validKkJSON = `{
	"nama": "Budi Santoso",
	"nik": "3204123456789012",
	"kk_no": "3204012312310001",
	"tempat_lahir": "Bandung",
	"tanggal_lahir": "2010-05-10",
	"jenis_kelamin": "LAKI-LAKI",
	"agama": "Islam",
	"alamat": "Jl. Raya Soreang No. 1",
	"nama_ayah": "Ahmad Santoso",
	"nama_ibu": "Siti Aminah"
}`

func TestService_ParseKk(t *testing.T) {
	t.Run("success forwards and normalizes", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("SpmbParseKk", mock.Anything, "base64data", "image/jpeg", "Budi").Return(validKkJSON, nil)

		svc := spmb.NewService(client)
		got, err := svc.ParseKk(context.Background(), "base64data", "image/jpeg", "  Budi  ")
		require.NoError(t, err)
		require.Equal(t, "Budi Santoso", got.Nama)
		require.Equal(t, "3204123456789012", got.Nik)
		require.Equal(t, "Laki-laki", got.JenisKelamin)
	})

	t.Run("child name required", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := spmb.NewService(client)
		_, err := svc.ParseKk(context.Background(), "data", "image/jpeg", "   ")
		require.ErrorIs(t, err, spmb.ErrChildNameRequired)
	})

	t.Run("child name too long", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := spmb.NewService(client)
		_, err := svc.ParseKk(context.Background(), "data", "image/jpeg", strings.Repeat("a", content.QuestionMaxLen+10))
		require.ErrorIs(t, err, spmb.ErrChildNameTooLong)
	})

	t.Run("empty file rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := spmb.NewService(client)
		_, err := svc.ParseKk(context.Background(), "", "image/jpeg", "Budi")
		require.ErrorIs(t, err, spmb.ErrKkFileRequired)
	})

	t.Run("invalid model output rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("SpmbParseKk", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("bukan json", nil)

		svc := spmb.NewService(client)
		_, err := svc.ParseKk(context.Background(), "data", "image/jpeg", "Budi")
		require.ErrorIs(t, err, spmb.ErrOutputInvalid)
	})

	t.Run("bad nik rejected", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("SpmbParseKk", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(`{"nama":"Budi","nik":"123"}`, nil)

		svc := spmb.NewService(client)
		_, err := svc.ParseKk(context.Background(), "data", "image/jpeg", "Budi")
		require.ErrorIs(t, err, spmb.ErrOutputInvalid)
	})
}

func TestService_Ask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		client.On("SpmbAsk", mock.Anything, "Kapan SPMB dibuka?", "sess").Return("Juni 2026", nil)

		svc := spmb.NewService(client)
		got, err := svc.Ask(context.Background(), "  Kapan SPMB dibuka?  ", "sess")
		require.NoError(t, err)
		require.Equal(t, "Juni 2026", got.Output)
	})

	t.Run("question required", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := spmb.NewService(client)
		_, err := svc.Ask(context.Background(), "   ", "")
		require.ErrorIs(t, err, spmb.ErrQuestionRequired)
	})

	t.Run("question too long", func(t *testing.T) {
		client := mocks.NewN8NClient(t)
		svc := spmb.NewService(client)
		_, err := svc.Ask(context.Background(), strings.Repeat("q", content.QuestionMaxLen+1), "")
		require.ErrorIs(t, err, spmb.ErrQuestionTooLong)
	})
}

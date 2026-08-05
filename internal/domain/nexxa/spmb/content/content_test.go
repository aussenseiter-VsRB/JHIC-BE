package content_test

import (
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/spmb/content"
	"github.com/stretchr/testify/require"
)

func TestNormalizeKkOutput(t *testing.T) {
	t.Run("plain json", func(t *testing.T) {
		res, errs := content.NormalizeKkOutput(`{"nama":"Budi","nik":"3204123456789012","kk_no":"KK123","jenis_kelamin":"perempuan"}`)
		require.Nil(t, errs)
		require.Equal(t, "Budi", res.Nama)
		require.Equal(t, "Perempuan", res.JenisKelamin)
	})

	t.Run("fenced json extracted", func(t *testing.T) {
		res, errs := content.NormalizeKkOutput("```json\n" + `{"nik":"3204123456789012"}` + "\n```")
		require.Nil(t, errs)
		require.Equal(t, "3204123456789012", res.Nik)
	})

	t.Run("nik with spaces normalized", func(t *testing.T) {
		res, errs := content.NormalizeKkOutput(`{"nik":"3204 1234 5678 9012"}`)
		require.Nil(t, errs)
		require.Equal(t, "3204123456789012", res.Nik)
	})

	t.Run("empty raw", func(t *testing.T) {
		_, errs := content.NormalizeKkOutput("  ")
		require.Len(t, errs, 1)
	})

	t.Run("missing nik", func(t *testing.T) {
		_, errs := content.NormalizeKkOutput(`{"nama":"Budi"}`)
		require.Len(t, errs, 1)
		require.Equal(t, "nik", errs[0].Field)
	})

	t.Run("invalid nik", func(t *testing.T) {
		_, errs := content.NormalizeKkOutput(`{"nik":"123"}`)
		require.Len(t, errs, 1)
	})

	t.Run("unparseable", func(t *testing.T) {
		_, errs := content.NormalizeKkOutput("totally not json")
		require.Len(t, errs, 1)
	})
}

func TestIsValidJurusan(t *testing.T) {
	require.True(t, content.IsValidJurusan("PPLG"))
	require.True(t, content.IsValidJurusan("Akuntansi"))
	require.False(t, content.IsValidJurusan("Fisika"))
}

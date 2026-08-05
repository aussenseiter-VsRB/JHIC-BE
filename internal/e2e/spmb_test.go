//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb"
	"github.com/stretchr/testify/require"
)

const kkStubOutput = `{"nama":"Budi Santoso","nik":"3204123456789012","kk_no":"3204012312310001","tempat_lahir":"Bandung","tanggal_lahir":"2010-05-10","jenis_kelamin":"LAKI-LAKI","agama":"Islam","alamat":"Jl. Raya Soreang No. 1","nama_ayah":"Ahmad Santoso","nama_ibu":"Siti Aminah"}`

func TestE2E_SpmbRegistrationLifecycle(t *testing.T) {
	e := startE2E(t)
	adminID, adminToken := register(t, e.server.URL, "spmb-admin@test.dev")
	promoteToAdmin(t, e, adminID)

	resp := doJSON(t, http.MethodPost, e.server.URL+"/api/v1/spmb", "", map[string]any{
		"nama":          "Budi Santoso",
		"nik":           "3204123456789012",
		"jenis_kelamin": "Laki-laki",
		"alamat":        "Jl. Raya Soreang No. 1",
		"jurusan":       "PPLG",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg spmb.SpmbRegistration
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reg))
	resp.Body.Close()
	require.Equal(t, spmb.StatusProses, reg.Status)
	require.NotEmpty(t, reg.ID.String())

	ctx := context.Background()
	row := e.pool.QueryRow(ctx, `SELECT status FROM spmb_registrations WHERE id = $1`, reg.ID)
	var status string
	require.NoError(t, row.Scan(&status))
	require.Equal(t, spmb.StatusProses, status)

	resp = doJSON(t, http.MethodGet, e.server.URL+"/api/v1/spmb", adminToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list []spmb.SpmbRegistration
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	resp.Body.Close()
	require.Len(t, list, 1)

	resp = doJSON(t, http.MethodGet, e.server.URL+"/api/v1/spmb/"+reg.ID.String(), adminToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = doJSON(t, http.MethodPost, e.server.URL+"/api/v1/spmb/"+reg.ID.String()+"/status", adminToken, map[string]string{"status": "approve"})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var updated spmb.SpmbRegistration
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	resp.Body.Close()
	require.Equal(t, spmb.StatusApprove, updated.Status)

	row = e.pool.QueryRow(ctx, `SELECT status FROM spmb_registrations WHERE id = $1`, reg.ID)
	require.NoError(t, row.Scan(&status))
	require.Equal(t, spmb.StatusApprove, status)
}

func TestE2E_SpmbCreateValidation(t *testing.T) {
	e := startE2E(t)

	resp := doJSON(t, http.MethodPost, e.server.URL+"/api/v1/spmb", "", map[string]any{
		"nama": "Budi",
		"nik":  "123",
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_SpmbAdminRequiresAuth(t *testing.T) {
	e := startE2E(t)

	resp := doJSON(t, http.MethodGet, e.server.URL+"/api/v1/spmb", "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_SpmbParseKk(t *testing.T) {
	e := startE2E(t)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "kk.jpg")
	require.NoError(t, err)
	_, err = fw.Write([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10})
	require.NoError(t, err)
	require.NoError(t, mw.WriteField("child_name", "Budi"))
	require.NoError(t, mw.Close())

	resp := doMultipart(t, e.server.URL+"/api/v1/nexxa/spmb/parse-kk", mw.FormDataContentType(), body.Bytes())
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		Data map[string]string `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	resp.Body.Close()
	require.Equal(t, "Budi Santoso", out.Data["nama"])
	require.Equal(t, "3204123456789012", out.Data["nik"])
	require.Equal(t, "Laki-laki", out.Data["jenis_kelamin"])
}

func TestE2E_SpmbAsk(t *testing.T) {
	e := startE2E(t)

	resp := doJSON(t, http.MethodPost, e.server.URL+"/api/v1/nexxa/spmb/ask", "", map[string]string{
		"question": "Kapan SPMB dibuka?",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		Output string `json:"output"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	resp.Body.Close()
	require.Equal(t, "SPMB dibuka Juni 2026", out.Output)
}

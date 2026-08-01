//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2E_AIChat(t *testing.T) {
	e := startE2E(t)
	url := e.server.URL + "/api/v1/ai/chat"

	t.Run("valid message returns n8n output", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]string{
			"chatInput": "halo",
			"sessionId": "session-1",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var out struct {
			Output string `json:"output"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		resp.Body.Close()
		require.Equal(t, "hai dari nexxa", out.Output)
	})

	t.Run("empty message rejected", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]string{"chatInput": "   "})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("overlong message rejected", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]string{
			"chatInput": strings.Repeat("a", 301),
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("malformed body rejected", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]any{
			"chatInput": "halo",
			"sessionId": []string{"a", "b"},
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestE2E_AINexxaMatch(t *testing.T) {
	e := startE2E(t)
	url := e.server.URL + "/api/v1/ai/nexxa-match"

	valid := map[string]string{
		"jawaban_1": "a", "jawaban_2": "b", "jawaban_3": "c", "jawaban_4": "d",
		"jawaban_5": "e", "jawaban_6": "f", "jawaban_7": "g", "jawaban_8": "h",
	}

	t.Run("valid answers return recommendation", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", valid)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var out struct {
			NamaJurusan string `json:"nama_jurusan"`
			Alasan      string `json:"alasan"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		resp.Body.Close()
		require.Equal(t, "PPLG", out.NamaJurusan)
		require.Equal(t, "cocok", out.Alasan)
	})

	t.Run("missing answers rejected", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]string{"jawaban_1": "a"})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("blank answer rejected", func(t *testing.T) {
		invalid := make(map[string]string, 8)
		for k, v := range valid {
			invalid[k] = v
		}
		invalid["jawaban_4"] = "  "
		resp := doJSON(t, http.MethodPost, url, "", invalid)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

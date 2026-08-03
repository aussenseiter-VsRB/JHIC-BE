//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const cvReviewStubOutput = `{
	"audit_summary": {
		"score": 84,
		"tier_label": "Kandidat Kuat",
		"grade_label": "B+",
		"summary_text": "Profil yang solid dengan pengalaman relevan.",
		"key_strengths": ["pengalaman relevan", "portofolio lengkap"],
		"key_improvements": ["tambahkan kata kunci"]
	},
	"metrics": {"format_score": 88, "ats_status": "good"},
	"grammar_issues": [
		{"text": "Saya bekerja di PT ABC", "suggestion": "Saya bekerja di PT ABC.", "location": "Pengalaman Kerja"}
	],
	"recommendations": [
		{
			"id": 1,
			"priority": "Normal",
			"category": "keywords",
			"section": "Ringkasan Profil",
			"title": "Tambahkan kata kunci",
			"description": "Optimalkan ringkasan dengan kata kunci peran target."
		}
	],
	"strengths_detail": [
		{"id": 1, "category": "structure", "title": "Struktur jelas", "description": "Bagian CV mudah dipindai."}
	]
}`

func TestE2E_CvReview(t *testing.T) {
	e := startE2E(t)
	_, token := register(t, e.server.URL, "cvuser@example.com")
	url := e.server.URL + "/api/v1/nexxa/cv-review"

	t.Run("requires auth", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]any{
			"cv_text": "CV saya", "word_count": 3, "page_count": 1,
		})
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("valid cv returns normalized audit", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, token, map[string]any{
			"cv_text":    "Nama saya Budi, lulusan SMK.",
			"word_count": 7,
			"page_count": 1,
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var out struct {
			AuditSummary struct {
				Score int `json:"score"`
			} `json:"audit_summary"`
			Metrics struct {
				ATSStatus string `json:"ats_status"`
			} `json:"metrics"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		resp.Body.Close()
		require.Equal(t, 84, out.AuditSummary.Score)
		require.Equal(t, "good", out.Metrics.ATSStatus)
	})

	t.Run("blank cv rejected", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, token, map[string]any{"cv_text": "   "})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestE2E_CvReviewValidateInput(t *testing.T) {
	e := startE2E(t)
	url := e.server.URL + "/api/v1/nexxa/cv-review/validate-input"

	t.Run("public and sanitizes", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]any{
			"cv_text":    "<b>Nama</b> <script>alert(1)</script>Saya suka komputer",
			"word_count": 3,
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var out struct {
			Success bool           `json:"success"`
			Data    map[string]any `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		resp.Body.Close()
		require.True(t, out.Success)
		require.Equal(t, "Nama Saya suka komputer", out.Data["cv_text"])
		require.Equal(t, float64(3), out.Data["word_count"])
	})

	t.Run("missing cv_text rejected", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]any{"word_count": 1})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestE2E_CvReviewNormalizeOutput(t *testing.T) {
	e := startE2E(t)
	url := e.server.URL + "/api/v1/nexxa/cv-review/normalize-output"

	t.Run("well-formed output returns 200", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]string{"raw": cvReviewStubOutput})
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var out struct {
			Success bool `json:"success"`
			Data    struct {
				Metrics struct {
					ATSStatus string `json:"ats_status"`
				} `json:"metrics"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		resp.Body.Close()
		require.True(t, out.Success)
		require.Equal(t, "good", out.Data.Metrics.ATSStatus)
	})

	t.Run("unparseable output returns 422", func(t *testing.T) {
		resp := doJSON(t, http.MethodPost, url, "", map[string]string{"raw": "bukan json"})
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		resp.Body.Close()
	})
}
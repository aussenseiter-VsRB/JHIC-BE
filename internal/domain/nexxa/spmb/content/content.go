package content

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

const QuestionMaxLen = 300

type APIError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

var (
	wsRe          = regexp.MustCompile(`\s+`)
	fenceRe       = regexp.MustCompile("(?i)```[a-z]*")
	nikRe         = regexp.MustCompile(`^\d{16}$`)
	genderRe      = regexp.MustCompile(`(?i)^(laki.laki|perempuan|l|p)$`)
	jurusanValues = []string{"PPLG", "Akuntansi", "Perhotelan", "Hotel"}

	injectionRe = regexp.MustCompile(`(?i)(ignore\s+(all\s+)?previous\s+instructions|disregard\s+(all\s+)?previous\s+instructions|system\s*:|you\s+are\s+now|as\s+(an\s+)?(ai|language\s+model)|forget\s+(all\s+)?(previous|prior)\s+instructions|developer\s+mode|jailbreak)`)
)

type ParseKkResult struct {
	Nama         string `json:"nama"`
	Nik          string `json:"nik"`
	KkNo         string `json:"kk_no"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	JenisKelamin string `json:"jenis_kelamin"`
	Agama        string `json:"agama"`
	Alamat       string `json:"alamat"`
	NamaAyah     string `json:"nama_ayah"`
	NamaIbu      string `json:"nama_ibu"`
}

func SanitizeQuestion(s string) string {
	return strings.TrimSpace(s)
}

func HasPromptInjection(s string) bool {
	return injectionRe.MatchString(s)
}

func SanitizeChildName(s string) string {
	return strings.TrimSpace(s)
}

// NormalizeKkOutput parses and repairs the vision model's KK extraction JSON.
// Missing optional fields become empty strings; a missing NIK fails the parse.
func NormalizeKkOutput(raw string) (*ParseKkResult, []APIError) {
	if strings.TrimSpace(raw) == "" {
		return nil, []APIError{{Message: "Empty output from model."}}
	}

	obj, err := parseModelOutput(raw)
	if err != nil {
		return nil, []APIError{{Message: "Could not parse a valid JSON object from model output."}}
	}

	nik, ok := objString(obj, "nik")
	if !ok {
		return nil, []APIError{{Field: "nik", Message: "Missing NIK from KK."}}
	}
	nik = strings.ReplaceAll(nik, " ", "")
	if !nikRe.MatchString(nik) {
		return nil, []APIError{{Field: "nik", Message: "NIK must be 16 digits."}}
	}

	res := &ParseKkResult{
		Nama:         objStringOrEmpty(obj, "nama"),
		Nik:          nik,
		KkNo:         objStringOrEmpty(obj, "kk_no"),
		TempatLahir:  objStringOrEmpty(obj, "tempat_lahir"),
		TanggalLahir: objStringOrEmpty(obj, "tanggal_lahir"),
		JenisKelamin: normalizeGender(objStringOrEmpty(obj, "jenis_kelamin")),
		Agama:        objStringOrEmpty(obj, "agama"),
		Alamat:       objStringOrEmpty(obj, "alamat"),
		NamaAyah:     objStringOrEmpty(obj, "nama_ayah"),
		NamaIbu:      objStringOrEmpty(obj, "nama_ibu"),
	}
	return res, nil
}

func parseModelOutput(raw string) (map[string]json.RawMessage, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = fenceRe.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)

	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(cleaned), &obj); err == nil {
		return obj, nil
	}
	if start := strings.IndexByte(cleaned, '{'); start >= 0 {
		if end := strings.LastIndexByte(cleaned, '}'); end > start {
			var fallback map[string]json.RawMessage
			if err := json.Unmarshal([]byte(cleaned[start:end+1]), &fallback); err == nil {
				return fallback, nil
			}
		}
	}
	return nil, errors.New("no valid JSON object found in model output")
}

func objString(m map[string]json.RawMessage, key string) (string, bool) {
	rv, ok := m[key]
	if !ok {
		return "", false
	}
	var s string
	if err := json.Unmarshal(rv, &s); err != nil {
		return "", false
	}
	s = strings.TrimSpace(s)
	return s, s != ""
}

func objStringOrEmpty(m map[string]json.RawMessage, key string) string {
	s, _ := objString(m, key)
	return s
}

func normalizeGender(s string) string {
	switch {
	case genderRe.MatchString(s) && (strings.HasPrefix(strings.ToLower(s), "laki") || strings.EqualFold(s, "l") || strings.EqualFold(s, "laki-laki")):
		return "Laki-laki"
	case genderRe.MatchString(s):
		return "Perempuan"
	}
	return s
}

func IsValidJurusan(j string) bool {
	for _, v := range jurusanValues {
		if v == j {
			return true
		}
	}
	return false
}

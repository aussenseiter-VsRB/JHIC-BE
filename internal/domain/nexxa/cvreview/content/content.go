package content

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const CvTextMaxLen = 50_000

type APIError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

var (
	scriptBlockRe = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleBlockRe  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	htmlTagRe     = regexp.MustCompile(`(?i)<[^>]+>`)
	wsRe          = regexp.MustCompile(`\s+`)
	fenceRe       = regexp.MustCompile("(?i)```[a-z]*")

	injectionRe = regexp.MustCompile(`(?i)(ignore\s+(all\s+)?previous\s+instructions|disregard\s+(all\s+)?previous\s+instructions|system\s*:|you\s+are\s+now|as\s+(an\s+)?(ai|language\s+model)|forget\s+(all\s+)?(previous|prior)\s+instructions|developer\s+mode|jailbreak)`)

	atsStatusRe = regexp.MustCompile(`(?i)^(good|needs_improvement|poor)$`)
	priorityRe  = regexp.MustCompile(`(?i)^(urgent|normal)$`)
	categoryRe  = regexp.MustCompile(`(?i)^(content|ats_format|structure|keywords)$`)

	priorityValues  = []string{"Urgent", "Normal"}
	atsStatusValues = []string{"good", "needs_improvement", "poor"}
	categoryValues  = []string{"content", "ats_format", "structure", "keywords"}
)

func sanitizeText(s string) string {
	s = scriptBlockRe.ReplaceAllString(s, " ")
	s = styleBlockRe.ReplaceAllString(s, " ")
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = wsRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func HasPromptInjection(s string) bool {
	return injectionRe.MatchString(s)
}

func sanitizeForLog(s string) string {
	s = wsRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) > 100 {
		s = s[:100] + "..."
	}
	return s
}

// ValidateCvInput validates and sanitizes a raw CV payload. It returns a map
// containing the sanitized cv_text and the (optional, non-negative) word_count
// and page_count, or a list of field errors. word_count/page_count are
// backend-computed context for the model, not instructions.
func ValidateCvInput(raw map[string]json.RawMessage) (map[string]any, []APIError) {
	var errs []APIError

	rv, ok := raw["cv_text"]
	if !ok {
		errs = append(errs, APIError{Field: "cv_text", Message: "This field is required."})
		return nil, errs
	}
	var s string
	if err := json.Unmarshal(rv, &s); err != nil {
		errs = append(errs, APIError{Field: "cv_text", Message: "Must be a plain string."})
		return nil, errs
	}
	s = strings.TrimSpace(s)
	if s == "" {
		errs = append(errs, APIError{Field: "cv_text", Message: "This field is required."})
		return nil, errs
	}
	if len(s) > CvTextMaxLen {
		errs = append(errs, APIError{Field: "cv_text", Message: "Must be 50000 characters or fewer."})
		return nil, errs
	}
	sanitized := sanitizeText(s)
	if sanitized == "" {
		errs = append(errs, APIError{Field: "cv_text", Message: "This field is required."})
		return nil, errs
	}

	out := map[string]any{
		"cv_text":    sanitized,
		"word_count": 0,
		"page_count": 0,
	}

	if wc, wcOK := parseString(raw, "word_count"); wcOK {
		w, err := strconv.Atoi(wc)
		if err != nil || w < 0 {
			errs = append(errs, APIError{Field: "word_count", Message: "Must be zero or a positive integer."})
		} else {
			out["word_count"] = w
		}
	}
	if pc, pcOK := parseString(raw, "page_count"); pcOK {
		p, err := strconv.Atoi(pc)
		if err != nil || p < 0 {
			errs = append(errs, APIError{Field: "page_count", Message: "Must be zero or a positive integer."})
		} else {
			out["page_count"] = p
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return out, nil
}

func parseString(raw map[string]json.RawMessage, key string) (string, bool) {
	rv, ok := raw[key]
	if !ok {
		return "", false
	}
	// Accept JSON numbers (50 => "50") as well as strings.
	trimmed := strings.TrimSpace(string(rv))
	if trimmed == "" {
		return "", false
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(rv, &s); err != nil {
			return "", false
		}
		s = strings.TrimSpace(s)
		return s, s != ""
	}
	return trimmed, true
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

func clampInt(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
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

func objInt(m map[string]json.RawMessage, key string) (int, bool) {
	rv, ok := m[key]
	if !ok {
		return 0, false
	}
	var n int
	if err := json.Unmarshal(rv, &n); err == nil {
		return n, true
	}
	var s string
	if err := json.Unmarshal(rv, &s); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			return n, true
		}
	}
	return 0, false
}

func objStringOrEmpty(m map[string]json.RawMessage, key string) string {
	s, _ := objString(m, key)
	return s
}

func normalizeStringSlice(m map[string]json.RawMessage, key string, max int) []string {
	var out []string
	if rv, ok := m[key]; ok {
		var list []string
		if err := json.Unmarshal(rv, &list); err == nil {
			for _, s := range list {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				out = append(out, s)
				if len(out) >= max {
					break
				}
			}
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

func isValidEnum(re *regexp.Regexp, s string) bool {
	return re.MatchString(s)
}

func normalizeEnum(values []string, s string) string {
	for _, v := range values {
		if strings.EqualFold(v, s) {
			return v
		}
	}
	return s
}

func normalizeGrammarIssues(obj map[string]json.RawMessage) ([]GrammarIssue, bool) {
	rv, ok := obj["grammar_issues"]
	if !ok {
		return []GrammarIssue{}, true
	}
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(rv, &list); err != nil {
		return nil, false
	}
	var out []GrammarIssue
	for _, item := range list {
		text, ok1 := objString(item, "text")
		suggestion, ok2 := objString(item, "suggestion")
		location, ok3 := objString(item, "location")
		if !ok1 || !ok2 || !ok3 {
			return nil, false
		}
		out = append(out, GrammarIssue{
			Text:       text,
			Suggestion: suggestion,
			Location:   location,
		})
		if len(out) >= 8 {
			break
		}
	}
	if out == nil {
		out = []GrammarIssue{}
	}
	return out, true
}

func normalizeRecommendations(obj map[string]json.RawMessage) ([]Recommendation, bool) {
	rv, ok := obj["recommendations"]
	if !ok {
		return []Recommendation{}, true
	}
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(rv, &list); err != nil {
		return nil, false
	}
	var out []Recommendation
	for _, item := range list {
		priority, ok1 := objString(item, "priority")
		category, ok2 := objString(item, "category")
		section, ok3 := objString(item, "section")
		title, ok4 := objString(item, "title")
		description, ok5 := objString(item, "description")
		if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
			return nil, false
		}
		if !isValidEnum(priorityRe, priority) || !isValidEnum(categoryRe, category) {
			return nil, false
		}
		before := objStringOrEmpty(item, "before_text")
		after := objStringOrEmpty(item, "after_text")
		r := Recommendation{
			ID:          len(out) + 1,
			Priority:    normalizeEnum(priorityValues, priority),
			Category:    normalizeEnum(categoryValues, category),
			Section:     section,
			Title:       title,
			Description: description,
			BeforeText:  before,
			AfterText:   after,
			HasExample:  before != "" || after != "",
		}
		if !r.HasExample {
			r.BeforeText = ""
			r.AfterText = ""
		}
		out = append(out, r)
		if len(out) >= 10 {
			break
		}
	}
	if out == nil {
		out = []Recommendation{}
	}
	return out, true
}

func normalizeStrengthsDetail(obj map[string]json.RawMessage) ([]StrengthDetail, bool) {
	rv, ok := obj["strengths_detail"]
	if !ok {
		return []StrengthDetail{}, true
	}
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(rv, &list); err != nil {
		return nil, false
	}
	var out []StrengthDetail
	for _, item := range list {
		category, ok1 := objString(item, "category")
		title, ok2 := objString(item, "title")
		description, ok3 := objString(item, "description")
		if !ok1 || !ok2 || !ok3 || !isValidEnum(categoryRe, category) {
			return nil, false
		}
		out = append(out, StrengthDetail{
			ID:          len(out) + 1,
			Category:    normalizeEnum(categoryValues, category),
			Title:       title,
			Description: description,
		})
		if len(out) >= 6 {
			break
		}
	}
	if out == nil {
		out = []StrengthDetail{}
	}
	return out, true
}

type AuditSummary struct {
	Score           int      `json:"score"`
	TierLabel       string   `json:"tier_label"`
	GradeLabel      string   `json:"grade_label"`
	SummaryText     string   `json:"summary_text"`
	KeyStrengths    []string `json:"key_strengths"`
	KeyImprovements []string `json:"key_improvements"`
}

type Metrics struct {
	FormatScore int    `json:"format_score"`
	ATSStatus   string `json:"ats_status"`
}

type GrammarIssue struct {
	Text       string `json:"text"`
	Suggestion string `json:"suggestion"`
	Location   string `json:"location"`
}

type Recommendation struct {
	ID          int    `json:"id"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Section     string `json:"section"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BeforeText  string `json:"before_text,omitempty"`
	AfterText   string `json:"after_text,omitempty"`
	HasExample  bool   `json:"has_example"`
}

type StrengthDetail struct {
	ID          int    `json:"id"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type NormalizeOutputData struct {
	AuditSummary    AuditSummary     `json:"audit_summary"`
	Metrics         Metrics          `json:"metrics"`
	GrammarIssues   []GrammarIssue   `json:"grammar_issues"`
	Recommendations []Recommendation `json:"recommendations"`
	StrengthsDetail []StrengthDetail `json:"strengths_detail"`
}

func NormalizeCvOutput(raw string) (*NormalizeOutputData, []APIError) {
	if strings.TrimSpace(raw) == "" {
		return nil, []APIError{{Message: "Empty output from model."}}
	}

	obj, err := parseModelOutput(raw)
	if err != nil {
		return nil, []APIError{{Message: "Could not parse a valid JSON object from model output."}}
	}

	as, ok := obj["audit_summary"]
	if !ok {
		return nil, []APIError{{Message: "Missing audit_summary."}}
	}
	var auditMap map[string]json.RawMessage
	if err := json.Unmarshal(as, &auditMap); err != nil {
		return nil, []APIError{{Message: "Invalid audit_summary."}}
	}

	score, scoreOK := objInt(auditMap, "score")
	tier, tierOK := objString(auditMap, "tier_label")
	grade, gradeOK := objString(auditMap, "grade_label")
	summary, summaryOK := objString(auditMap, "summary_text")
	if !scoreOK || !tierOK || !gradeOK || !summaryOK {
		return nil, []APIError{{Message: "Incomplete audit_summary."}}
	}

	ms, ok := obj["metrics"]
	if !ok {
		return nil, []APIError{{Message: "Missing metrics."}}
	}
	var metricsMap map[string]json.RawMessage
	if err := json.Unmarshal(ms, &metricsMap); err != nil {
		return nil, []APIError{{Message: "Invalid metrics."}}
	}
	formatScore, fsOK := objInt(metricsMap, "format_score")
	atsStatus, atsOK := objString(metricsMap, "ats_status")
	if !fsOK || !atsOK || !isValidEnum(atsStatusRe, atsStatus) {
		return nil, []APIError{{Message: "Invalid metrics."}}
	}

	grammarIssues, ok := normalizeGrammarIssues(obj)
	if !ok {
		return nil, []APIError{{Message: "Invalid grammar_issues."}}
	}
	recommendations, ok := normalizeRecommendations(obj)
	if !ok {
		return nil, []APIError{{Message: "Invalid recommendations."}}
	}
	strengthsDetail, ok := normalizeStrengthsDetail(obj)
	if !ok {
		return nil, []APIError{{Message: "Invalid strengths_detail."}}
	}

	return &NormalizeOutputData{
		AuditSummary: AuditSummary{
			Score:           clampInt(score),
			TierLabel:       tier,
			GradeLabel:      grade,
			SummaryText:     summary,
			KeyStrengths:    normalizeStringSlice(auditMap, "key_strengths", 6),
			KeyImprovements: normalizeStringSlice(auditMap, "key_improvements", 6),
		},
		Metrics: Metrics{
			FormatScore: clampInt(formatScore),
			ATSStatus:   normalizeEnum(atsStatusValues, atsStatus),
		},
		GrammarIssues:   grammarIssues,
		Recommendations: recommendations,
		StrengthsDetail: strengthsDetail,
	}, nil
}
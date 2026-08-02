package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func logNexxaInput(r *http.Request, ok bool, data map[string]string, errs []APIError) {
	flag := "ok"
	if !ok {
		flag = "failed"
	}
	fields := make([]string, 0, NexxaAnswerCount)
	for i := 1; i <= NexxaAnswerCount; i++ {
		key := fmt.Sprintf("jawaban_%d", i)
		fields = append(fields, fmt.Sprintf("%s:len=%d,sha=%s", key, len(data[key]), shortSHA(data[key])))
	}
	detail := strings.Join(fields, " ")
	if len(errs) > 0 {
		detail += " errors=" + errorFields(errs)
	}
	log.Printf("nexxa validate-input: %s %s %s", flag, r.URL.Path, detail)
	for i := 1; i <= NexxaAnswerCount; i++ {
		key := fmt.Sprintf("jawaban_%d", i)
		if hasPromptInjection(data[key]) {
			log.Printf("nexxa validate-input: WARNING suspicious input flagged in %s (sha=%s)", key, shortSHA(data[key]))
		}
	}
}

func logNexxaOutput(r *http.Request, ok bool, raw string, data *NormalizeOutputData, errs []APIError) {
	flag := "ok"
	if !ok {
		flag = "failed"
	}
	detail := fmt.Sprintf("raw:len=%d,sha=%s", len(raw), shortSHA(raw))
	if data != nil {
		detail += fmt.Sprintf(" major=%s pct=%d/%d/%d", data.NamaJurusan, data.PersentasePPLG, data.PersentaseAkuntansi, data.PersentaseHotel)
	}
	if len(errs) > 0 {
		detail += " errors=" + errorFields(errs)
	}
	log.Printf("nexxa normalize-output: %s %s %s", flag, r.URL.Path, detail)
}

func errorFields(errs []APIError) string {
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		if e.Field != "" {
			out = append(out, fmt.Sprintf("%s:%s", e.Field, e.Message))
		} else {
			out = append(out, e.Message)
		}
	}
	return strings.Join(out, ",")
}

func shortSHA(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:4])
}

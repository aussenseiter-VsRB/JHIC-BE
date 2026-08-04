package analytics

import (
	"context"
	"encoding/json"
	"time"
)

type Event struct {
	Name       string
	SessionID  string
	UserID     *int64
	Properties map[string]any
	CreatedAt  time.Time
}

type Repository interface {
	Record(context.Context, Event) error
	Summary(context.Context, time.Time, string) ([]Summary, error)
}

type Summary struct {
	Name  string `json:"event"`
	Count int    `json:"count"`
}

func Properties(v map[string]any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

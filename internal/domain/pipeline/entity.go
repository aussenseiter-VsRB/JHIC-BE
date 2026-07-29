package pipeline

import (
	"encoding/json"
	"time"
)

type Pipeline struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Status      string          `json:"status"`
	Config      json.RawMessage `json:"config,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

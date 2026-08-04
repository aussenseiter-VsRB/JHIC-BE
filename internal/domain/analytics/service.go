package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Record(ctx context.Context, name, session string, userID *int64, properties map[string]any) {
	if s == nil || s.repo == nil {
		return
	}
	_ = s.repo.Record(ctx, Event{Name: name, SessionID: HashSession(session), UserID: userID, Properties: properties, CreatedAt: time.Now().UTC()})
}

func HashSession(session string) string {
	sum := sha256.Sum256([]byte(session))
	return hex.EncodeToString(sum[:])
}

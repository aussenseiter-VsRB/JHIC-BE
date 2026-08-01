package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

type Service struct {
	client N8NClient
}

func NewService(client N8NClient) *Service {
	return &Service{client: client}
}

func (s *Service) Chat(ctx context.Context, chatInput, sessionID string) (*ChatResponse, error) {
	chatInput = strings.TrimSpace(chatInput)
	if chatInput == "" {
		return nil, ErrChatMessageRequired
	}
	if len(chatInput) > ChatMessageMaxLen {
		return nil, ErrChatMessageTooLong
	}
	if sessionID == "" {
		sessionID = newSessionID()
	}
	return s.client.Chat(ctx, chatInput, sessionID)
}

func (s *Service) NexxaMatch(ctx context.Context, answers []string) (*NexxaResponse, error) {
	if len(answers) != NexxaAnswerCount {
		return nil, ErrAnswersRequired
	}
	normalized := make([]string, NexxaAnswerCount)
	for i, a := range answers {
		a = strings.TrimSpace(a)
		if a == "" {
			return nil, ErrAnswersRequired
		}
		if len(a) > NexxaAnswerMaxLen {
			return nil, ErrAnswerTooLong
		}
		normalized[i] = a
	}
	return s.client.NexxaMatch(ctx, normalized)
}

func newSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

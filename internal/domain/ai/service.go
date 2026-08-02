package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	raw, err := s.client.NexxaMatch(ctx, normalized)
	if err != nil {
		return nil, err
	}

	data, errs := normalizeNexxaOutput(raw)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrNexxaOutputInvalid, errs[0].Message)
	}

	return &NexxaResponse{
		NamaJurusan:         data.NamaJurusan,
		Alasan:              data.Alasan,
		PersentasePPLG:      data.PersentasePPLG,
		PersentaseAkuntansi: data.PersentaseAkuntansi,
		PersentaseHotel:     data.PersentaseHotel,
	}, nil
}

func (s *Service) ValidateNexxaInput(raw map[string]json.RawMessage) (map[string]string, []APIError) {
	return validateNexxaInput(raw)
}

func (s *Service) NormalizeNexxaOutput(raw string) (*NormalizeOutputData, []APIError) {
	return normalizeNexxaOutput(raw)
}

func newSessionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

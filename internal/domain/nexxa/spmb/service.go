package spmb

import (
	"context"
	"errors"
	"strings"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/spmb/content"
)

type Service struct {
	client nexxa.N8NClient
}

func NewService(client nexxa.N8NClient) *Service {
	return &Service{client: client}
}

func (s *Service) ParseKk(ctx context.Context, imageBase64, mimeType, childName string) (*content.ParseKkResult, error) {
	childName = content.SanitizeChildName(childName)
	if childName == "" {
		return nil, ErrChildNameRequired
	}
	if len(childName) > ChildNameMaxLen {
		return nil, ErrChildNameTooLong
	}
	if imageBase64 == "" {
		return nil, ErrKkFileRequired
	}

	raw, err := s.client.SpmbParseKk(ctx, imageBase64, mimeType, childName)
	if err != nil {
		return nil, err
	}

	data, errs := content.NormalizeKkOutput(raw)
	if len(errs) > 0 {
		return nil, errors.Join(ErrOutputInvalid, errors.New(errs[0].Message))
	}
	return data, nil
}

func (s *Service) Ask(ctx context.Context, question, sessionID string) (*AskResponse, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, ErrQuestionRequired
	}
	if len(question) > content.QuestionMaxLen {
		return nil, ErrQuestionTooLong
	}

	raw, err := s.client.SpmbAsk(ctx, question, sessionID)
	if err != nil {
		return nil, err
	}
	return &AskResponse{Output: raw}, nil
}

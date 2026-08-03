package cvreview

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/cvreview/content"
)

type Service struct {
	client nexxa.N8NClient
}

func NewService(client nexxa.N8NClient) *Service {
	return &Service{client: client}
}

func (s *Service) CvReview(ctx context.Context, req CvReviewRequest) (*content.NormalizeOutputData, error) {
	cvText := strings.TrimSpace(req.CvText)
	if cvText == "" {
		return nil, ErrCvTextRequired
	}
	if len(cvText) > content.CvTextMaxLen {
		return nil, ErrCvTextTooLong
	}
	if req.WordCount < 0 || req.PageCount < 0 {
		return nil, ErrInvalidCounts
	}

	raw, err := s.client.CvReview(ctx, cvText, req.WordCount, req.PageCount)
	if err != nil {
		return nil, err
	}

	data, errs := content.NormalizeCvOutput(raw)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrCvOutputInvalid, errs[0].Message)
	}

	return data, nil
}

func (s *Service) ValidateCvInput(raw map[string]json.RawMessage) (map[string]any, []content.APIError) {
	return content.ValidateCvInput(raw)
}

func (s *Service) NormalizeCvOutput(raw string) (*content.NormalizeOutputData, []content.APIError) {
	return content.NormalizeCvOutput(raw)
}
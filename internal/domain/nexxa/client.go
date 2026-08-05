package nexxa

import "context"

type N8NClient interface {
	Chat(ctx context.Context, chatInput, sessionID string) (*ChatResponse, error)
	NexxaMatch(ctx context.Context, answers []string) (string, error)
	CvReview(ctx context.Context, cvText string, wordCount, pageCount int) (string, error)
	SpmbParseKk(ctx context.Context, imageBase64, mimeType, childName string) (string, error)
	SpmbAsk(ctx context.Context, question, sessionID string) (string, error)
}

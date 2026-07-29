package pipeline

import "context"

type Repository interface {
	Create(ctx context.Context, p *Pipeline) error
	ByID(ctx context.Context, id string) (*Pipeline, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]Pipeline, error)
	Update(ctx context.Context, p *Pipeline) error
	Delete(ctx context.Context, id string) error
}

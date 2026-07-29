package workspace

import "context"

type Repository interface {
	Create(ctx context.Context, w *Workspace) error
	ByID(ctx context.Context, id string) (*Workspace, error)
	ListByOwner(ctx context.Context, ownerID string) ([]Workspace, error)
	Update(ctx context.Context, w *Workspace) error
	Delete(ctx context.Context, id string) error
}

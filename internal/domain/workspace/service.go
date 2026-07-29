package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	now := time.Now().UTC()
	ws := &Workspace{
		ID:        id.New(),
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *Service) ByID(ctx context.Context, id string) (*Workspace, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]Workspace, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

func (s *Service) Update(ctx context.Context, id, name string) (*Workspace, error) {
	ws, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	ws.Name = name
	ws.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

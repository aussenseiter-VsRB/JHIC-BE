package pipeline

import (
	"context"
	"encoding/json"
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

func (s *Service) Create(ctx context.Context, workspaceID, name, description string) (*Pipeline, error) {
	now := time.Now().UTC()
	p := &Pipeline{
		ID:          id.New(),
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Status:      "inactive",
		Config:      json.RawMessage(`{}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ByID(ctx context.Context, id string) (*Pipeline, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string) ([]Pipeline, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *Service) Update(ctx context.Context, id, name, description, status string, config json.RawMessage) (*Pipeline, error) {
	p, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("pipeline not found")
	}
	p.Name = name
	p.Description = description
	p.Status = status
	if config != nil {
		p.Config = config
	}
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

package user

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

func (s *Service) Create(ctx context.Context, email, name string) (*User, error) {
	existing, err := s.repo.ByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	now := time.Now().UTC()
	user := &User{
		ID:        id.New(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) ByID(ctx context.Context, id string) (*User, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id, email, name string) (*User, error) {
	user, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	user.Email = email
	user.Name = name
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}



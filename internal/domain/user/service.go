package user

import (
	"context"
	"fmt"
	"slices"
	"time"
)

var ValidRoles = []string{"jurnal", "guru", "admin", "user"}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) ByID(ctx context.Context, id string) (*User, error) {
	return s.repo.ByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, name, avatarURL string) (*User, error) {
	user, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	user.Name = name
	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateRole(ctx context.Context, id, role string) error {
	if !slices.Contains(ValidRoles, role) {
		return fmt.Errorf("invalid role: must be one of %v", ValidRoles)
	}
	user, err := s.repo.ByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	return s.repo.UpdateRole(ctx, id, role)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

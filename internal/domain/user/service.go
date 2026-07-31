package user

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"golang.org/x/crypto/bcrypt"
)

var ValidRoles = []string{"jurnal", "guru", "admin", "user"}

var ValidPositions = []string{"wali_kelas", "bk", "kesiswaan", "kaprog"}

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

func (s *Service) Create(ctx context.Context, email, password, name, role, class, jurusan, position string) (*User, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	if !slices.Contains(ValidRoles, role) {
		return nil, fmt.Errorf("invalid role: must be one of %v", ValidRoles)
	}
	if err := validatePosition(position, role); err != nil {
		return nil, err
	}

	existing, err := s.repo.ByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	u := &User{
		ID:        id.New(),
		Email:     email,
		Name:      name,
		Role:      role,
		Class:     class,
		Jurusan:   jurusan,
		Position:  position,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, u, string(hash)); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id, name, avatarURL, class, jurusan, position string) (*User, error) {
	user, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if err := validatePosition(position, user.Role); err != nil {
		return nil, err
	}
	user.Name = name
	user.AvatarURL = avatarURL
	user.Class = class
	user.Jurusan = jurusan
	user.Position = position
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

func validatePosition(position, role string) error {
	if position == "" {
		return nil
	}
	if role != "guru" {
		return fmt.Errorf("position is only valid for role guru")
	}
	if !slices.Contains(ValidPositions, position) {
		return fmt.Errorf("invalid position: must be one of %v", ValidPositions)
	}
	return nil
}

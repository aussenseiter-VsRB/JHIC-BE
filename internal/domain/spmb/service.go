package spmb

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

func (s *Service) Create(ctx context.Context, input *SpmbRegistration) (*SpmbRegistration, error) {
	if err := validateCreate(input); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	input.ID = id.New()
	input.Status = StatusProses
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.repo.Create(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

func (s *Service) List(ctx context.Context) ([]SpmbRegistration, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) Get(ctx context.Context, id id.ID) (*SpmbRegistration, error) {
	r, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("registration not found")
	}
	return r, nil
}

func (s *Service) SetStatus(ctx context.Context, id id.ID, status, reason string) (*SpmbRegistration, error) {
	r, err := s.repo.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("registration not found")
	}

	expected := r.Status
	switch status {
	case StatusApprove:
		if r.Status != StatusProses {
			return nil, fmt.Errorf("only proses registrations can be approved")
		}
		r.Status = StatusApprove
	case StatusCancel:
		if r.Status != StatusProses && r.Status != StatusApprove {
			return nil, fmt.Errorf("only proses or approve registrations can be cancelled")
		}
		if reason == "" {
			return nil, fmt.Errorf("cancellation reason is required")
		}
		r.Status = StatusCancel
		r.CancelReason = reason
	default:
		return nil, fmt.Errorf("invalid status: must be approve or cancel")
	}

	r.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateStatus(ctx, r, expected); err != nil {
		return nil, err
	}
	return r, nil
}

func validateCreate(r *SpmbRegistration) error {
	if r.Nama == "" {
		return fmt.Errorf("nama is required")
	}
	if len(r.Nik) != 16 {
		return fmt.Errorf("nik must be 16 digits")
	}
	if r.JenisKelamin == "" {
		return fmt.Errorf("jenis_kelamin is required")
	}
	if r.Alamat == "" {
		return fmt.Errorf("alamat is required")
	}
	if r.Jurusan == "" {
		return fmt.Errorf("jurusan is required")
	}
	if !validJurusan(r.Jurusan) {
		return fmt.Errorf("invalid jurusan")
	}
	return nil
}

func validJurusan(j string) bool {
	switch j {
	case "PPLG", "Akuntansi", "Perhotelan", "Hotel":
		return true
	}
	return false
}

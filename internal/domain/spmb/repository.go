package spmb

import (
	"context"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Repository interface {
	Create(ctx context.Context, r *SpmbRegistration) error
	ByID(ctx context.Context, id id.ID) (*SpmbRegistration, error)
	ListAll(ctx context.Context) ([]SpmbRegistration, error)
	UpdateStatus(ctx context.Context, r *SpmbRegistration, expectedStatus string) error
}

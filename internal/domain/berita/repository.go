package berita

import (
	"context"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Repository interface {
	List(ctx context.Context) ([]Berita, error)
	ByID(ctx context.Context, id id.ID) (*Berita, error)
	Create(ctx context.Context, b *Berita) error
	Update(ctx context.Context, b *Berita) error
	Delete(ctx context.Context, id id.ID) error
}

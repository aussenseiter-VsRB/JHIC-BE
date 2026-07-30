package berita

import "context"

type Repository interface {
	List(ctx context.Context) ([]Berita, error)
	ByID(ctx context.Context, id string) (*Berita, error)
	Create(ctx context.Context, b *Berita) error
	Update(ctx context.Context, b *Berita) error
	Delete(ctx context.Context, id string) error
}
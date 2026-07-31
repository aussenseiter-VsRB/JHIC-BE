package user

import (
	"context"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Repository interface {
	List(ctx context.Context) ([]User, error)
	ByID(ctx context.Context, id id.ID) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User, passwordHash string) error
	FindByPosition(ctx context.Context, position, class, jurusan string) (*User, error)
	Update(ctx context.Context, u *User) error
	UpdateRole(ctx context.Context, id id.ID, role string) error
	Delete(ctx context.Context, id id.ID) error
}

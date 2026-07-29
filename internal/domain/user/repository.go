package user

import "context"

type Repository interface {
	List(ctx context.Context) ([]User, error)
	ByID(ctx context.Context, id string) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	UpdateRole(ctx context.Context, id, role string) error
	Delete(ctx context.Context, id string) error
}

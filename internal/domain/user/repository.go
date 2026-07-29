package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) error
	ByID(ctx context.Context, id string) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id string) error
}

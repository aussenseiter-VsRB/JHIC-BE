package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		expected := []user.User{{ID: "u1", Email: "a@example.com"}, {ID: "u2", Email: "b@example.com"}}
		repo.On("List", mock.Anything).Return(expected, nil)

		svc := user.NewService(repo)
		got, err := svc.List(context.Background())
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("List", mock.Anything).Return(nil, errors.New("db down"))

		svc := user.NewService(repo)
		_, err := svc.List(context.Background())
		require.EqualError(t, err, "db down")
	})
}

func TestService_ByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		expected := &user.User{ID: "u1", Email: "a@example.com"}
		repo.On("ByID", mock.Anything, "u1").Return(expected, nil)

		svc := user.NewService(repo)
		got, err := svc.ByID(context.Background(), "u1")
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*user.User)(nil), nil)

		svc := user.NewService(repo)
		got, err := svc.ByID(context.Background(), "missing")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "u1").Return((*user.User)(nil), errors.New("db down"))

		svc := user.NewService(repo)
		_, err := svc.ByID(context.Background(), "u1")
		require.EqualError(t, err, "db down")
	})
}

func TestService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &user.User{ID: "u1", Email: "a@example.com", Name: "Old", AvatarURL: ""}
		repo.On("ByID", mock.Anything, "u1").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
			return u.ID == "u1" && u.Name == "New" && u.AvatarURL == "https://cdn.example.com/a.png" && !u.UpdatedAt.IsZero()
		})).Return(nil)

		svc := user.NewService(repo)
		got, err := svc.Update(context.Background(), "u1", "New", "https://cdn.example.com/a.png")
		require.NoError(t, err)
		require.Equal(t, "New", got.Name)
		require.Equal(t, "https://cdn.example.com/a.png", got.AvatarURL)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*user.User)(nil), nil)

		svc := user.NewService(repo)
		_, err := svc.Update(context.Background(), "missing", "New", "")
		require.EqualError(t, err, "user not found")
	})

	t.Run("by id error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "u1").Return((*user.User)(nil), errors.New("db down"))

		svc := user.NewService(repo)
		_, err := svc.Update(context.Background(), "u1", "New", "")
		require.EqualError(t, err, "db down")
	})

	t.Run("update error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &user.User{ID: "u1", Email: "a@example.com"}
		repo.On("ByID", mock.Anything, "u1").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))

		svc := user.NewService(repo)
		_, err := svc.Update(context.Background(), "u1", "New", "")
		require.EqualError(t, err, "update failed")
	})
}

func TestService_UpdateRole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &user.User{ID: "u1", Role: "user"}
		repo.On("ByID", mock.Anything, "u1").Return(existing, nil)
		repo.On("UpdateRole", mock.Anything, "u1", "admin").Return(nil)

		svc := user.NewService(repo)
		require.NoError(t, svc.UpdateRole(context.Background(), "u1", "admin"))
	})

	t.Run("invalid role", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		svc := user.NewService(repo)
		err := svc.UpdateRole(context.Background(), "u1", "superuser")
		require.EqualError(t, err, "invalid role: must be one of [jurnal guru admin user]")
	})

	t.Run("user not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*user.User)(nil), nil)

		svc := user.NewService(repo)
		err := svc.UpdateRole(context.Background(), "missing", "admin")
		require.EqualError(t, err, "user not found")
	})

	t.Run("by id error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "u1").Return((*user.User)(nil), errors.New("db down"))

		svc := user.NewService(repo)
		err := svc.UpdateRole(context.Background(), "u1", "admin")
		require.EqualError(t, err, "db down")
	})

	t.Run("update role error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &user.User{ID: "u1", Role: "user"}
		repo.On("ByID", mock.Anything, "u1").Return(existing, nil)
		repo.On("UpdateRole", mock.Anything, "u1", "admin").Return(errors.New("update failed"))

		svc := user.NewService(repo)
		err := svc.UpdateRole(context.Background(), "u1", "admin")
		require.EqualError(t, err, "update failed")
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("Delete", mock.Anything, "u1").Return(nil)

		svc := user.NewService(repo)
		require.NoError(t, svc.Delete(context.Background(), "u1"))
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("Delete", mock.Anything, "u1").Return(errors.New("delete failed"))

		svc := user.NewService(repo)
		require.EqualError(t, svc.Delete(context.Background(), "u1"), "delete failed")
	})
}

package berita_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/berita"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/berita/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		expected := []berita.Berita{{ID: "b1", Title: "First"}, {ID: "b2", Title: "Second"}}
		repo.On("List", mock.Anything).Return(expected, nil)

		svc := berita.NewService(repo)
		got, err := svc.List(context.Background())
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("List", mock.Anything).Return(nil, errors.New("db down"))

		svc := berita.NewService(repo)
		_, err := svc.List(context.Background())
		require.EqualError(t, err, "db down")
	})
}

func TestService_ByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		expected := &berita.Berita{ID: "b1", Title: "First"}
		repo.On("ByID", mock.Anything, "b1").Return(expected, nil)

		svc := berita.NewService(repo)
		got, err := svc.ByID(context.Background(), "b1")
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*berita.Berita)(nil), nil)

		svc := berita.NewService(repo)
		got, err := svc.ByID(context.Background(), "missing")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "b1").Return((*berita.Berita)(nil), errors.New("db down"))

		svc := berita.NewService(repo)
		_, err := svc.ByID(context.Background(), "b1")
		require.EqualError(t, err, "db down")
	})
}

func TestService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("Create", mock.Anything, mock.MatchedBy(func(b *berita.Berita) bool {
			return b.ID != "" && b.AuthorID == "u1" && b.Title == "News" && b.Content == "Body" && !b.CreatedAt.IsZero() && !b.UpdatedAt.IsZero()
		})).Return(nil)

		svc := berita.NewService(repo)
		got, err := svc.Create(context.Background(), "u1", "News", "Body")
		require.NoError(t, err)
		require.NotEmpty(t, got.ID)
		require.Equal(t, "u1", got.AuthorID)
		require.Equal(t, "News", got.Title)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

		svc := berita.NewService(repo)
		_, err := svc.Create(context.Background(), "u1", "News", "Body")
		require.EqualError(t, err, "insert failed")
	})
}

func TestService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1", Title: "Old", Content: "Old body"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.MatchedBy(func(b *berita.Berita) bool {
			return b.ID == "b1" && b.Title == "New" && b.Content == "New body" && !b.UpdatedAt.IsZero()
		})).Return(nil)

		svc := berita.NewService(repo)
		got, err := svc.Update(context.Background(), "b1", "u1", "New", "New body")
		require.NoError(t, err)
		require.Equal(t, "New", got.Title)
	})

	t.Run("berita not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*berita.Berita)(nil), nil)

		svc := berita.NewService(repo)
		_, err := svc.Update(context.Background(), "missing", "u1", "New", "Body")
		require.EqualError(t, err, "berita not found")
	})

	t.Run("forbidden not author", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1", Title: "Old"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)

		svc := berita.NewService(repo)
		_, err := svc.Update(context.Background(), "b1", "intruder", "New", "Body")
		require.EqualError(t, err, "forbidden: not the author")
	})

	t.Run("by id error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "b1").Return((*berita.Berita)(nil), errors.New("db down"))

		svc := berita.NewService(repo)
		_, err := svc.Update(context.Background(), "b1", "u1", "New", "Body")
		require.EqualError(t, err, "db down")
	})

	t.Run("update error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))

		svc := berita.NewService(repo)
		_, err := svc.Update(context.Background(), "b1", "u1", "New", "Body")
		require.EqualError(t, err, "update failed")
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)
		repo.On("Delete", mock.Anything, "b1").Return(nil)

		svc := berita.NewService(repo)
		require.NoError(t, svc.Delete(context.Background(), "b1", "u1"))
	})

	t.Run("berita not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*berita.Berita)(nil), nil)

		svc := berita.NewService(repo)
		err := svc.Delete(context.Background(), "missing", "u1")
		require.EqualError(t, err, "berita not found")
	})

	t.Run("forbidden not author", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)

		svc := berita.NewService(repo)
		err := svc.Delete(context.Background(), "b1", "intruder")
		require.EqualError(t, err, "forbidden: not the author")
	})

	t.Run("by id error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "b1").Return((*berita.Berita)(nil), errors.New("db down"))

		svc := berita.NewService(repo)
		err := svc.Delete(context.Background(), "b1", "u1")
		require.EqualError(t, err, "db down")
	})

	t.Run("delete error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)
		repo.On("Delete", mock.Anything, "b1").Return(errors.New("delete failed"))

		svc := berita.NewService(repo)
		err := svc.Delete(context.Background(), "b1", "u1")
		require.EqualError(t, err, "delete failed")
	})
}

func TestService_SetImage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.MatchedBy(func(b *berita.Berita) bool {
			return b.ID == "b1" && b.ImageURL == "berita/b1/photo.png" && !b.UpdatedAt.IsZero()
		})).Return(nil)

		svc := berita.NewService(repo)
		got, err := svc.SetImage(context.Background(), "b1", "u1", "berita/b1/photo.png")
		require.NoError(t, err)
		require.Equal(t, "berita/b1/photo.png", got.ImageURL)
	})

	t.Run("berita not found", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		repo.On("ByID", mock.Anything, "missing").Return((*berita.Berita)(nil), nil)

		svc := berita.NewService(repo)
		_, err := svc.SetImage(context.Background(), "missing", "u1", "berita/x.png")
		require.EqualError(t, err, "berita not found")
	})

	t.Run("forbidden not author", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)

		svc := berita.NewService(repo)
		_, err := svc.SetImage(context.Background(), "b1", "intruder", "berita/b1/x.png")
		require.EqualError(t, err, "forbidden: not the author")
	})

	t.Run("update error", func(t *testing.T) {
		repo := mocks.NewRepository(t)

		existing := &berita.Berita{ID: "b1", AuthorID: "u1"}
		repo.On("ByID", mock.Anything, "b1").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))

		svc := berita.NewService(repo)
		_, err := svc.SetImage(context.Background(), "b1", "u1", "berita/b1/x.png")
		require.EqualError(t, err, "update failed")
	})
}

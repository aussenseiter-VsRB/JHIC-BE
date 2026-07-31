package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		users.On("ByEmail", mock.Anything, "u@example.com").Return((*auth.User)(nil), nil)
		users.On("Create", mock.Anything, mock.MatchedBy(func(u *auth.User) bool {
			return u.ID != "" && u.Email == "u@example.com" && u.Name == "U" && u.Role == "user" && u.PasswordHash != ""
		})).Return(nil)
		sessions.On("Create", mock.Anything,
			mock.MatchedBy(func(token string) bool { return len(token) == 64 }),
			mock.Anything,
			mock.MatchedBy(func(expiresAt int64) bool {
				return expiresAt > time.Now().Add(71*time.Hour).Unix() && expiresAt < time.Now().Add(73*time.Hour).Unix()
			}),
		).Return(nil)

		svc := auth.NewService(users, sessions)
		u, token, err := svc.Register(context.Background(), "u@example.com", "secret-password", "U")
		require.NoError(t, err)
		require.Equal(t, "user", u.Role)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("secret-password")))
		require.Len(t, token, 64)
	})

	t.Run("duplicate email", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		existing := &auth.User{ID: "u1", Email: "u@example.com"}
		users.On("ByEmail", mock.Anything, "u@example.com").Return(existing, nil)

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Register(context.Background(), "u@example.com", "secret", "U")
		require.EqualError(t, err, "user with email u@example.com already exists")
	})

	t.Run("by email error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		users.On("ByEmail", mock.Anything, "u@example.com").Return((*auth.User)(nil), errors.New("db down"))

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Register(context.Background(), "u@example.com", "secret", "U")
		require.EqualError(t, err, "db down")
	})

	t.Run("create error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		users.On("ByEmail", mock.Anything, mock.Anything).Return((*auth.User)(nil), nil)
		users.On("Create", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Register(context.Background(), "u@example.com", "secret", "U")
		require.EqualError(t, err, "insert failed")
	})

	t.Run("session create error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		users.On("ByEmail", mock.Anything, mock.Anything).Return((*auth.User)(nil), nil)
		users.On("Create", mock.Anything, mock.Anything).Return(nil)
		sessions.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("session failed"))

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Register(context.Background(), "u@example.com", "secret", "U")
		require.EqualError(t, err, "session failed")
	})
}

func TestService_Login(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret-password"), bcrypt.MinCost)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		existing := &auth.User{ID: "u1", Email: "u@example.com", PasswordHash: string(hash)}
		users.On("ByEmail", mock.Anything, "u@example.com").Return(existing, nil)
		sessions.On("Create", mock.Anything,
			mock.MatchedBy(func(token string) bool { return len(token) == 64 }),
			"u1",
			mock.Anything,
		).Return(nil)

		svc := auth.NewService(users, sessions)
		u, token, err := svc.Login(context.Background(), "u@example.com", "secret-password")
		require.NoError(t, err)
		require.Equal(t, "u1", u.ID)
		require.Len(t, token, 64)
	})

	t.Run("unknown email", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		users.On("ByEmail", mock.Anything, "nobody@example.com").Return((*auth.User)(nil), nil)

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Login(context.Background(), "nobody@example.com", "secret-password")
		require.EqualError(t, err, "invalid email or password")
	})

	t.Run("wrong password", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		existing := &auth.User{ID: "u1", Email: "u@example.com", PasswordHash: string(hash)}
		users.On("ByEmail", mock.Anything, "u@example.com").Return(existing, nil)

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Login(context.Background(), "u@example.com", "wrong-password")
		require.EqualError(t, err, "invalid email or password")
	})

	t.Run("by email error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		users.On("ByEmail", mock.Anything, mock.Anything).Return((*auth.User)(nil), errors.New("db down"))

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Login(context.Background(), "u@example.com", "secret-password")
		require.EqualError(t, err, "db down")
	})

	t.Run("session create error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		existing := &auth.User{ID: "u1", Email: "u@example.com", PasswordHash: string(hash)}
		users.On("ByEmail", mock.Anything, mock.Anything).Return(existing, nil)
		sessions.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("session failed"))

		svc := auth.NewService(users, sessions)
		_, _, err := svc.Login(context.Background(), "u@example.com", "secret-password")
		require.EqualError(t, err, "session failed")
	})
}

func TestService_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		sessions.On("DeleteByUserID", mock.Anything, "u1").Return(nil)

		svc := auth.NewService(users, sessions)
		require.NoError(t, svc.Logout(context.Background(), "u1"))
	})

	t.Run("delete error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		sessions.On("DeleteByUserID", mock.Anything, "u1").Return(errors.New("delete failed"))

		svc := auth.NewService(users, sessions)
		require.EqualError(t, svc.Logout(context.Background(), "u1"), "delete failed")
	})
}

func TestService_ValidateToken(t *testing.T) {
	t.Run("valid session", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		expected := &auth.Session{Token: "tok", UserID: "u1"}
		sessions.On("ByToken", mock.Anything, "tok").Return(expected, nil)

		svc := auth.NewService(users, sessions)
		got, err := svc.ValidateToken(context.Background(), "tok")
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("repository error", func(t *testing.T) {
		users := mocks.NewUserRepository(t)
		sessions := mocks.NewSessionRepository(t)

		sessions.On("ByToken", mock.Anything, "tok").Return((*auth.Session)(nil), errors.New("db down"))

		svc := auth.NewService(users, sessions)
		_, err := svc.ValidateToken(context.Background(), "tok")
		require.EqualError(t, err, "db down")
	})
}

func TestNewTokenValidator(t *testing.T) {
	t.Run("valid session returns user id", func(t *testing.T) {
		sessions := mocks.NewSessionRepository(t)

		sessions.On("ByToken", mock.Anything, "tok").Return(&auth.Session{UserID: "u1"}, nil)

		validate := auth.NewTokenValidator(sessions)
		userID, err := validate(context.Background(), "tok")
		require.NoError(t, err)
		require.Equal(t, "u1", userID)
	})

	t.Run("unknown session returns empty user id", func(t *testing.T) {
		sessions := mocks.NewSessionRepository(t)

		sessions.On("ByToken", mock.Anything, "tok").Return((*auth.Session)(nil), nil)

		validate := auth.NewTokenValidator(sessions)
		userID, err := validate(context.Background(), "tok")
		require.NoError(t, err)
		require.Empty(t, userID)
	})

	t.Run("repository error", func(t *testing.T) {
		sessions := mocks.NewSessionRepository(t)

		sessions.On("ByToken", mock.Anything, "tok").Return((*auth.Session)(nil), errors.New("db down"))

		validate := auth.NewTokenValidator(sessions)
		_, err := validate(context.Background(), "tok")
		require.EqualError(t, err, "db down")
	})
}

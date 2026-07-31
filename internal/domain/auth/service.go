package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users    UserRepository
	sessions SessionRepository
}

func NewService(users UserRepository, sessions SessionRepository) *Service {
	return &Service{users: users, sessions: sessions}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*User, string, error) {
	existing, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", fmt.Errorf("user with email %s already exists", email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &User{
		ID:           id.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, error) {
	user, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	token, err := s.generateToken(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) Logout(ctx context.Context, userID id.ID) error {
	return s.sessions.DeleteByUserID(ctx, userID)
}

func (s *Service) ValidateToken(ctx context.Context, token string) (*Session, error) {
	return s.sessions.ByToken(ctx, token)
}

type TokenValidator func(ctx context.Context, token string) (id.ID, error)

func NewTokenValidator(repo SessionRepository) TokenValidator {
	return func(ctx context.Context, token string) (id.ID, error) {
		session, err := repo.ByToken(ctx, token)
		if err != nil {
			return 0, err
		}
		if session == nil {
			return 0, nil
		}
		return session.UserID, nil
	}
}

func (s *Service) generateToken(ctx context.Context, userID id.ID) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	if err := s.sessions.Create(ctx, token, userID, time.Now().Add(72*time.Hour).Unix()); err != nil {
		return "", err
	}
	return token, nil
}

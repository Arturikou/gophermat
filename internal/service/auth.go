package service

import (
	"context"
	"fmt"

	"github.com/Arturikou/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepo interface {
	AddUser(ctx context.Context, login string, password string) (int, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
}

type AuthService struct {
	repo AuthRepo
}

func NewAuthService(
	repo AuthRepo,
) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) Register(ctx context.Context, login string, password string) (int, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return 0, fmt.Errorf("could not generate password: %w", err)
	}

	return s.repo.AddUser(ctx, login, hashedPassword)
}

func (s *AuthService) Login(ctx context.Context, login string, password string) (int, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return 0, fmt.Errorf("could not get user: %w", err)
	}

	err = verifyPassword(user.PasswordHash, password)
	if err != nil {
		return 0, fmt.Errorf("invalid password: %w", err)
	}

	return user.ID, nil
}

func hashPassword(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	return string(hashedPassword), nil
}

func verifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return models.ErrInvalidPassword
	}

	return nil
}

package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/Arturikou/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	type setup func(r *MockAuthRepo)

	tests := []struct {
		name          string
		login         string
		password      string
		setup         setup
		expectedID    int
		expectedError bool
	}{
		{
			name:     "Success registration",
			login:    "login",
			password: "password",
			setup: func(r *MockAuthRepo) {
				r.On("WithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					})

				r.On("AddUser", mock.Anything, "login", mock.AnythingOfType("string")).
					Return(1, nil)
				r.On("CreateUserBalance", mock.Anything, 1).
					Return(nil)
			},
			expectedID:    1,
			expectedError: false,
		},
		{
			name:     "AddUser fails",
			login:    "login",
			password: "password",
			setup: func(r *MockAuthRepo) {
				r.On("WithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(errors.New("db error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					})

				r.On("AddUser", mock.Anything, "login", mock.AnythingOfType("string")).
					Return(0, errors.New("db error"))
			},
			expectedID:    0,
			expectedError: true,
		},
		{
			name:     "Login already exists",
			login:    "login",
			password: "password",
			setup: func(r *MockAuthRepo) {
				r.On("WithTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(models.ErrLoginAlreadyExists).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					})

				r.On("AddUser", mock.Anything, "login", mock.AnythingOfType("string")).
					Return(0, models.ErrLoginAlreadyExists)
			},
			expectedID:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockAuthRepo(t)
			tt.setup(repo)
			service := NewAuthService(repo)

			id, err := service.Register(context.Background(), tt.login, tt.password)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	type setup func(r *MockAuthRepo, password string)

	tests := []struct {
		name          string
		login         string
		password      string
		setup         setup
		expectedID    int
		expectedError error
	}{
		{
			name:     "Success login",
			login:    "login",
			password: "password",
			setup: func(r *MockAuthRepo, password string) {
				hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				r.On("GetUserByLogin", mock.Anything, "login").
					Return(models.User{ID: 10, PasswordHash: string(hash)}, nil)
			},
			expectedID:    10,
			expectedError: nil,
		},
		{
			name:     "User not found",
			login:    "login",
			password: "password",
			setup: func(r *MockAuthRepo, password string) {
				r.On("GetUserByLogin", mock.Anything, "login").
					Return(models.User{}, models.ErrLoginNotFound)
			},
			expectedID:    0,
			expectedError: models.ErrLoginNotFound,
		},
		{
			name:     "Invalid password",
			login:    "login",
			password: "wrong_password",
			setup: func(r *MockAuthRepo, _ string) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
				r.On("GetUserByLogin", mock.Anything, "login").
					Return(models.User{ID: 10, PasswordHash: string(hash)}, nil)
			},
			expectedID:    0,
			expectedError: models.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockAuthRepo(t)
			tt.setup(repo, tt.password)
			service := NewAuthService(repo)

			id, err := service.Login(context.Background(), tt.login, tt.password)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

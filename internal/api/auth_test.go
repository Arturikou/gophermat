package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arturikou/internal/models"
	"github.com/go-playground/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Register(t *testing.T) {
	type setup func(auth *MockAuthService, tm *MockTokenManager)

	tests := []struct {
		name           string
		body           interface{}
		setup          setup
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "Success",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Register", mock.Anything, "login", "password").
					Return(1, nil)
				tm.On("BuildJWTString", 1).
					Return("valid-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "Bearer valid-token",
		},
		{
			name: "Empty Login",
			body: UserReq{Login: "", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Login Already Exists",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Register", mock.Anything, "login", "password").
					Return(0, models.ErrLoginAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Auth Service Fails",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Register", mock.Anything, "login", "password").
					Return(0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "BuildJWTString Fails",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Register", mock.Anything, "login", "password").
					Return(1, nil)
				tm.On("BuildJWTString", 1).
					Return("", errors.New("jwt error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := NewMockAuthService(t)
			mockTM := NewMockTokenManager(t)

			discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

			h := &Handler{
				logger:       discardLogger,
				authService:  mockAuth,
				tokenManager: mockTM,
				validate:     validator.New(),
			}

			tt.setup(mockAuth, mockTM)

			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/user/register", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			h.Register(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Authorization"))
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	type setup func(auth *MockAuthService, tm *MockTokenManager)

	tests := []struct {
		name           string
		body           interface{}
		setup          setup
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "Success",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Login", mock.Anything, "login", "password").
					Return(1, nil)
				tm.On("BuildJWTString", 1).
					Return("valid-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "Bearer valid-token",
		},
		{
			name: "Invalid Credentials",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Login", mock.Anything, "login", "password").
					Return(0, models.ErrInvalidPassword)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "User Not Found",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Login", mock.Anything, "login", "password").
					Return(0, models.ErrLoginNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Body",
			body:           "not-json",
			setup:          func(auth *MockAuthService, tm *MockTokenManager) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Auth Service Fails",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Login", mock.Anything, "login", "password").
					Return(0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "BuildJWTString Fails",
			body: UserReq{Login: "login", Password: "password"},
			setup: func(auth *MockAuthService, tm *MockTokenManager) {
				auth.On("Login", mock.Anything, "login", "password").
					Return(1, nil)
				tm.On("BuildJWTString", 1).
					Return("", errors.New("jwt error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := NewMockAuthService(t)
			mockTM := NewMockTokenManager(t)

			h := &Handler{
				logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
				authService:  mockAuth,
				tokenManager: mockTM,
				validate:     validator.New(),
			}

			tt.setup(mockAuth, mockTM)

			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			h.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Authorization"))
			}
		})
	}
}

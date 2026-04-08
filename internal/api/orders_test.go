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
	"time"

	"github.com/Arturikou/internal/ctxkeys"
	"github.com/Arturikou/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Orders(t *testing.T) {
	type setup func(os *MockOrderService)

	tests := []struct {
		name           string
		userID         int
		body           string
		setup          setup
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			body:   "12345678903",
			setup: func(os *MockOrderService) {
				os.On("AddOrder", mock.Anything, "12345678903", 1).
					Return(nil).Once()
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:   "Already Uploaded By Same User",
			userID: 1,
			body:   "12345678903",
			setup: func(os *MockOrderService) {
				os.On("AddOrder", mock.Anything, "12345678903", 1).
					Return(models.ErrOrderAlreadyExists).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Uploaded By Another User",
			userID: 1,
			body:   "12345678903",
			setup: func(os *MockOrderService) {
				os.On("AddOrder", mock.Anything, "12345678903", 1).
					Return(models.ErrOrderConflict).Once()
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Invalid Order Number",
			userID:         1,
			body:           "12345",
			setup:          func(os *MockOrderService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Empty Body",
			userID:         1,
			body:           "  ",
			setup:          func(os *MockOrderService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Internal Server Error",
			userID: 1,
			body:   "12345678903",
			setup: func(os *MockOrderService) {
				os.On("AddOrder", mock.Anything, "12345678903", 1).
					Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := NewMockOrderService(t)
			h := &Handler{
				logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
				orderService: os,
			}

			tt.setup(os)

			req := httptest.NewRequest(http.MethodPost, "/user/orders", bytes.NewBufferString(tt.body))
			req = req.WithContext(ctxkeys.SetUserID(req.Context(), tt.userID))
			w := httptest.NewRecorder()

			h.Orders(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			os.AssertExpectations(t)
		})
	}
}

func TestHandler_GetOrders(t *testing.T) {
	userID := 1

	t.Run("Success", func(t *testing.T) {
		os := NewMockOrderService(t)
		h := &Handler{
			logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
			orderService: os,
		}

		now := time.Now().Truncate(time.Second)
		accrualValue := decimal.NewFromFloat(500.5)

		orders := []models.Order{
			{
				Number:     "12345678903",
				Status:     models.OrderStatusProcessed,
				Accrual:    accrualValue,
				UploadedAt: now,
			},
		}

		os.On("GetUserOrders", mock.Anything, userID).Return(orders, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetOrders(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp []OrderResp
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "12345678903", resp[0].Number)
		assert.Equal(t, JSONDecimal{accrualValue}, *resp[0].Accrual)
		assert.Equal(t, now.Format(time.RFC3339), resp[0].UploadedAt)
		os.AssertExpectations(t)
	})

	t.Run("No Content", func(t *testing.T) {
		os := NewMockOrderService(t)
		h := &Handler{
			logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
			orderService: os,
		}

		os.On("GetUserOrders", mock.Anything, userID).Return([]models.Order{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetOrders(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		os.AssertExpectations(t)
	})

	t.Run("Internal Error", func(t *testing.T) {
		os := NewMockOrderService(t)
		h := &Handler{
			logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
			orderService: os,
		}

		os.On("GetUserOrders", mock.Anything, userID).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetOrders(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

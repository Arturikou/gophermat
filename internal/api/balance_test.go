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
	"github.com/go-playground/validator"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Withdraw(t *testing.T) {
	type setup func(bs *MockBalanceService)

	tests := []struct {
		name           string
		userID         int
		body           interface{}
		setup          setup
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			body:   WithdrawReq{OrderNumber: "12345678903", Amount: decimal.NewFromFloat(100.5)},
			setup: func(bs *MockBalanceService) {
				bs.On("Withdraw", mock.Anything, mock.MatchedBy(func(w models.Withdrawal) bool {
					return w.UserID == 1 && w.OrderNumber == "12345678903"
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Order Number",
			userID:         1,
			body:           WithdrawReq{OrderNumber: "123", Amount: decimal.NewFromInt(10)},
			setup:          func(bs *MockBalanceService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "Insufficient Funds",
			userID: 1,
			body:   WithdrawReq{OrderNumber: "12345678903", Amount: decimal.NewFromInt(1000)},
			setup: func(bs *MockBalanceService) {
				bs.On("Withdraw", mock.Anything, mock.Anything).Return(models.ErrInsufficientFunds).Once()
			},
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:   "Internal Server Error",
			userID: 1,
			body:   WithdrawReq{OrderNumber: "12345678903", Amount: decimal.NewFromInt(10)},
			setup: func(bs *MockBalanceService) {
				bs.On("Withdraw", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewMockBalanceService(t)
			h := &Handler{
				logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
				balanceService: bs,
				validate:       validator.New(),
			}

			tt.setup(bs)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/user/balance/withdraw", bytes.NewBuffer(body))
			req = req.WithContext(ctxkeys.SetUserID(req.Context(), tt.userID))

			w := httptest.NewRecorder()
			h.Withdraw(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			bs.AssertExpectations(t)
		})
	}
}

func TestHandler_GetBalance(t *testing.T) {
	userID := 1

	t.Run("Success", func(t *testing.T) {
		bs := NewMockBalanceService(t)
		h := &Handler{
			logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
			balanceService: bs,
		}

		bs.On("GetUserBalance", mock.Anything, userID).Return(models.Balance{
			Current:   decimal.NewFromFloat(500.5),
			Withdrawn: decimal.NewFromFloat(100.1),
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/user/balance", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetBalance(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp BalanceResp
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, 500.5, resp.Current)
		assert.Equal(t, 100.1, resp.Withdrawn)
	})

	t.Run("Internal Error", func(t *testing.T) {
		bs := NewMockBalanceService(t)
		h := &Handler{
			logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
			balanceService: bs,
		}

		bs.On("GetUserBalance", mock.Anything, mock.Anything).
			Return(models.Balance{}, errors.New("db error")).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/user/balance", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetBalance(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		bs.AssertExpectations(t)
	})
}

func TestHandler_GetWithdrawals(t *testing.T) {
	userID := 1

	t.Run("Success", func(t *testing.T) {
		bs := NewMockBalanceService(t)
		h := &Handler{
			logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
			balanceService: bs,
		}

		now := time.Now().Truncate(time.Second)
		data := []models.Withdrawal{
			{
				OrderNumber: "12345678903",
				Amount:      decimal.NewFromFloat(10.5),
				ProcessedAt: now,
			},
		}

		bs.On("GetWithdrawals", mock.Anything, userID).Return(data, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/user/withdrawals", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetWithdrawals(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []WithdrawalResp
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, now.Format(time.RFC3339), resp[0].ProcessedAt)
	})

	t.Run("No Content", func(t *testing.T) {
		bs := NewMockBalanceService(t)
		h := &Handler{
			logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
			balanceService: bs,
		}

		bs.On("GetWithdrawals", mock.Anything, mock.Anything).
			Return([]models.Withdrawal{}, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/user/withdrawals", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetWithdrawals(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		bs.AssertExpectations(t)
	})

	t.Run("Internal Error", func(t *testing.T) {
		bs := NewMockBalanceService(t)
		h := &Handler{
			logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
			balanceService: bs,
		}

		bs.On("GetWithdrawals", mock.Anything, userID).
			Return(nil, errors.New("db error")).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/user/withdrawals", nil)
		req = req.WithContext(ctxkeys.SetUserID(req.Context(), userID))
		w := httptest.NewRecorder()

		h.GetWithdrawals(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		bs.AssertExpectations(t)
	})
}

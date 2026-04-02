package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Arturikou/internal/logger"
	"github.com/Arturikou/internal/models"
	"github.com/Arturikou/internal/utils"
	"github.com/shopspring/decimal"
)

type WithdrawReq struct {
	OrderNumber string          `json:"order" validate:"required"`
	Amount      decimal.Decimal `json:"sum" validate:"required, gt=0"`
}

type BalanceResp struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type WithdrawalResp struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req WithdrawReq
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("failed to decode withdraw request", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = h.validate.Struct(req); err != nil {
		h.logger.Debug("validation error", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, ok := h.userIDFromCtx(ctx, w)
	if !ok {
		return
	}

	withdraw := models.Withdrawal{
		OrderNumber: req.OrderNumber,
		Amount:      req.Amount,
		UserID:      userID,
	}

	if !utils.IsValidNumber(withdraw.OrderNumber) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err = h.balanceService.Withdraw(ctx, withdraw)
	if err != nil {
		if errors.Is(err, models.ErrInsufficientFunds) {
			h.logger.Debug("insufficient funds", logger.Err(err))
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
			return
		}
		h.logger.Error("failed to withdraw", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := h.userIDFromCtx(ctx, w)
	if !ok {
		return
	}

	balance, err := h.balanceService.GetUserBalance(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get user balance", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := BalanceResp{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode balance", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := h.userIDFromCtx(ctx, w)
	if !ok {
		return
	}

	withdrawals, err := h.balanceService.GetWithdrawals(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get withdrawals", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var resp []WithdrawalResp
	for _, withdrawal := range withdrawals {
		resp = append(resp, WithdrawalResp{
			OrderNumber: withdrawal.OrderNumber,
			Amount:      withdrawal.Amount,
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode withdrawals", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

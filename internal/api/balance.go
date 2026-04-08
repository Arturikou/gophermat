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
	Amount      decimal.Decimal `json:"sum"`
}

type BalanceResp struct {
	Current   JSONDecimal `json:"current"`
	Withdrawn JSONDecimal `json:"withdrawn"`
}

type WithdrawalResp struct {
	OrderNumber string      `json:"order"`
	Amount      JSONDecimal `json:"sum"`
	ProcessedAt string      `json:"processed_at"`
}

// Withdraw godoc
// @Summary      Списание средств
// @Tags         balance
// @Accept       json
// @Param        input  body  WithdrawReq  true  "Номер заказа и сумма списания"
// @Success      200  {string}  string  "Успешное списание"
// @Failure      400  {string}  string  "Неверный формат запроса"
// @Failure      402  {string}  string  "Недостаточно средств"
// @Failure      422  {string}  string  "Неверный формат номера заказа"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /user/balance/withdraw [post]
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

// GetBalance godoc
// @Summary      Получение текущего баланса
// @Tags         balance
// @Produce      json
// @Success      200  {object}  BalanceResp  "Текущий баланс"
// @Failure      500  {string}  string       "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /user/balance [get]
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
		Current:   JSONDecimal{balance.Current},
		Withdrawn: JSONDecimal{balance.Withdrawn},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode balance", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// GetWithdrawals godoc
// @Summary      Получение информации о выводе средств
// @Tags         balance
// @Produce      json
// @Success      200  {array}   WithdrawalResp  "Список списаний"
// @Success      204  {string}  string          "Списаний нет"
// @Failure      500  {string}  string          "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /user/withdrawals [get]
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
			Amount:      JSONDecimal{withdrawal.Amount},
			ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode withdrawals", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

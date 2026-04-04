package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Arturikou/internal/logger"
	"github.com/Arturikou/internal/models"
	"github.com/Arturikou/internal/utils"
	"github.com/shopspring/decimal"
)

type OrderResp struct {
	Number     string           `json:"number"`
	Status     string           `json:"status"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt string           `json:"uploaded_at"`
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))

	if orderNumber == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !utils.IsValidNumber(orderNumber) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	userID, ok := h.userIDFromCtx(ctx, w)
	if !ok {
		return
	}

	err = h.orderService.AddOrder(ctx, orderNumber, userID)
	if err != nil {
		if errors.Is(err, models.ErrOrderConflict) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		if errors.Is(err, models.ErrOrderAlreadyExists) {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.logger.Error("failed to add order", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := h.userIDFromCtx(ctx, w)
	if !ok {
		return
	}

	orders, err := h.orderService.GetUserOrders(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get orders for user", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ordersResp := make([]OrderResp, 0, len(orders))
	for _, order := range orders {
		resp := OrderResp{
			Number:     order.Number,
			Status:     string(order.Status),
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		if !order.Accrual.IsZero() {
			resp.Accrual = &order.Accrual
		}
		ordersResp = append(ordersResp, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ordersResp); err != nil {
		h.logger.Error("failed to encode orders", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

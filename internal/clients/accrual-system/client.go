package accrualsystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type OrderResponse struct {
	Order   string           `json:"order"`
	Status  string           `json:"status"`
	Accrual *decimal.Decimal `json:"accrual,omitempty"`
}

type ErrRateLimited struct {
	RetryAfter int
}

var ErrOrderNotRegistered = errors.New("order not registered in accrual system")

type Client struct {
	httpClient http.Client
	address    string
}

func New(
	address string,
	timeout time.Duration,
) *Client {
	return &Client{
		address: address,
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetOrder(ctx context.Context, orderNumber string) (OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.address, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("accrual get order: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("accrual get order: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var res OrderResponse
		if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return OrderResponse{}, fmt.Errorf("accrual get order: %w", err)
		}
		return res, nil

	case http.StatusNoContent:
		return OrderResponse{}, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return OrderResponse{}, &ErrRateLimited{RetryAfter: retryAfter}

	default:
		return OrderResponse{}, fmt.Errorf("unexpected error, status code: %d", resp.StatusCode)
	}
}

func (e *ErrRateLimited) Error() string {
	return fmt.Sprintf("too many requests, retry after %d seconds", e.RetryAfter)
}

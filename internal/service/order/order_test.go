package order

import (
	"context"
	"errors"
	"testing"

	"github.com/Arturikou/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestOrderService_AddOrder_TableDriven(t *testing.T) {
	type setup func(r *MockOrderRepo, ctx context.Context, orderNum string, userID int)

	tests := []struct {
		name          string
		orderNumber   string
		userID        int
		setup         setup
		expectedError error
	}{
		{
			name:        "Success",
			orderNumber: "123456",
			userID:      1,
			setup: func(r *MockOrderRepo, ctx context.Context, orderNum string, userID int) {
				r.On("AddOrder", ctx, orderNum, userID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "Already exists by same user",
			orderNumber: "123456",
			userID:      1,
			setup: func(r *MockOrderRepo, ctx context.Context, orderNum string, userID int) {
				r.On("AddOrder", ctx, orderNum, userID).Return(models.ErrDuplicateOrder)
				r.On("GetUserIDByOrderNumber", ctx, orderNum).Return(1, nil)
			},
			expectedError: models.ErrOrderAlreadyExists,
		},
		{
			name:        "Owned by another user",
			orderNumber: "123456",
			userID:      1,
			setup: func(r *MockOrderRepo, ctx context.Context, orderNum string, userID int) {
				r.On("AddOrder", ctx, orderNum, userID).Return(models.ErrDuplicateOrder)
				r.On("GetUserIDByOrderNumber", ctx, orderNum).Return(2, nil)
			},
			expectedError: models.ErrOrderConflict,
		},
		{
			name:        "DB error on GetUserID",
			orderNumber: "123456",
			userID:      1,
			setup: func(r *MockOrderRepo, ctx context.Context, orderNum string, userID int) {
				r.On("AddOrder", ctx, orderNum, userID).Return(models.ErrDuplicateOrder)
				r.On("GetUserIDByOrderNumber", ctx, orderNum).Return(0, errors.New("db fail"))
			},
			expectedError: errors.New("db fail"),
		},
		{
			name:        "DB error on AddOrder",
			orderNumber: "123456",
			userID:      1,
			setup: func(r *MockOrderRepo, ctx context.Context, orderNum string, userID int) {
				r.On("AddOrder", ctx, orderNum, userID).Return(errors.New("fatal error"))
			},
			expectedError: errors.New("fatal error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := NewMockOrderRepo(t)
			tt.setup(repo, ctx, tt.orderNumber, tt.userID)
			service := NewOrderService(repo)

			err := service.AddOrder(ctx, tt.orderNumber, tt.userID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderService_GetUserOrders(t *testing.T) {
	ctx := context.Background()
	userID := 1
	expected := []models.Order{{Number: "123456"}}

	repo := NewMockOrderRepo(t)
	service := NewOrderService(repo)

	repo.
		On("GetOrdersByUserID", ctx, userID).
		Return(expected, nil)

	res, err := service.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestOrderService_GetPendingOrders(t *testing.T) {
	ctx := context.Background()
	expected := []models.OrderUpdate{{Number: "123456", Status: "PROCESSING"}}

	repo := NewMockOrderRepo(t)
	service := NewOrderService(repo)

	repo.
		On("GetOrdersInStatus", ctx, models.PendingOrderStatuses).
		Return(expected, nil)

	res, err := service.GetPendingOrders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestOrderService_UpdateOrders(t *testing.T) {
	ctx := context.Background()
	updates := []models.OrderUpdate{{Number: "123456"}}

	repo := NewMockOrderRepo(t)
	service := NewOrderService(repo)

	repo.
		On("UpdateOrders", ctx, updates).
		Return(nil)

	err := service.UpdateOrders(ctx, updates)
	assert.NoError(t, err)
}

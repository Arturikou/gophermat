package service

import (
	"context"
	"errors"

	"github.com/Arturikou/internal/models"
)

type OrderRepo interface {
	AddOrder(ctx context.Context, orderNumber string, userID int) error
	GetUserIDByOrderNumber(ctx context.Context, number string) (int, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)
	GetOrdersInStatus(ctx context.Context, status []models.OrderStatus) ([]models.OrderUpdate, error)
	UpdateOrders(ctx context.Context, orders []models.OrderUpdate) error
}

type OrderService struct {
	repo OrderRepo
}

func NewOrderService(repo OrderRepo) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (s *OrderService) AddOrder(ctx context.Context, orderNumber string, userID int) error {
	err := s.repo.AddOrder(ctx, orderNumber, userID)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateOrder) {
			ownerID, err := s.repo.GetUserIDByOrderNumber(ctx, orderNumber)
			if err != nil {
				return err
			}

			if ownerID == userID {
				return models.ErrOrderAlreadyExists
			}

			return models.ErrOrderConflict
		}

		return err
	}

	return nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	return s.repo.GetOrdersByUserID(ctx, userID)
}

func (s *OrderService) GetPendingOrders(ctx context.Context) ([]models.OrderUpdate, error) {
	return s.repo.GetOrdersInStatus(ctx, models.PendingOrderStatuses)
}

func (s *OrderService) UpdateOrders(ctx context.Context, orders []models.OrderUpdate) error {
	return s.repo.UpdateOrders(ctx, orders)
}

package accrual

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"sync"
	"time"

	accrualsystem "github.com/Arturikou/internal/clients/accrual-system"
	"github.com/Arturikou/internal/logger"
	"github.com/Arturikou/internal/models"
	"github.com/shopspring/decimal"
)

//go:generate mockery --name=OrderService --filename=mock_order_service_test.go --inpackage --disable-version-string
type OrderService interface {
	GetPendingOrders(ctx context.Context) ([]models.OrderUpdate, error)
	UpdateOrders(ctx context.Context, orders []models.OrderUpdate) error
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

//go:generate mockery --name=BalanceService --filename=mock_balance_service_test.go --inpackage --disable-version-string
type BalanceService interface {
	UpdateUserBalance(ctx context.Context, userID int, amount decimal.Decimal) error
}

//go:generate mockery --name=Client --filename=mock_client_test.go --inpackage --disable-version-string
type Client interface {
	GetOrder(ctx context.Context, orderNumber string) (accrualsystem.OrderResponse, error)
}

const (
	generatorInterval = 10 * time.Second
	flusherInterval   = 5 * time.Second
	batchSize         = 50
)

type Processor struct {
	logger         *slog.Logger
	orderService   OrderService
	balanceService BalanceService
	accrualClient  Client
	orderChan      chan models.OrderUpdate
	processedChan  chan models.OrderUpdate
	semaphore      chan struct{}
	workerCount    int
	wg             sync.WaitGroup
	mu             sync.RWMutex
	retryUntil     time.Time
}

func NewProcessor(
	log *slog.Logger,
	workerCount,
	maxConcurrency int,
	orderService OrderService,
	accrualClient Client,
	balanceService BalanceService,
) *Processor {
	return &Processor{
		logger:         log,
		orderChan:      make(chan models.OrderUpdate, 1024),
		processedChan:  make(chan models.OrderUpdate, 1024),
		semaphore:      make(chan struct{}, maxConcurrency),
		workerCount:    workerCount,
		orderService:   orderService,
		accrualClient:  accrualClient,
		balanceService: balanceService,
	}
}

func (p *Processor) Run(ctx context.Context) {
	p.wg.Add(1)
	go p.runGenerator(ctx)

	for range p.workerCount {
		p.wg.Add(1)
		go p.processOrder(ctx)
	}

	p.wg.Add(1)
	go p.runFlusher(ctx)

	p.wg.Wait()
}

func (p *Processor) runGenerator(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(generatorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for order := range p.pendingOrders(ctx) {
				select {
				case p.orderChan <- order:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (p *Processor) pendingOrders(ctx context.Context) iter.Seq[models.OrderUpdate] {
	return func(yield func(models.OrderUpdate) bool) {
		orders, err := p.orderService.GetPendingOrders(ctx)
		if err != nil {
			p.logger.Error("failed to get pending orders", logger.Err(err))
			return
		}

		for _, order := range orders {
			if !yield(order) {
				return
			}
		}
	}
}

func (p *Processor) processOrder(ctx context.Context) {
	defer p.wg.Done()

	for {
		p.mu.RLock()
		pause := time.Until(p.retryUntil)
		p.mu.RUnlock()

		if pause > 0 {
			p.logger.Debug("rate limit pause", "duration", pause)
			select {
			case <-time.After(pause):
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		case order := <-p.orderChan:
			result, err := p.pollAccrual(ctx, order)
			if err != nil {
				p.logger.Error("failed to poll accrual", "order", order.Number, logger.Err(err))
				continue
			}
			select {
			case p.processedChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (p *Processor) pollAccrual(ctx context.Context, order models.OrderUpdate) (models.OrderUpdate, error) {
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
	case <-ctx.Done():
		return models.OrderUpdate{}, ctx.Err()
	}

	resp, err := p.accrualClient.GetOrder(ctx, order.Number)
	if err != nil {
		if errRateLimited, ok := errors.AsType[*accrualsystem.ErrRateLimited](err); ok {
			p.mu.Lock()
			p.retryUntil = time.Now().Add(time.Duration(errRateLimited.RetryAfter) * time.Second)
			p.mu.Unlock()

			return models.OrderUpdate{}, fmt.Errorf("rate limit hit, pausing for %ds", errRateLimited.RetryAfter)
		}

		if errors.Is(err, accrualsystem.ErrOrderNotRegistered) {
			return models.OrderUpdate{
				Number: order.Number,
				UserID: order.UserID,
				Status: models.OrderStatusInvalid,
			}, nil
		}

		return models.OrderUpdate{}, err
	}

	accrual := decimal.Zero
	if resp.Accrual != nil {
		accrual = *resp.Accrual
	}

	return models.OrderUpdate{
		Number:  order.Number,
		UserID:  order.UserID,
		Status:  models.OrderStatus(resp.Status),
		Accrual: accrual,
	}, nil
}

func (p *Processor) runFlusher(ctx context.Context) {
	defer p.wg.Done()
	ticker := time.NewTicker(flusherInterval)
	defer ticker.Stop()

	buffer := make([]models.OrderUpdate, 0, batchSize)

	for {
		select {
		case res := <-p.processedChan:
			buffer = append(buffer, res)
			if len(buffer) >= batchSize {
				p.flush(ctx, buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				p.flush(ctx, buffer)
				buffer = buffer[:0]
			}
		case <-ctx.Done():
			if len(buffer) > 0 {
				p.flush(context.Background(), buffer)
			}
			return
		}
	}
}

func (p *Processor) flush(ctx context.Context, orders []models.OrderUpdate) {
	err := p.orderService.WithTx(ctx, func(ctx context.Context) error {
		if err := p.orderService.UpdateOrders(ctx, orders); err != nil {
			return err
		}

		for _, order := range orders {
			if order.Status == models.OrderStatusProcessed && !order.Accrual.IsZero() {
				if err := p.balanceService.UpdateUserBalance(ctx, order.UserID, order.Accrual); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		p.logger.Error("failed to flush orders", logger.Err(err))
	}
}

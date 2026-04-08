package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arturikou/internal/api"
	"github.com/Arturikou/internal/auth"
	accrualsystem "github.com/Arturikou/internal/clients/accrual-system"
	"github.com/Arturikou/internal/config"
	"github.com/Arturikou/internal/logger"
	"github.com/Arturikou/internal/repository"
	auth2 "github.com/Arturikou/internal/service/auth"
	"github.com/Arturikou/internal/service/balance"
	"github.com/Arturikou/internal/service/order"
	"github.com/Arturikou/internal/storage/postgresql"
	"github.com/Arturikou/internal/worker/accrual"
	"github.com/Arturikou/migrations"

	_ "github.com/Arturikou/docs"
)

// @title        Gophermart API
// @version      1.0
// @host         localhost:8080
// @BasePath     /api
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.MustLoad()
	log := logger.New(cfg.Env, cfg.LogLevel)

	if err := postgresql.RunMigrations(cfg.DB.DSN, migrations.FS); err != nil {
		log.Error("failed to run migrations", logger.Err(err))
		os.Exit(1)
	}

	pool, err := postgresql.New(ctx, postgresql.Config{
		DSN:             cfg.DB.DSN,
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.DB.ConnMaxIdleTime,
	})
	if err != nil {
		log.Error("failed to create storage", logger.Err(err))
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.New(pool)
	authService := auth2.NewAuthService(repo)
	tokenManager := auth.NewManager(cfg.Auth.SecretKey)
	balanceService := balance.NewBalanceService(repo)
	orderService := order.NewOrderService(repo)
	accrualClient := accrualsystem.New(cfg.AccrualSystem.Address, cfg.AccrualSystem.Timeout)

	accrualProcessor := accrual.NewProcessor(
		log,
		cfg.AccrualWorker.WorkerCount,
		cfg.AccrualWorker.MaxConcurrency,
		orderService,
		accrualClient,
		balanceService,
	)

	go accrualProcessor.Run(ctx)

	handler := api.New(
		log,
		authService,
		tokenManager,
		balanceService,
		orderService,
	)

	srv := http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      handler.Router(),
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("starting server", "address", cfg.HTTPServer.Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", logger.Err(err))
			cancel()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("failed to shutdown server", logger.Err(err))
	}
}

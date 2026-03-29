package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/Arturikou/internal/api"
	"github.com/Arturikou/internal/auth"
	"github.com/Arturikou/internal/config"
	"github.com/Arturikou/internal/logger"
	"github.com/Arturikou/internal/repository"
	"github.com/Arturikou/internal/service"
	"github.com/Arturikou/internal/storage/postgresql"
	"github.com/Arturikou/migrations"
)

func main() {
	ctx := context.Background()
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
	authService := service.NewAuthService(repo)
	tokenManager := auth.NewManager(cfg.Auth.SecretKey)

	handler := api.New(
		log,
		authService,
		tokenManager,
	)

	srv := http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      handler.Router(),
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info("starting server")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("operation failed", logger.Err(err))
		os.Exit(1)
	}
}

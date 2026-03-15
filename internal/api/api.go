package api

import (
	"log/slog"
)

type Handler struct {
	logger *slog.Logger
}

func New(
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger: logger,
	}
}

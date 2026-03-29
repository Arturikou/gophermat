package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Arturikou/internal/logger"
	"github.com/Arturikou/internal/models"
)

type UserReq struct {
	Login    string `json:"login" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=7"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req UserReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("failed to decode request", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = h.validate.Struct(req); err != nil {
		h.logger.Debug("validation error", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrLoginAlreadyExists) {
			h.logger.Debug("login already exists", logger.Err(err))
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		h.logger.Error("failed to register user", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := h.tokenManager.BuildJWTString(userID)
	if err != nil {
		h.logger.Error("failed to generate token", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req UserReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("failed to decode request", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = h.validate.Struct(req); err != nil {
		h.logger.Debug("validation error", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userID, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrLoginNotFound) || errors.Is(err, models.ErrInvalidPassword) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		h.logger.Error("failed to login", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := h.tokenManager.BuildJWTString(userID)
	if err != nil {
		h.logger.Error("failed to generate token", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

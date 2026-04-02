package api

import (
	"context"
	"net/http"

	"github.com/Arturikou/internal/ctxkeys"
	"github.com/Arturikou/internal/logger"
)

func (h *Handler) userIDFromCtx(ctx context.Context, w http.ResponseWriter) (int, bool) {
	userID, err := ctxkeys.UserIDFromContext(ctx)
	if err != nil {
		h.logger.Error("userID not found in context", logger.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return 0, false
	}
	return userID, true
}

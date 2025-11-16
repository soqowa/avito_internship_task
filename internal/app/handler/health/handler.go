package health

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/user/reviewer-svc/internal/app/httpserver"
)

type Handler struct {
	log *slog.Logger
	db  DBPinger
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

func NewHandler(log *slog.Logger, db DBPinger) *Handler {
	return &Handler{log: log, db: db}
}


// @Summary     Liveness probe
// @Tags        health
// @Success     200
// @Router      /healthz [get]
func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	httpserver.WriteNoContent(w)
}


// @Summary     Readiness probe
// @Tags        health
// @Success     200
// @Failure     503 {object} httpserver.ErrorResponse
// @Router      /readyz [get]
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		h.log.Error("readyz: database not ready", "err", err)
		httpserver.WriteError(w, http.StatusServiceUnavailable, "not_ready", "database not ready", nil)
		return
	}

	httpserver.WriteNoContent(w)
}

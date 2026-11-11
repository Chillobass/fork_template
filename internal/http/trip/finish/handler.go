package finish

import (
	"context"
	"log/slog"
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/respond"
	"github.com/Chillobass/fork_template/internal/http/trip"
)

type Handler struct {
	usecase usecase
	logger  *slog.Logger
	count   func(context.Context)
}

func New(uc usecase, logger *slog.Logger, count func(context.Context)) *Handler {
	return &Handler{usecase: uc, logger: logger, count: count}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	ctx := r.Context()
	finished, err := h.usecase.Finish(ctx, id)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	h.count(ctx)
	h.logger.InfoContext(ctx, "trip completed", "trip_id", id)
	respond.JSON(w, http.StatusOK, trip.Response(finished))
}

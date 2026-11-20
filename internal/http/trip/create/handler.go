package create

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

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	input, err := decodeCreateTripRequest(w, r)
	if err != nil {
		domainerror.InvalidRequest(w, r, err.Error())
		return
	}

	ctx := r.Context()
	created, err := h.usecase.Create(ctx, input)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	h.count(ctx)
	h.logger.InfoContext(ctx, "trip created", "trip_id", created.ID)
	w.Header().Set("Location", "/api/v1/trips/"+created.ID.String())
	respond.JSON(w, http.StatusCreated, trip.Response(created))
}

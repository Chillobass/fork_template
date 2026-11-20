package add

import (
	"context"
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/position"
	"github.com/Chillobass/fork_template/internal/http/respond"
)

type Handler struct {
	usecase   usecase
	canceller requestCanceller
	count     func(context.Context)
}

func New(uc usecase, canceller requestCanceller, count func(context.Context)) *Handler {
	return &Handler{usecase: uc, canceller: canceller, count: count}
}

func (h *Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request, id api.TripId) {
	input, err := decodeAddPositionRequest(w, r)
	if err != nil {
		domainerror.InvalidRequest(w, r, err.Error())
		return
	}

	ctx := r.Context()
	saved, err := h.usecase.Add(ctx, id, input)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	h.canceller.Cancel(ctx, id)
	h.count(ctx)
	respond.JSON(w, http.StatusCreated, position.Response(saved))
}

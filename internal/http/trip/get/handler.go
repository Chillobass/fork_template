package get

import (
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/respond"
	"github.com/Chillobass/fork_template/internal/http/trip"
)

type Handler struct {
	usecase usecase
}

func New(uc usecase) *Handler {
	return &Handler{usecase: uc}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	found, err := h.usecase.Get(r.Context(), id)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, trip.Response(found))
}

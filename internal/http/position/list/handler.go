package list

import (
	"net/http"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/position"
	"github.com/Chillobass/fork_template/internal/http/respond"
)

type Handler struct {
	usecase usecase
}

func New(uc usecase) *Handler {
	return &Handler{usecase: uc}
}

func (h *Handler) ListTripPositions(w http.ResponseWriter, r *http.Request, id api.TripId) {
	items, err := h.usecase.List(r.Context(), id)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, position.ListResponse(items))
}

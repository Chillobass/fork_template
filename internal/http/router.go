package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/health"
	"github.com/Chillobass/fork_template/internal/http/middleware"
	"github.com/Chillobass/fork_template/internal/http/trip/create"
	"github.com/Chillobass/fork_template/internal/http/trip/finish"
	"github.com/Chillobass/fork_template/internal/http/trip/get"
)

type (
	CreateHandler = create.Handler
	GetHandler    = get.Handler
	FinishHandler = finish.Handler
	HealthHandler = health.Handler
)

type Server struct {
	*CreateHandler
	*GetHandler
	*FinishHandler
	*HealthHandler
}

func NewRouter(server api.ServerInterface, metrics *middleware.HTTP) http.Handler {
	router := chi.NewRouter()
	api.HandlerWithOptions(server, api.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			detail := "invalid request parameters"
			var formatErr *api.InvalidParamFormatError
			if errors.As(err, &formatErr) {
				detail = fmt.Sprintf("parameter %s has invalid format", formatErr.ParamName)
			}
			domainerror.InvalidRequest(w, r, detail)
		},
	})
	return metrics.Wrap(router)
}

package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"

	api "github.com/Chillobass/fork_template/internal/generated"
)

func NewRouter(h *Handler, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(recoverPanic(logger))

	return api.HandlerWithOptions(h, api.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			detail := "invalid request parameters"

			var formatErr *api.InvalidParamFormatError
			if errors.As(err, &formatErr) {
				detail = fmt.Sprintf("parameter %s has invalid format", formatErr.ParamName)
			}

			writeProblem(w, r, problemInvalidRequest, detail)
		},
	})
}

func recoverPanic(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				p := recover()
				if p == nil {
					return
				}
				if p == http.ErrAbortHandler {
					panic(p)
				}

				logger.ErrorContext(r.Context(), "panic in http handler",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", p,
					"stack", string(debug.Stack()),
				)
				writeProblem(w, r, problemInternal, "")
			}()

			next.ServeHTTP(w, r)
		})
	}
}

package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Chillobass/fork_template/internal/domain"
	api "github.com/Chillobass/fork_template/internal/generated"
)

const (
	contentTypeJSON    = "application/json"
	contentTypeProblem = "application/problem+json"

	problemTypeBaseURI = "https://tripgo.example/problems/"
)

type problemKind struct {
	status        int
	code          string
	slug          string
	title         string
	defaultDetail string
}

var (
	problemInvalidRequest = problemKind{
		status:        http.StatusBadRequest,
		code:          "invalid_request",
		slug:          "invalid-request",
		title:         "Invalid request",
		defaultDetail: "Request validation failed",
	}
	problemTripNotFound = problemKind{
		status:        http.StatusNotFound,
		code:          "trip_not_found",
		slug:          "trip-not-found",
		title:         "Trip not found",
		defaultDetail: "Trip was not found",
	}
	problemTripCompleted = problemKind{
		status:        http.StatusConflict,
		code:          "trip_completed",
		slug:          "trip-completed",
		title:         "Trip completed",
		defaultDetail: "Operation is not allowed for a completed trip",
	}
	problemDriverBusy = problemKind{
		status:        http.StatusConflict,
		code:          "driver_busy",
		slug:          "driver-busy",
		title:         "Driver busy",
		defaultDetail: "Driver already has an active trip",
	}
	problemInternal = problemKind{
		status:        http.StatusInternalServerError,
		code:          "internal_error",
		slug:          "internal-error",
		title:         "Internal Server Error",
		defaultDetail: "Internal server error",
	}
)

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrTripNotFound):
		writeProblem(w, r, problemTripNotFound, "")
	case errors.Is(err, domain.ErrTripCompleted):
		writeProblem(w, r, problemTripCompleted, "")
	case errors.Is(err, domain.ErrDriverBusy):
		writeProblem(w, r, problemDriverBusy, "")
	default:
		h.logger.ErrorContext(r.Context(), "request failed",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
		writeProblem(w, r, problemInternal, "")
	}
}

func writeProblem(w http.ResponseWriter, r *http.Request, kind problemKind, detail string) {
	if detail == "" {
		detail = kind.defaultDetail
	}
	instance := r.URL.Path

	writeBody(w, kind.status, contentTypeProblem, api.Problem{
		Type:     problemTypeBaseURI + kind.slug,
		Title:    kind.title,
		Status:   int32(kind.status),
		Detail:   &detail,
		Instance: &instance,
		Code:     kind.code,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	writeBody(w, status, contentTypeJSON, body)
}

func writeBody(w http.ResponseWriter, status int, contentType string, body any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(body)
}

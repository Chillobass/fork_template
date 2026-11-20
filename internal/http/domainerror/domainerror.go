package domainerror

import (
	"context"
	"errors"
	"net/http"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/http/respond"
)

const (
	contentTypeProblem = "application/problem+json"

	problemTypeBaseURI = "https://tripgo.example/problems/"
)

type problemKind struct {
	status        int32
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

func Problem(err error) (int, api.Problem) {
	var validation *trip.ValidationError
	switch {
	case errors.As(err, &validation):
		return problemInvalidRequest.problem(validation.Error())
	case errors.Is(err, trip.ErrTripNotFound):
		return problemTripNotFound.problem("")
	case errors.Is(err, trip.ErrTripCompleted):
		return problemTripCompleted.problem("")
	case errors.Is(err, trip.ErrDriverBusy):
		return problemDriverBusy.problem("")
	default:
		return problemInternal.problem("")
	}
}

type reportKey struct{}

// WithReport gives request logging access to the original internal error
func WithReport(ctx context.Context, report func(error)) context.Context {
	return context.WithValue(ctx, reportKey{}, report)
}

func Write(w http.ResponseWriter, r *http.Request, err error) {
	status, body := Problem(err)
	if status == http.StatusInternalServerError {
		if report, ok := r.Context().Value(reportKey{}).(func(error)); ok {
			report(err)
		}
	}
	write(w, r, status, body)
}

func InvalidRequest(w http.ResponseWriter, r *http.Request, detail string) {
	status, body := problemInvalidRequest.problem(detail)
	write(w, r, status, body)
}

func (k problemKind) problem(detail string) (int, api.Problem) {
	if detail == "" {
		detail = k.defaultDetail
	}
	return int(k.status), api.Problem{
		Type:   problemTypeBaseURI + k.slug,
		Title:  k.title,
		Status: k.status,
		Detail: &detail,
		Code:   k.code,
	}
}

func write(w http.ResponseWriter, r *http.Request, status int, body api.Problem) {
	instance := r.URL.Path
	body.Instance = &instance
	respond.Write(w, status, contentTypeProblem, body)
}

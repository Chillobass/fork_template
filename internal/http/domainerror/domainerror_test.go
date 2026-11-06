package domainerror_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	"github.com/Chillobass/fork_template/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	const private = "failed SQL query with confidential data"
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{trip.ErrTripNotFound, http.StatusNotFound, "trip_not_found"},
		{trip.ErrTripCompleted, http.StatusConflict, "trip_completed"},
		{trip.ErrDriverBusy, http.StatusConflict, "driver_busy"},
		{&trip.ValidationError{Field: "price", Detail: "must be greater than or equal to 0"}, http.StatusBadRequest, "invalid_request"},
		{errors.New(private), http.StatusInternalServerError, "internal_error"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			status, body := domainerror.Problem(fmt.Errorf("usecase: %w", tc.err))
			if status != tc.status || int(body.Status) != status || body.Code != tc.code || body.Detail == nil ||
				body.Title == "" || body.Type == "" || strings.Contains(*body.Detail, private) {
				t.Fatalf("problem = (%d, %+v)", status, body)
			}
			if status == http.StatusBadRequest && !strings.Contains(*body.Detail, "price") {
				t.Fatalf("validation detail = %q", *body.Detail)
			}
		})
	}
}

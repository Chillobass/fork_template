package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/oapi-codegen/nullable"

	"github.com/Chillobass/fork_template/internal/domain"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/service"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	trips        *service.TripService
	db           Pinger
	readyTimeout time.Duration
	logger       *slog.Logger
}

var _ api.ServerInterface = (*Handler)(nil)

func NewHandler(trips *service.TripService, db Pinger, readyTimeout time.Duration, logger *slog.Logger) *Handler {
	return &Handler{
		trips:        trips,
		db:           db,
		readyTimeout: readyTimeout,
		logger:       logger,
	}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	input, err := decodeCreateTripRequest(w, r)
	if err != nil {
		writeProblem(w, r, problemInvalidRequest, err.Error())
		return
	}

	trip, err := h.trips.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	h.logger.InfoContext(r.Context(), "trip created", "trip_id", trip.ID)

	w.Header().Set("Location", "/api/v1/trips/"+trip.ID.String())
	writeJSON(w, http.StatusCreated, toAPITrip(trip))
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripID api.TripId) {
	trip, err := h.trips.Get(r.Context(), tripID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toAPITrip(trip))
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripID api.TripId) {
	trip, err := h.trips.Finish(r.Context(), tripID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	h.logger.InfoContext(r.Context(), "trip finished", "trip_id", trip.ID)

	writeJSON(w, http.StatusOK, toAPITrip(trip))
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.readyTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		h.logger.WarnContext(ctx, "readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}

	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func toAPITrip(trip domain.Trip) api.Trip {
	resp := api.Trip{
		Id:       trip.ID,
		UserId:   trip.UserID,
		DriverId: trip.DriverID,
		StartPoint: api.Coordinates{
			Latitude:  trip.StartPoint.Latitude,
			Longitude: trip.StartPoint.Longitude,
		},
		EndPoint: api.Coordinates{
			Latitude:  trip.EndPoint.Latitude,
			Longitude: trip.EndPoint.Longitude,
		},
		Price:          trip.Price,
		Status:         api.TripStatus(trip.Status),
		StartedAt:      trip.StartedAt.UTC(),
		FinishedAt:     nullable.NewNullNullable[time.Time](),
		LastPositionAt: nullable.NewNullNullable[time.Time](),
	}

	if trip.FinishedAt != nil {
		resp.FinishedAt = nullable.NewNullableWithValue(trip.FinishedAt.UTC())
	}

	return resp
}

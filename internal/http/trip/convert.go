package trip

import (
	"time"

	"github.com/oapi-codegen/nullable"

	domain "github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
)

func Response(trip domain.Trip) api.Trip {
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

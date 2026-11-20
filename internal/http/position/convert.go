package position

import (
	domain "github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
)

func Response(p domain.Position) api.TripPosition {
	return api.TripPosition{
		Id:     p.ID,
		TripId: p.TripID,
		Coordinates: api.Coordinates{
			Latitude:  p.Coordinates.Latitude,
			Longitude: p.Coordinates.Longitude,
		},
		RecordedAt: p.RecordedAt.UTC(),
	}
}

func ListResponse(items []domain.Position) api.TripPositionsResponse {
	positions := make([]api.TripPosition, 0, len(items))
	for _, item := range items {
		positions = append(positions, Response(item))
	}
	return api.TripPositionsResponse{Positions: positions}
}

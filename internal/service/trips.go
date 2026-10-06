package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain"
	"github.com/Chillobass/fork_template/internal/repository"
)

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type TripService struct {
	txManager TxManager
	trips     *repository.TripRepository
	history   *repository.StatusHistoryRepository
}

func NewTripService(
	txManager TxManager,
	trips *repository.TripRepository,
	history *repository.StatusHistoryRepository,
) *TripService {
	return &TripService{txManager: txManager, trips: trips, history: history}
}

type CreateTripInput struct {
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint domain.Point
	EndPoint   domain.Point
	Price      int64
}

func (s *TripService) Create(ctx context.Context, in CreateTripInput) (domain.Trip, error) {
	now := currentTime()

	trip := domain.Trip{
		ID:         uuid.New(),
		UserID:     in.UserID,
		DriverID:   in.DriverID,
		StartPoint: in.StartPoint,
		EndPoint:   in.EndPoint,
		Price:      in.Price,
		Status:     domain.TripStatusActive,
		StartedAt:  now,
	}

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		if err := s.trips.Create(ctx, trip); err != nil {
			return err
		}

		return s.history.Add(ctx, domain.StatusChange{
			TripID:    trip.ID,
			From:      nil,
			To:        domain.TripStatusActive,
			Reason:    "trip created",
			ChangedAt: now,
		})
	})
	if err != nil {
		return domain.Trip{}, fmt.Errorf("create trip: %w", err)
	}

	return trip, nil
}

func (s *TripService) Get(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	trip, err := s.trips.GetByID(ctx, id)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("get trip: %w", err)
	}

	return trip, nil
}

func (s *TripService) Finish(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	var trip domain.Trip

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		var err error

		trip, err = s.trips.Finish(ctx, id, currentTime())
		if err != nil {
			return err
		}

		from := domain.TripStatusActive

		return s.history.Add(ctx, domain.StatusChange{
			TripID:    trip.ID,
			From:      &from,
			To:        domain.TripStatusCompleted,
			Reason:    "trip finished",
			ChangedAt: *trip.FinishedAt,
		})
	})
	if err != nil {
		return domain.Trip{}, fmt.Errorf("finish trip: %w", err)
	}

	return trip, nil
}

// PostgreSQL хранит время с точностью до микросекунд: без Truncate ответ на создание разошёлся бы с GET
func currentTime() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

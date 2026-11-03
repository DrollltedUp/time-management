package shifts

import (
	"context"
	"time"
)

type Service struct {
	store Store
}

func NewService(s Store) *Service {
	return &Service{
		store: s,
	}
}

func (s *Service) Create(ctx context.Context, userID, storeID int64, startsAt, endsAt time.Time) (Shift, error) {
	if userID <= 0 || storeID <= 0 || !startsAt.Before(endsAt) {
		return Shift{}, ErrInvalidInput
	}

	if endsAt.Sub(startsAt) > 24*time.Hour {
		return Shift{}, ErrInvalidInput
	}

	assigned, err := s.store.IsUserAssigned(ctx, userID, storeID)
	if err != nil {
		return Shift{}, err
	}
	if !assigned {
		return Shift{}, ErrNotAssigned
	}

	return s.store.Create(ctx, Shift{
		UserID:   userID,
		StoreID:  storeID,
		StartsAt: startsAt,
		EndsAt:   endsAt,
	})
}

func (s *Service) Get(ctx context.Context, id int64) (Shift, error) {
	if id <= 0 {
		return Shift{}, ErrInvalidInput
	}
	return s.store.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, userID, storeID int64, startsAt, endsAt time.Time) (Shift, error) {
	if id <= 0 || userID <= 0 || storeID <= 0 || !startsAt.Before(startsAt) {
		return Shift{}, ErrInvalidInput
	}

	if endsAt.Sub(startsAt) > time.Hour*24 {
		return Shift{}, ErrInvalidInput
	}

	assigned, err := s.store.IsUserAssigned(ctx, userID, storeID)
	if err != nil {
		return Shift{}, err
	}

	if !assigned {
		return Shift{}, err
	}

	return s.store.Update(ctx, Shift{
		ID:       id,
		UserID:   userID,
		StoreID:  storeID,
		StartsAt: startsAt,
		EndsAt:   endsAt,
	})
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}
	return s.store.Delete(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userId int64, from, to time.Time) ([]WithStore, error) {
	if userId <= 0 || !from.Before(to) {
		return nil, ErrInvalidInput
	}
	return s.store.ListByUser(ctx, userId, from, to)
}

func (s *Service) ListByStore(ctx context.Context, storeID int64, from, to time.Time) ([]WithUser, error) {
	if storeID <= 0 || !from.Before(to) {
		return nil, ErrInvalidInput
	}

	return s.store.ListByStore(ctx, storeID, from, to)
}

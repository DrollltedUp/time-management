package store

import "context"

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Create(ctx context.Context, store Stores) (Stores, error) {
	return Stores{}, nil
}

package postgres

import (
	"context"

	"github.com/drollltedUp/time-managment/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoresStore struct {
	pool *pgxpool.Pool
}

func NewStoresStore(pool *pgxpool.Pool) *StoresStore {
	return &StoresStore{
		pool: pool,
	}
}

func (s *StoresStore) Create(ctx context.Context, st store.Stores) (store.Stores, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = ` INSERT INTO stores (number, sort_number, address) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := s.pool.QueryRow(ctx, q, st.Number, st.Sort_number, st.Address).Scan(&st.ID, &st.Created_at)
	if err != nil {
		return store.Stores{}, err
	}

	return st, nil
}

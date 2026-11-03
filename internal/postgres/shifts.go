package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/drollltedUp/time-managment/internal/shifts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	codeForeighKeyViolation = "23503"
	codeExclusionViolation  = "23P01"
)

type ShiftStore struct {
	pool *pgxpool.Pool
}

func NewShiftStore(pool *pgxpool.Pool) *ShiftStore {
	return &ShiftStore{
		pool: pool,
	}
}

func (s *ShiftStore) Create(ctx context.Context, sh shifts.Shift) (shifts.Shift, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `INSERT INTO shifts (user_id, store_id, starts_at, ends_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := s.pool.QueryRow(ctx, q, sh.UserID, sh.StoreID, sh.StartsAt, sh.EndsAt).Scan(&sh.ID, &sh.CreatedAt)
	if err != nil {
		return shifts.Shift{}, err
	}

	return sh, nil

}

func (s *ShiftStore) GetByID(ctx context.Context, id int64) (shifts.Shift, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `SELECT id, user_id, store_id, starts_at, ends_at, created_at FROM shifts WHERE id = $1`

	var sh shifts.Shift

	err := s.pool.QueryRow(ctx, q, id).Scan(&sh.ID, &sh.UserID, &sh.StoreID, &sh.StartsAt, &sh.EndsAt, &sh.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return shifts.Shift{}, shifts.ErrNotFound
	}
	if err != nil {
		return shifts.Shift{}, err
	}
	return sh, nil
}

func (s *ShiftStore) Update(ctx context.Context, sh shifts.Shift) (shifts.Shift, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `UPDATE shifts SET user_id = $1, store_id = $2, starts_at = $3, ends_at = $4 WHERE id = $5 RETURNING id, user_id, store_id, starts_at, ends_at, created_at`
	err := s.pool.QueryRow(ctx, q, sh.UserID, sh.StoreID, sh.StartsAt, sh.EndsAt).Scan(&sh.ID, &sh.UserID, &sh.StoreID, &sh.StartsAt, &sh.EndsAt, &sh.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return shifts.Shift{}, shifts.ErrNotFound
	}

	if err != nil {
		return shifts.Shift{}, err
	}

	return sh, nil
}

func (s *ShiftStore) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var shift shifts.Shift

	const q = `DELETE FROM shifts WHERE id = $1`
	tag, err := s.pool.Exec(ctx, q, shift.ID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return err
	}

	return nil
}

func (s *ShiftStore) ListByUser(ctx context.Context, userID int64, from, to time.Time) ([]shifts.WithStore, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
        SELECT sh.id, sh.user_id, sh.store_id, sh.starts_at, sh.ends_at, sh.created_at,
               st.number, st.address
        FROM shifts sh
        JOIN stores st ON st.id = sh.store_id
        WHERE sh.user_id = $1
          AND sh.starts_at < $3
          AND sh.ends_at   > $2
        ORDER BY sh.starts_at
    `, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []shifts.WithStore
	for rows.Next() {
		var w shifts.WithStore
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.StoreID, &w.StartsAt, &w.EndsAt, &w.CreatedAt,
			&w.StoreNumber, &w.StoreAddress,
		); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()

}

func (s *ShiftStore) ListByStore(ctx context.Context, storeID int64, from, to time.Time) ([]shifts.WithUser, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
        SELECT sh.id, sh.user_id, sh.store_id, sh.starts_at, sh.ends_at, sh.created_at,
               u.full_name
        FROM shifts sh
        JOIN users u ON u.id = sh.user_id
        WHERE sh.store_id = $1
          AND sh.starts_at < $3
          AND sh.ends_at   > $2
        ORDER BY sh.starts_at
    `, storeID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []shifts.WithUser
	for rows.Next() {
		var w shifts.WithUser
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.StoreID, &w.StartsAt, &w.EndsAt, &w.CreatedAt,
			&w.Fullname,
		); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *ShiftStore) IsUserAssigned(ctx context.Context, userID, storeID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_stores WHERE user_id = $1 AND store_id = $2)`, userID, storeID).Scan(&exists)
	return exists, err
}

func mapShiftError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case codeExclusionViolation:
			return shifts.ErrOverlap
		case codeForeighKeyViolation:
			return shifts.ErrUserOrStoreNotFount
		}
	}
	return err
}

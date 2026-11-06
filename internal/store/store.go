package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("stores no found")
	ErrInvalidInput = errors.New("invalid shift input")
)

type Stores struct {
	ID          int64
	Number      string
	Sort_number string
	Address     string
	Created_at  time.Time
}

type Store interface {
	Create(ctx context.Context, store Stores) (Stores, error)
	Get(ctx context.Context, id int64) (Stores, error)
	Update(ctx context.Context, store Stores) (Stores, error)
	Delete(ctx context.Context, id int64) error
}

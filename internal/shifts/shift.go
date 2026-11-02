package shifts

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound            = errors.New("shift not found")
	ErrInvalidInput        = errors.New("invalid shift input")
	ErrOverlap             = errors.New("shift overlaps existing")
	ErrUserOrStoreNotFount = errors.New("user or shore not found")
	ErrNotAssigned         = errors.New("user is not assigned to store")
)

type DayNight string

const (
	Day   DayNight = "day"
	Night DayNight = "night"
	Other DayNight = "other"
)

type Shift struct {
	ID        int64
	UserID    int64
	StoreID   int64
	StartsAt  time.Time
	EndsAt    time.Time
	CreatedAt time.Time
}

func (s Shift) Type() DayNight {
	switch s.StartsAt.UTC().Hour() {
	case 8:
		return Day
	case 20:
		return Night
	}
	return Other
}

type WithUser struct {
	Shift
	Fullname string
}

type WithStore struct {
	Shift
	StoreNumber  string
	StoreAddress string
}

type Store interface {
	Create(ctx context.Context, s Shift) (Shift, error)
	GetByID(ctx context.Context, id int64) (Shift, error)
	Update(ctx context.Context, s Shift) (Shift, error)
	Delete(ctx context.Context, id int64) error

	ListByUser(ctx context.Context, userID int64, from, to time.Time) ([]WithStore, error)
	ListByStore(ctx context.Context, storeID int64, from, to time.Time) ([]WithUser, error)
	IsUserAssigned(ctx context.Context, userID, storeID int64) (bool, error)
}

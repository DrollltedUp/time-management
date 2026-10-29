package user

import (
	"context"
	"errors"
	"time"
)

type User struct {
	ID            int64
	Email         string
	Password_hash string
	Fullname      string
	role          string
	Active        bool
	Created_at    time.Time
}

var (
	ErrEmailTaken         = errors.New("error: email is taken")
	ErrNotFound           = errors.New("error: user not found")
	ErrInvalidInput       = errors.New("error: email or password is invalid format")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Store interface {
	Create(ctx context.Context, u User) (User, error)
	GetByMail(ctx context.Context, email string) (User, error)
}

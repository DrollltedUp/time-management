package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/drollltedUp/time-managment/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ user.Store = (*UserStore)(nil)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{
		pool: pool,
	}
}

func (u *UserStore) Create(ctx context.Context, usr user.User) (user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	err := u.pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, full_name, role, active, created_at)
	VALUES ($1, $2, $3, $4, $5, now()) RETURNING id, created_at
	`, usr.Email, usr.Password_hash, usr.Fullname, usr.Role, usr.Active).Scan(&usr.ID, &usr.Created_at)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return user.User{}, user.ErrEmailTaken
		}
		return user.User{}, err
	}
	return usr, err

}

func (u *UserStore) GetByMail(ctx context.Context, email string) (user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	query := `SELECT id, email, password_hash, full_name, role, active, created_at FROM users WHERE email = $1`
	var usr user.User

	err := u.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(&usr.ID, &usr.Email, &usr.Password_hash, &usr.Fullname, &usr.Role, &usr.Active, &usr.Created_at)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}

	return usr, err
}

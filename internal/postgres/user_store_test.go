package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/drollltedUp/time-managment/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return pool
}

// truncateUsers очищает таблицу между тестами.
func truncateUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `TRUNCATE users RESTART IDENTITY`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func sampleUser() user.User {
	return user.User{
		Email:         "alice@example.com",
		Password_hash: "$2a$10$abcdefghijklmnopqrstuv", // не важно для store
		Fullname:      "Alice Smith",
		Role:          user.RoleEmployee,
		Active:        true,
	}
}

func TestUserStore_CreateAndGet(t *testing.T) {
	pool := setupPool(t)
	truncateUsers(t, pool)

	store := NewUserStore(pool)
	ctx := context.Background()

	created, err := store.Create(ctx, sampleUser())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("id not assigned")
	}
	if created.Created_at.IsZero() {
		t.Error("created_at not set")
	}

	got, err := store.GetByMail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("id mismatch: got %d want %d", got.ID, created.ID)
	}
	if got.Email != created.Email {
		t.Errorf("email mismatch: got %q", got.Email)
	}
	if got.Fullname != created.Fullname {
		t.Errorf("fullname mismatch: got %q", got.Fullname)
	}
	if got.Role != created.Role {
		t.Errorf("role mismatch: got %q", got.Role)
	}
	if got.Active != created.Active {
		t.Errorf("active mismatch: got %v", got.Active)
	}
}

func TestUserStore_CreateDuplicate(t *testing.T) {
	pool := setupPool(t)
	truncateUsers(t, pool)

	store := NewUserStore(pool)
	ctx := context.Background()

	if _, err := store.Create(ctx, sampleUser()); err != nil {
		t.Fatal(err)
	}
	_, err := store.Create(ctx, sampleUser())
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("want ErrEmailTaken, got %v", err)
	}
}

func TestUserStore_GetByMail_NotFound(t *testing.T) {
	pool := setupPool(t)
	truncateUsers(t, pool)

	store := NewUserStore(pool)
	_, err := store.GetByMail(context.Background(), "nobody@example.com")
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUserStore_GetByMail_CaseInsensitive(t *testing.T) {
	pool := setupPool(t)
	truncateUsers(t, pool)

	store := NewUserStore(pool)
	ctx := context.Background()

	if _, err := store.Create(ctx, sampleUser()); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetByMail(ctx, "ALICE@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Email != "alice@example.com" {
		t.Errorf("got email %q", got.Email)
	}
}

func TestUserStore_Create_TimeoutContext(t *testing.T) {
	pool := setupPool(t)
	truncateUsers(t, pool)

	store := NewUserStore(pool)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // сразу отменён

	_, err := store.Create(ctx, sampleUser())
	if err == nil {
		t.Fatal("expected error on canceled context")
	}
}

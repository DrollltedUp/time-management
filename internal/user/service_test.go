package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func newTestService() *Service {
	return NewService(NewMemoryStore())
}

func TestRegister_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	u, err := svc.Register(ctx, "  Alice@Example.COM ", "password123", "  Alice   Smith ")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.ID == 0 {
		t.Error("id not assigned")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("email not normalized: %q", u.Email)
	}
	if u.Fullname != "Alice Smith" {
		t.Errorf("fullname not normalized: %q", u.Fullname)
	}
	if u.Role != RoleEmployee {
		t.Errorf("role: got %q", u.Role)
	}
	if !u.Active {
		t.Error("expected active")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password_hash), []byte("password123")); err != nil {
		t.Errorf("password not hashed correctly: %v", err)
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	cases := []struct {
		name     string
		email    string
		password string
		fullname string
	}{
		{"bad email", "not-an-email", "password123", "Alice"},
		{"short password", "a@b.com", "short", "Alice"},
		{"long password", "a@b.com", strings.Repeat("x", 73), "Alice"},
		{"empty name", "a@b.com", "password123", "   "},
		{"long name", "a@b.com", "password123", strings.Repeat("а", 101)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(ctx, tc.email, tc.password, tc.fullname)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("want ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "a@b.com", "password123", "Alice"); err != nil {
		t.Fatal(err)
	}
	// разный регистр — всё равно дубликат
	_, err := svc.Register(ctx, "A@B.COM", "password123", "Bob")
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("want ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "a@b.com", "password123", "Alice"); err != nil {
		t.Fatal(err)
	}
	u, err := svc.Login(ctx, "A@B.com", "password123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.Email != "a@b.com" {
		t.Errorf("got email %q", u.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "a@b.com", "password123", "Alice"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Login(ctx, "a@b.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	_, err := svc.Login(ctx, "nobody@b.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/drollltedUp/time-managment/internal/auth"
	"github.com/drollltedUp/time-managment/internal/postgres"
	"github.com/drollltedUp/time-managment/internal/shifts"
	"github.com/drollltedUp/time-managment/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	secret := os.Getenv("JWT_TOKEN")
	if secret == "" {
		log.Error("JWT_TOKEN is not set")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Error("Cannot create pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	usersrv := user.NewService(postgres.NewUserStore(pool))
	shiftssrv := shifts.NewService(postgres.NewShiftStore(pool))
	tokens := auth.NewTokenManager([]byte(secret), time.Hour)
	handler := NewServerHTTP(usersrv, shiftssrv, tokens, log).router()

	server := &http.Server{
		Addr:              ":9099",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Info("server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server failder", "err", err)
		os.Exit(1)
	}

}

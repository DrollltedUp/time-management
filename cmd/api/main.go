package main

import (
	"context"
	"log/slog"
	"os"

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

	// srv := &http.Server{
	// 	Addr:              ":9099",
	// 	Handler:           NewServerHTTP().router(),
	// 	ReadHeaderTimeout: 5 * time.Second,
	// 	ReadTimeout:       10 * time.Second,
	// 	WriteTimeout:      10 * time.Second,
	// }

	// log.Info("server started", "addr", srv.Addr)
	// if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	// 	log.Error("server failder", "err", err)
	// 	os.Exit(1)
	// }

}

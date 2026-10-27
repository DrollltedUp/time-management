package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerHTTP struct {
}

func NewServerHTTP() *ServerHTTP {
	return &ServerHTTP{}
}

func (s *ServerHTTP) router(pool *pgxpool.Pool, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(pool, log))

	return mux
}

func healthHandler(pool *pgxpool.Pool, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(ctx); err != nil {
			log.Warn("healthcheck: database unavailable", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/drollltedUp/time-managment/internal/auth"
	"github.com/drollltedUp/time-managment/internal/user"
	"github.com/gorilla/mux"
)

type ServerHTTP struct {
	authSVC *user.Service
	tokens  *auth.TokenManager
	slog    *slog.Logger
}

func NewServerHTTP(auth *user.Service, tokens *auth.TokenManager, slog *slog.Logger) *ServerHTTP {
	return &ServerHTTP{
		authSVC: auth,
		tokens:  tokens,
		slog:    slog,
	}
}

func (s *ServerHTTP) router() http.Handler {
	router := mux.NewRouter()
	router.Path("/register").Methods("POST").HandlerFunc(s.Register)
	router.Path("/login").Methods("POST").HandlerFunc(s.Login)

	protected := router.PathPrefix("/").Subrouter()
	protected.Use(s.authMiddleware)
	protected.Path("/me").Methods("GET").HandlerFunc(s.Me)

	return router
}

func (s *ServerHTTP) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	u, err := s.authSVC.Register(r.Context(), request.Email, request.Password, request.FullName)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidInput):
			errorJSON(w, http.StatusBadRequest, "bad request")
		case errors.Is(err, user.ErrEmailTaken):
			errorJSON(w, http.StatusConflict, "email already taken")
		default:
			errorJSON(w, http.StatusInternalServerError, "internal server")
		}
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{
		ID:       u.ID,
		Email:    u.Email,
		FullName: u.Fullname,
	})
}

func (s *ServerHTTP) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var login loginRequest
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	u, err := s.authSVC.Login(r.Context(), login.Email, login.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			errorJSON(w, http.StatusUnauthorized, "invalid email or password")
		default:
			errorJSON(w, http.StatusInternalServerError, "invalid server")
		}
		return
	}

	token, err := s.tokens.Issue(int(u.ID), auth.Role(u.Role))
	if err != nil {
		s.slog.Error("issue token failed ", "err", err)
		errorJSON(w, http.StatusInternalServerError, "internal server")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
	})
}

func (s *ServerHTTP) Me(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(ctxKeyUserID{}).(int)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}

	role, _ := r.Context().Value(ctxKeyRole{}).(string)
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": userId,
		"role":    role,
	})
}

func (s *ServerHTTP) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		token, ok := strings.CutPrefix(header, prefix)
		if !ok || token == "" {
			errorJSON(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, err := s.tokens.Parse(token)
		if errors.Is(err, auth.ErrExpiredToken) {
			errorJSON(w, http.StatusUnauthorized, "token expired")
			return
		}
		if err != nil {
			errorJSON(w, http.StatusUnauthorized, "invalid token")
			return
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			errorJSON(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID{}, userID)
		ctx = context.WithValue(ctx, ctxKeyRole{}, string(claims.Role))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

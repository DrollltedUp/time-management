package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type ctxKeyUserID struct{}
type ctxKeyRole struct{}

type registerResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type errorDTO struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// Struct For Shifts

type createShiftRequest struct {
	UserID   int64     `json:"user_id"`
	StoreID  int64     `json:"store_id"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

// Functions

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func errorJSON(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorDTO{Message: msg, Time: time.Now()})
}

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/drollltedUp/time-managment/internal/shifts"
	"github.com/gorilla/mux"
)

func (s *ServerHTTP) HandlerShiftCreate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	sh, err := s.shiftsSVC.Create(r.Context(), req.UserID, req.StoreID, req.StartsAt, req.EndsAt)
	if err != nil {
		switch {
		case errors.Is(err, shifts.ErrOverlap):
			errorJSON(w, http.StatusConflict, "shift overlaps existings")
		case errors.Is(err, shifts.ErrUserOrStoreNotFount):
			errorJSON(w, http.StatusNotFound, "user or store not found")
		case errors.Is(err, shifts.ErrNotAssigned):
			errorJSON(w, http.StatusBadRequest, "user is not assigned to this store")
		case errors.Is(err, shifts.ErrInvalidInput):
			errorJSON(w, http.StatusBadRequest, "bad request")
		default:
			s.slog.Error("create shift", "err", err)
			errorJSON(w, http.StatusInternalServerError, "internal server")
		}
		return
	}
	writeJSON(w, http.StatusCreated, shifts.ToDTO(sh))

}

func (s *ServerHTTP) HandlerShiftGetID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}
	sh, err := s.shiftsSVC.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, shifts.ErrNotFound) {
			errorJSON(w, http.StatusNotFound, "shift not found")
			return
		}
		s.slog.Error("get shift", "err", err)
		errorJSON(w, http.StatusInternalServerError, "internal server")
		return
	}
	writeJSON(w, http.StatusOK, shifts.ToDTO(sh))

}

func (s *ServerHTTP) HandlerShiftUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	var req createShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	sh, err := s.shiftsSVC.Update(r.Context(), id, req.UserID, req.StoreID, req.StartsAt, req.EndsAt)
	if err != nil {
		switch {
		case errors.Is(err, shifts.ErrOverlap):
			errorJSON(w, http.StatusConflict, "shift overlaps existings")
		case errors.Is(err, shifts.ErrUserOrStoreNotFount):
			errorJSON(w, http.StatusNotFound, "user or store not found")
		case errors.Is(err, shifts.ErrNotAssigned):
			errorJSON(w, http.StatusBadRequest, "user is not assigned to this store")
		case errors.Is(err, shifts.ErrInvalidInput):
			errorJSON(w, http.StatusBadRequest, "bad request")
		default:
			s.slog.Error("create shift", "err", err)
			errorJSON(w, http.StatusInternalServerError, "internal server")
		}
		return
	}

	writeJSON(w, http.StatusOK, shifts.ToDTO(sh))
}

func (s *ServerHTTP) HandlerShiftDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	err = s.shiftsSVC.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, shifts.ErrNotFound) {
			errorJSON(w, http.StatusNotFound, "not found for delete")
			return
		}
		s.slog.Error("delete shift", "err", err)
		errorJSON(w, http.StatusInternalServerError, "internal server")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *ServerHTTP) HandlerMyShifts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxKeyUserID{}).(int)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	from, to, err := parseWindow(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid from/to")
		return
	}

	list, err := s.shiftsSVC.ListByUser(r.Context(), int64(userID), from, to)
	if err != nil {
		s.slog.Error("list my shifts", "err", err)
		errorJSON(w, http.StatusInternalServerError, "internal server")
		return
	}

	out := make([]shifts.ShiftWithStoreDTO, 0, len(list))
	for _, w := range list {
		out = append(out, shifts.ShiftWithStoreDTO{
			ShiftDTO:     shifts.ToDTO(w.Shift),
			StoreNumber:  w.StoreNumber,
			StoreAddress: w.StoreAddress,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"shifts": out})
}

func (s *ServerHTTP) HandlerShiftsStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || storeID <= 0 {
		errorJSON(w, http.StatusBadRequest, "bad request")
		return
	}

	from, to, err := parseWindow(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid from/to")
		return
	}
	list, err := s.shiftsSVC.ListByStore(r.Context(), storeID, from, to)
	if err != nil {
		s.slog.Error("list store shifts", "err", err)
		errorJSON(w, http.StatusInternalServerError, "internal server")
		return
	}

	out := make([]shifts.ShiftWithUSerDTO, 0, len(list))
	for _, w := range list {
		out = append(out, shifts.ShiftWithUSerDTO{
			ShiftDTO: shifts.ToDTO(w.Shift),
			FullName: w.Fullname,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"shifts": out})
}

func parseWindow(r *http.Request) (time.Time, time.Time, error) {
	q := r.URL.Query()
	fromStr, toStr := q.Get("from"), q.Get("to")

	var from, to time.Time
	var err error

	if fromStr == "" {
		from = time.Now().UTC().Truncate(24 * time.Hour)
	} else if from, err = time.Parse(time.RFC3339, fromStr); err != nil {
		return time.Time{}, time.Time{}, err
	}

	if toStr == "" {
		to = from.AddDate(0, 0, 7)
	} else if to, err = time.Parse(time.RFC3339, toStr); err != nil {
		return time.Time{}, time.Time{}, err
	}

	if !from.Before(to) {
		return time.Time{}, time.Time{}, errors.New("from >= to")
	}
	if to.Sub(from) > 90*24*time.Hour {
		return time.Time{}, time.Time{}, errors.New("window too large")
	}
	return from, to, nil
}

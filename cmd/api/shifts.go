package main

import (
	"net/http"
	"time"
)

func (s *ServerHTTP) HandlerShiftCreate(w http.ResponseWriter, r *http.Request) {

}

func (s *ServerHTTP) HandlerShiftGetID(w http.ResponseWriter, r *http.Request) {

}

func (s *ServerHTTP) HandlerShiftUpdate(w http.ResponseWriter, r *http.Request) {

}

func (s *ServerHTTP) HandlerShiftDelete(w http.ResponseWriter, r *http.Request) {

}

func (s *ServerHTTP) HandlerMyShifts(w http.ResponseWriter, r *http.Request) {

}

func (s *ServerHTTP) HandlerShiftsStore(w http.ResponseWriter, r *http.Request) {

}

func parseWindow(r *http.Request) (time.Time, time.Time, error) {
	return time.Now(), time.Now(), nil
}

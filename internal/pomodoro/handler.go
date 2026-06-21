package pomodoro

import (
	"encoding/json"
	"net/http"

	"simple-daily-termux/internal/httputil"
)

type Handler struct {
	svc      *Service
	timezone string
}

func RegisterHandler(mux *http.ServeMux, svc *Service, timezone string) {
	h := &Handler{svc: svc, timezone: timezone}
	mux.HandleFunc("POST /api/pomodoro/start", h.Start)
	mux.HandleFunc("POST /api/pomodoro/{id}/finish", h.Finish)
	mux.HandleFunc("GET /api/pomodoro/today", h.GetToday)
	mux.HandleFunc("GET /api/pomodoro", h.ListRange)
}

type startReq struct {
	PlannedMinutes int    `json:"planned_minutes"`
	LinkedTodoID   string `json:"linked_todo_id"`
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	var req startReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.PlannedMinutes <= 0 {
		req.PlannedMinutes = 25
	}

	session, err := h.svc.Start(r.Context(), req.PlannedMinutes, req.LinkedTodoID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, session)
}

type finishReq struct {
	Status string `json:"status"`
}

func (h *Handler) Finish(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req finishReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Status != StatusCompleted && req.Status != StatusAborted {
		req.Status = StatusCompleted
	}

	session, err := h.svc.Finish(r.Context(), id, req.Status)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, session)
}

func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	minutes, err := h.svc.GetTodayMinutes(r.Context(), h.timezone)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]int{"total_minutes": minutes})
}

func (h *Handler) ListRange(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	sessions, err := h.svc.ListRange(r.Context(), from, to)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sessions == nil {
		sessions = []*Session{}
	}
	httputil.JSON(w, http.StatusOK, sessions)
}

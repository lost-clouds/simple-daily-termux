package calendar

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simple-daily-termux/internal/httputil"
)

type Handler struct {
	svc *Service
}

func RegisterHandler(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("GET /api/calendar", h.GetMonthView)
	mux.HandleFunc("POST /api/calendar/events", h.CreateEvent)
	mux.HandleFunc("PUT /api/calendar/events/{id}", h.UpdateEvent)
	mux.HandleFunc("DELETE /api/calendar/events/{id}", h.DeleteEvent)
}

func (h *Handler) GetMonthView(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	year, mon := parseYM(month)
	if year == 0 {
		now := time.Now()
		year, mon = now.Year(), int(now.Month())
	}

	view, err := h.svc.GetMonthView(r.Context(), year, mon)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, view)
}

type createEventReq struct {
	Title   string  `json:"title"`
	StartAt string  `json:"start_at"`
	EndAt   *string `json:"end_at"`
	AllDay  bool    `json:"all_day"`
	Note    string  `json:"note"`
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req createEventReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Title == "" || req.StartAt == "" {
		httputil.Error(w, http.StatusBadRequest, "title and start_at are required")
		return
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid start_at format")
		return
	}

	var endAt *time.Time
	if req.EndAt != nil && *req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid end_at format")
			return
		}
		endAt = &t
	}

	e, err := h.svc.CreateEvent(r.Context(), req.Title, startAt, endAt, req.AllDay, req.Note)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, e)
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req createEventReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Title == "" || req.StartAt == "" {
		httputil.Error(w, http.StatusBadRequest, "title and start_at are required")
		return
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid start_at format")
		return
	}

	var endAt *time.Time
	if req.EndAt != nil && *req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid end_at format")
			return
		}
		endAt = &t
	}

	e, err := h.svc.UpdateEvent(r.Context(), id, req.Title, startAt, endAt, req.AllDay, req.Note)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteEvent(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, nil)
}

func parseYM(val string) (int, int) {
	if val == "" {
		return 0, 0
	}
	parts := strings.SplitN(val, "-", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	y, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return y, m
}

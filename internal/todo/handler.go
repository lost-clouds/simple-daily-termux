package todo

import (
	"encoding/json"
	"net/http"
	"time"

	"simple-daily-termux/internal/httputil"
)

type Handler struct {
	svc *Service
}

func RegisterHandler(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("GET /api/todos", h.List)
	mux.HandleFunc("POST /api/todos", h.Create)
	mux.HandleFunc("GET /api/todos/{id}", h.Get)
	mux.HandleFunc("PUT /api/todos/{id}", h.Update)
	mux.HandleFunc("DELETE /api/todos/{id}", h.Delete)
}

type createTodoReq struct {
	Title      string  `json:"title"`
	Notes      string  `json:"notes"`
	Priority   int     `json:"priority"`
	DeadlineAt *string `json:"deadline_at"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f := ListFilter{
		Status:      r.URL.Query().Get("status"),
		HasDeadline: r.URL.Query().Get("has_deadline"),
	}
	todos, err := h.svc.List(r.Context(), f)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if todos == nil {
		todos = []*Todo{}
	}
	httputil.JSON(w, http.StatusOK, todos)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTodoReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	var deadline *time.Time
	if req.DeadlineAt != nil && *req.DeadlineAt != "" {
		t, err := time.Parse(time.RFC3339, *req.DeadlineAt)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid deadline_at format, use RFC3339")
			return
		}
		deadline = &t
	}

	t, err := h.svc.Create(r.Context(), req.Title, req.Notes, req.Priority, deadline)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, t)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "todo not found")
		return
	}
	httputil.JSON(w, http.StatusOK, t)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "todo not found")
		return
	}

	var req struct {
		Title      string  `json:"title"`
		Notes      string  `json:"notes"`
		Status     string  `json:"status"`
		Priority   int     `json:"priority"`
		DeadlineAt *string `json:"deadline_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title != "" {
		t.Title = req.Title
	}
	t.Notes = req.Notes
	if req.Status != "" {
		t.Status = req.Status
	}
	t.Priority = req.Priority
	if req.DeadlineAt != nil {
		if *req.DeadlineAt == "" || *req.DeadlineAt == "null" {
			t.DeadlineAt = nil
		} else {
			parsed, err := time.Parse(time.RFC3339, *req.DeadlineAt)
			if err != nil {
				httputil.Error(w, http.StatusBadRequest, "invalid deadline_at format, use RFC3339")
				return
			}
			t.DeadlineAt = &parsed
		}
	}

	if err := h.svc.Update(r.Context(), t); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, t)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, nil)
}

package todo

import (
	"encoding/json"
	"net/http"
	"time"

	"simple-daily-termux/internal/httputil"
)

type Handler struct{ svc *Service }

func RegisterHandler(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("GET /api/todos", h.List)
	mux.HandleFunc("POST /api/todos", h.Create)
	mux.HandleFunc("GET /api/todos/{id}", h.Get)
	mux.HandleFunc("PUT /api/todos/{id}", h.Update)
	mux.HandleFunc("DELETE /api/todos/{id}", h.Delete)
	mux.HandleFunc("POST /api/todos/ensure-daily", h.EnsureDaily)
}

type createReq struct {
	Title      string  `json:"title"`
	Notes      string  `json:"notes"`
	TaskType   string  `json:"task_type"`
	DeadlineAt *string `json:"deadline_at"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f := ListFilter{
		Status:      r.URL.Query().Get("status"),
		TaskType:    r.URL.Query().Get("task_type"),
		HasDeadline: r.URL.Query().Get("has_deadline"),
		EntryDate:   r.URL.Query().Get("entry_date"),
	}
	todos, err := h.svc.List(r.Context(), f)
	if err != nil { httputil.Error(w, 500, err.Error()); return }
	if todos == nil { todos = []*Todo{} }
	httputil.JSON(w, 200, todos)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { httputil.Error(w, 400, "invalid JSON"); return }
	if req.Title == "" { httputil.Error(w, 400, "title required"); return }
	if req.TaskType == "" { req.TaskType = TaskOneTime }

	var deadline *time.Time
	if req.DeadlineAt != nil && *req.DeadlineAt != "" {
		t, err := time.Parse(time.RFC3339, *req.DeadlineAt)
		if err != nil { httputil.Error(w, 400, "invalid deadline_at"); return }
		deadline = &t
	}
	td, err := h.svc.Create(r.Context(), req.Title, req.Notes, req.TaskType, deadline)
	if err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 201, td)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil { httputil.Error(w, 404, "not found"); return }
	httputil.JSON(w, 200, t)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.svc.Get(r.Context(), id)
	if err != nil { httputil.Error(w, 404, "not found"); return }

	var req struct {
		Title      string  `json:"title"`
		Notes      string  `json:"notes"`
		Status     string  `json:"status"`
		TaskType   string  `json:"task_type"`
		DeadlineAt *string `json:"deadline_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { httputil.Error(w, 400, "invalid JSON"); return }
	if req.Title != "" { t.Title = req.Title }
	t.Notes = req.Notes
	if req.Status != "" { t.Status = req.Status }
	if req.TaskType != "" { t.TaskType = req.TaskType }
	if req.DeadlineAt != nil {
		if *req.DeadlineAt == "" || *req.DeadlineAt == "null" {
			t.DeadlineAt = nil
		} else {
			parsed, err := time.Parse(time.RFC3339, *req.DeadlineAt)
			if err != nil { httputil.Error(w, 400, "invalid deadline_at"); return }
			t.DeadlineAt = &parsed
		}
	}
	if err := h.svc.Update(r.Context(), t); err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 200, t)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 200, nil)
}

func (h *Handler) EnsureDaily(w http.ResponseWriter, r *http.Request) {
	today := r.URL.Query().Get("date")
	if today == "" {
		var req struct{ Date string `json:"date"` }
		json.NewDecoder(r.Body).Decode(&req)
		today = req.Date
	}
	if today == "" { today = time.Now().Format("2006-01-02") }
	todos, err := h.svc.EnsureDailyTasks(r.Context(), today)
	if err != nil { httputil.Error(w, 500, err.Error()); return }
	if todos == nil { todos = []*Todo{} }
	httputil.JSON(w, 200, todos)
}

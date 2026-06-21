package countdown

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
	mux.HandleFunc("GET /api/countdown", h.List)
	mux.HandleFunc("POST /api/countdown", h.Create)
	mux.HandleFunc("DELETE /api/countdown/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if events == nil {
		events = []*Event{}
	}
	httputil.JSON(w, http.StatusOK, events)
}

type createCountdownReq struct {
	Title    string `json:"title"`
	TargetAt string `json:"target_at"`
	Note     string `json:"note"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCountdownReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Title == "" || req.TargetAt == "" {
		httputil.Error(w, http.StatusBadRequest, "title and target_at are required")
		return
	}

	targetAt, err := time.Parse(time.RFC3339, req.TargetAt)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid target_at format, use RFC3339")
		return
	}

	e, err := h.svc.Create(r.Context(), req.Title, targetAt, req.Note)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, e)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, nil)
}

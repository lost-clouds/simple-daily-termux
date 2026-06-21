package summary

import (
	"net/http"

	"simple-daily-termux/internal/httputil"
)

type Handler struct {
	svc *Service
}

func RegisterHandler(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("GET /api/summary", h.GetSummary)
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetSummary(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, data)
}

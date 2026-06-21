package diary

import (
	"encoding/json"
	"errors"
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
	mux.HandleFunc("GET /api/diary/{date}", h.Get)
	mux.HandleFunc("PUT /api/diary/{date}", h.Save)
	mux.HandleFunc("GET /api/diary", h.ListMonth)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	entry, err := h.svc.Get(r.Context(), date)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, http.StatusOK, nil)
		} else {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httputil.JSON(w, http.StatusOK, entry)
}

type saveDiaryReq struct {
	ContentMD string `json:"content_md"`
	Mood      string `json:"mood"`
}

func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	var req saveDiaryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	entry, err := h.svc.Save(r.Context(), date, req.ContentMD, req.Mood)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, entry)
}

func (h *Handler) ListMonth(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	year, mon := parseYearMonth(month)
	if year == 0 {
		now := time.Now()
		year, mon = now.Year(), int(now.Month())
	}

	entries, err := h.svc.ListMonth(r.Context(), year, mon)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entries == nil {
		entries = []*MonthEntry{}
	}
	httputil.JSON(w, http.StatusOK, entries)
}

func parseYearMonth(val string) (int, int) {
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

package diary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"simple-daily-termux/internal/dateutil"
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
	mux.HandleFunc("GET /api/diary/export", h.ExportMD)
	mux.HandleFunc("POST /api/diary/import", h.ImportMD)
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
	year, mon := dateutil.ParseYearMonth(month)
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

// ExportMD returns a month's diary entries as a downloadable Markdown file.
func (h *Handler) ExportMD(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	year, mon := dateutil.ParseYearMonth(month)
	if year == 0 {
		now := time.Now()
		year, mon = now.Year(), int(now.Month())
	}

	md, err := h.svc.ExportMonthMD(r.Context(), year, mon)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="diary-%04d-%02d.md"`, year, mon))
	w.Write([]byte(md))
}

// ImportMD accepts a Markdown file upload and imports diary entries from it.
func (h *Handler) ImportMD(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httputil.Error(w, http.StatusBadRequest, "failed to parse form (max 10MB)")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	count, err := h.svc.ImportMD(r.Context(), string(content))
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]int{"imported": count})
}


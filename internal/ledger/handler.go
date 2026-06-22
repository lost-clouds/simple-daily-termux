package ledger

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	mux.HandleFunc("GET /api/ledger", h.ListByMonth)
	mux.HandleFunc("GET /api/ledger/{id}", h.Get)
	mux.HandleFunc("GET /api/ledger/summary", h.MonthlySummary)
	mux.HandleFunc("POST /api/ledger", h.Create)
	mux.HandleFunc("DELETE /api/ledger/{id}", h.Delete)
	mux.HandleFunc("GET /api/ledger/export", h.ExportCSV)
	mux.HandleFunc("POST /api/ledger/import", h.ImportCSV)
}

func RegisterSettingsHandler(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("PUT /api/settings/{key}", h.SetSetting)
}

func (h *Handler) ListByMonth(w http.ResponseWriter, r *http.Request) {
	year, month := parseMonthDefault(r.URL.Query().Get("month"))
	entries, err := h.svc.ListByMonth(r.Context(), year, month)
	if err != nil { httputil.Error(w, 500, err.Error()); return }
	if entries == nil { entries = []*Entry{} }
	httputil.JSON(w, 200, entries)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	e, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil { httputil.Error(w, 404, "not found"); return }
	httputil.JSON(w, 200, e)
}

func (h *Handler) MonthlySummary(w http.ResponseWriter, r *http.Request) {
	year, month := parseMonthDefault(r.URL.Query().Get("month"))
	summary, err := h.svc.MonthlySummary(r.Context(), year, month)
	if err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 200, summary)
}

type createLedgerReq struct {
	EntryDate string  `json:"entry_date"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Category  string  `json:"category"`
	Note      string  `json:"note"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLedgerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { httputil.Error(w, 400, "invalid JSON"); return }
	if req.EntryDate == "" || req.Type == "" || req.Category == "" { httputil.Error(w, 400, "entry_date, type, category required"); return }
	if req.Type != TypeIncome && req.Type != TypeExpense { httputil.Error(w, 400, "type must be income or expense"); return }
	amountCents := int64(math.Round(req.Amount * 100))
	e, err := h.svc.Create(r.Context(), req.EntryDate, req.Type, amountCents, req.Category, req.Note)
	if err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 201, e)
}

func (h *Handler) SetSetting(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var req struct{ Value string `json:"value"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { httputil.Error(w, 400, "invalid JSON"); return }
	if err := h.svc.SetSetting(r.Context(), key, req.Value); err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 200, map[string]string{"key": key, "value": req.Value})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil { httputil.Error(w, 500, err.Error()); return }
	httputil.JSON(w, 200, nil)
}

// ExportCSV returns a month's ledger entries as a downloadable CSV file.
func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	year, month := parseMonthDefault(r.URL.Query().Get("month"))
	csv, err := h.svc.ExportMonthCSV(r.Context(), year, month)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="ledger-%04d-%02d.csv"`, year, month))
	w.Write([]byte(csv))
}

// ImportCSV accepts a CSV file upload and imports ledger entries from it.
func (h *Handler) ImportCSV(w http.ResponseWriter, r *http.Request) {
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

	count, err := h.svc.ImportCSV(r.Context(), string(content))
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]int{"imported": count})
}

// parseMonthDefault parses "YYYY-MM" with fallback to the current month.
func parseMonthDefault(val string) (int, int) {
	y, m := dateutil.ParseYearMonth(val)
	if y == 0 {
		now := time.Now()
		y, m = now.Year(), int(now.Month())
	}
	return y, m
}

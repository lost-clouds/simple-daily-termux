package ledger

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

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
}

func RegisterSettingsHandler(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("PUT /api/settings/{key}", h.SetSetting)
}

func (h *Handler) ListByMonth(w http.ResponseWriter, r *http.Request) {
	year, month, err := parseMonth(r.URL.Query().Get("month"))
	if err != nil { httputil.Error(w, 400, "invalid month"); return }
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
	year, month, err := parseMonth(r.URL.Query().Get("month"))
	if err != nil { httputil.Error(w, 400, "invalid month"); return }
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

func parseMonth(val string) (int, int, error) {
	if val == "" { return 0, 0, nil }
	parts := strings.SplitN(val, "-", 2)
	if len(parts) != 2 { return 0, 0, nil }
	year, err := strconv.Atoi(parts[0])
	if err != nil { return 0, 0, err }
	month, err := strconv.Atoi(parts[1])
	return year, month, err
}

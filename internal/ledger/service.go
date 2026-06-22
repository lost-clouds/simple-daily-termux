package ledger

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"simple-daily-termux/internal/idgen"
)

type Service struct {
	repo     Repository
	settings SettingsRepository
}

func NewService(r Repository, s SettingsRepository) *Service {
	return &Service{repo: r, settings: s}
}

func (s *Service) Create(ctx context.Context, entryDate, typ string, amountCents int64, category, note string) (*Entry, error) {
	now := time.Now().UTC()
	e := &Entry{
		ID:          idgen.New(),
		EntryDate:   entryDate,
		Type:        typ,
		AmountCents: amountCents,
		Amount:      float64(amountCents) / 100.0,
		Category:    category,
		Note:        note,
		Source:      SourceManual,
		CreatedAt:   now,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Entry, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) DeleteByDiarySource(ctx context.Context, diaryID string) error {
	return s.repo.DeleteBySourceDiary(ctx, diaryID)
}

func (s *Service) CreateFromDiary(ctx context.Context, diaryID, entryDate string, items []DiaryItem) error {
	now := time.Now().UTC()
	for _, item := range items {
		e := &Entry{
			ID:            idgen.New(),
			EntryDate:     entryDate,
			Type:          item.Type,
			AmountCents:   item.AmountCents,
			Amount:        float64(item.AmountCents) / 100.0,
			Category:      item.Category,
			Note:          item.Note,
			Source:        SourceDiary,
			SourceDiaryID: diaryID,
			CreatedAt:     now,
		}
		if err := s.repo.Create(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListByMonth(ctx context.Context, year, month int) ([]*Entry, error) {
	return s.repo.ListByMonth(ctx, year, month)
}

func (s *Service) MonthlySummary(ctx context.Context, year, month int) (*MonthlySummary, error) {
	summary, err := s.repo.MonthlySummary(ctx, year, month)
	if err != nil {
		return nil, err
	}

	openingStr, err := s.settings.Get(ctx, "opening_savings_cents")
	if err != nil {
		return nil, fmt.Errorf("get opening savings: %w", err)
	}
	opening, err := strconv.ParseInt(openingStr, 10, 64)
	if err != nil && openingStr != "" {
		opening = 0
	}

	cumulative, err := s.repo.CumulativeSavings(ctx, year, month)
	if err != nil {
		return nil, err
	}

	summary.Savings = float64(opening+cumulative) / 100.0
	return summary, nil
}

func (s *Service) SetSetting(ctx context.Context, key, value string) error {
	return s.settings.Set(ctx, key, value)
}

func (s *Service) SetOpeningSavings(ctx context.Context, amountCents int64) error {
	return s.settings.Set(ctx, "opening_savings_cents", fmt.Sprintf("%d", amountCents))
}

// ExportMonthCSV returns a month's ledger entries as a CSV string.
func (s *Service) ExportMonthCSV(ctx context.Context, year, month int) (string, error) {
	entries, err := s.repo.ListByMonth(ctx, year, month)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("date,type,amount,category,note\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "%s,%s,%.2f,%s,%s\n",
			e.EntryDate, e.Type, e.Amount, escapeCSV(e.Category), escapeCSV(e.Note))
	}
	return b.String(), nil
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// ImportCSV parses CSV content and creates ledger entries. Returns the number imported.
func (s *Service) ImportCSV(ctx context.Context, content string) (int, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("CSV must have a header row and at least one data row")
	}

	count := 0
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := parseCSVLine(line)
		if len(fields) < 4 {
			log.Printf("ledger: import skipping line %d: expected 4+ fields, got %d", i+1, len(fields))
			continue
		}
		date, typ, amountStr, category := fields[0], fields[1], fields[2], fields[3]
		note := ""
		if len(fields) > 4 {
			note = fields[4]
		}

		if typ != TypeIncome && typ != TypeExpense {
			log.Printf("ledger: import skipping line %d: invalid type %q", i+1, typ)
			continue
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Printf("ledger: import skipping line %d: invalid amount %q", i+1, amountStr)
			continue
		}
		amountCents := int64(math.Round(amount * 100))

		if _, err := s.Create(ctx, date, typ, amountCents, category, note); err != nil {
			log.Printf("ledger: import failed line %d: %v", i+1, err)
			continue
		}
		count++
	}
	return count, nil
}

// parseCSVLine handles simple CSV parsing with quoted fields.
func parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	for _, ch := range line {
		switch ch {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				current.WriteRune(ch)
			} else {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	fields = append(fields, current.String())
	return fields
}

type DiaryItem struct {
	Type        string
	AmountCents int64
	Category    string
	Note        string
}

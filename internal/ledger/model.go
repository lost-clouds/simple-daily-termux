package ledger

import (
	"context"
	"time"
)

const (
	TypeExpense  = "expense"
	TypeIncome   = "income"
	SourceManual = "manual"
	SourceDiary  = "diary"
)

type Entry struct {
	ID            string    `json:"id"`
	EntryDate     string    `json:"entry_date"`
	Type          string    `json:"type"`
	AmountCents   int64     `json:"amount_cents"`
	Amount        float64   `json:"amount"`
	Category      string    `json:"category"`
	Note          string    `json:"note"`
	Source        string    `json:"source"`
	SourceDiaryID string    `json:"source_diary_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type MonthlySummary struct {
	YearMonth string  `json:"year_month"`
	Income    float64 `json:"income"`
	Expense   float64 `json:"expense"`
	Balance   float64 `json:"balance"`
	Savings   float64 `json:"savings"`
}

type Repository interface {
	Create(ctx context.Context, e *Entry) error
	Get(ctx context.Context, id string) (*Entry, error)
	Delete(ctx context.Context, id string) error
	DeleteBySourceDiary(ctx context.Context, diaryID string) error
	ListByMonth(ctx context.Context, year, month int) ([]*Entry, error)
	MonthlySummary(ctx context.Context, year, month int) (*MonthlySummary, error)
	CumulativeSavings(ctx context.Context, toYear, toMonth int) (int64, error)
}

type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

package diary

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("diary entry not found")

type Entry struct {
	ID        string    `json:"id"`
	EntryDate string    `json:"entry_date"`
	ContentMD string    `json:"content_md"`
	Mood      string    `json:"mood"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MonthEntry struct {
	ID        string `json:"id"`
	EntryDate string `json:"entry_date"`
	Mood      string `json:"mood"`
}

type Repository interface {
	GetByDate(ctx context.Context, date string) (*Entry, error)
	Upsert(ctx context.Context, entry *Entry) error
	ListMonth(ctx context.Context, year, month int) ([]*MonthEntry, error)
}

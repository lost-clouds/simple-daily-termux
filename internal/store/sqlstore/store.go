// Package sqlstore implements the persistence layer for all domain aggregates
// using database/sql with SQLite or MySQL backends.
package sqlstore

import (
	"database/sql"
	"time"

	"simple-daily-termux/internal/calendar"
	"simple-daily-termux/internal/countdown"
	"simple-daily-termux/internal/diary"
	"simple-daily-termux/internal/ledger"
	"simple-daily-termux/internal/pomodoro"
	"simple-daily-termux/internal/todo"
)

// Store is the composite interface for all domain repositories.
type Store interface {
	Todos() todo.Repository
	Countdowns() countdown.Repository
	Pomodoros() pomodoro.Repository
	Diaries() diary.Repository
	Ledgers() ledger.Repository
	Calendars() calendar.Repository
	Settings() ledger.SettingsRepository
	Close() error
}

// SQLStore holds the database connection and provides repository accessors.
type SQLStore struct {
	db       *sql.DB
	driver   string
	timezone string
}

func (s *SQLStore) Close() error                        { return s.db.Close() }
func (s *SQLStore) Todos() todo.Repository               { return &todoRepo{s.db} }
func (s *SQLStore) Countdowns() countdown.Repository      { return &countdownRepo{db: s.db, timezone: s.timezone} }
func (s *SQLStore) Pomodoros() pomodoro.Repository        { return &pomodoroRepo{s.db} }
func (s *SQLStore) Diaries() diary.Repository             { return &diaryRepo{s.db, s.driver} }
func (s *SQLStore) Ledgers() ledger.Repository            { return &ledgerRepo{s.db} }
func (s *SQLStore) Calendars() calendar.Repository        { return &calendarRepo{s.db} }
func (s *SQLStore) Settings() ledger.SettingsRepository   { return &settingsRepo{s.db, s.driver} }

// --- shared helpers ---

func nt(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func pt(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t, _ := time.Parse(time.RFC3339, ns.String)
	return &t
}

func pt2(ns sql.NullString) time.Time {
	if !ns.Valid || ns.String == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, ns.String)
	return t
}

// daysUntil returns the number of calendar days from today to target, using the
// configured timezone (or UTC if unset).
func daysUntil(target time.Time, timezone string) int {
	loc := time.UTC
	if tz, err := time.LoadLocation(timezone); err == nil {
		loc = tz
	}
	now := time.Now().In(loc)
	tb := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, loc)
	td := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return int(tb.Sub(td).Hours() / 24)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"simple-daily-termux/internal/diary"
)

type diaryRepo struct {
	db     *sql.DB
	driver string
}

func (r *diaryRepo) GetByDate(ctx context.Context, date string) (*diary.Entry, error) {
	e := &diary.Entry{}
	var ca, ua sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,entry_date,content_md,mood,created_at,updated_at FROM diary_entries WHERE entry_date=?`, date,
	).Scan(&e.ID, &e.EntryDate, &e.ContentMD, &e.Mood, &ca, &ua)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.CreatedAt = pt2(ca)
	e.UpdatedAt = pt2(ua)
	return e, nil
}
func (r *diaryRepo) Upsert(ctx context.Context, e *diary.Entry) error {
	var err error
	if r.driver == "mysql" {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO diary_entries (id,entry_date,content_md,mood,created_at,updated_at) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE content_md=VALUES(content_md),mood=VALUES(mood),updated_at=VALUES(updated_at)`,
			e.ID, e.EntryDate, e.ContentMD, e.Mood, e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339))
	} else {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO diary_entries (id,entry_date,content_md,mood,created_at,updated_at) VALUES (?,?,?,?,?,?) ON CONFLICT(entry_date) DO UPDATE SET content_md=?,mood=?,updated_at=?`,
			e.ID, e.EntryDate, e.ContentMD, e.Mood, e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339),
			e.ContentMD, e.Mood, e.UpdatedAt.Format(time.RFC3339))
	}
	return err
}
func (r *diaryRepo) ListMonth(ctx context.Context, year, month int) ([]*diary.MonthEntry, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := r.db.QueryContext(ctx, `SELECT id,entry_date,mood FROM diary_entries WHERE entry_date LIKE ? ORDER BY entry_date ASC`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []*diary.MonthEntry
	for rows.Next() {
		e := &diary.MonthEntry{}
		if err := rows.Scan(&e.ID, &e.EntryDate, &e.Mood); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListMonthFull returns full diary entries (including content_md) for a month.
func (r *diaryRepo) ListMonthFull(ctx context.Context, year, month int) ([]*diary.Entry, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,entry_date,content_md,mood,created_at,updated_at FROM diary_entries WHERE entry_date LIKE ? ORDER BY entry_date ASC`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []*diary.Entry
	for rows.Next() {
		e := &diary.Entry{}
		var ca, ua sql.NullString
		if err := rows.Scan(&e.ID, &e.EntryDate, &e.ContentMD, &e.Mood, &ca, &ua); err != nil {
			return nil, err
		}
		e.CreatedAt = pt2(ca)
		e.UpdatedAt = pt2(ua)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

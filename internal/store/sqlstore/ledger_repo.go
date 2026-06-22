package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"simple-daily-termux/internal/dateutil"
	"simple-daily-termux/internal/ledger"
)

type ledgerRepo struct{ db *sql.DB }

func (r *ledgerRepo) Create(ctx context.Context, e *ledger.Entry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ledger_entries (id,entry_date,type,amount_cents,category,note,source,source_diary_id,created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		e.ID, e.EntryDate, e.Type, e.AmountCents, e.Category, e.Note, e.Source, e.SourceDiaryID, e.CreatedAt.Format(time.RFC3339))
	return err
}
func (r *ledgerRepo) Get(ctx context.Context, id string) (*ledger.Entry, error) {
	e := &ledger.Entry{}
	var ca sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,entry_date,type,amount_cents,category,note,source,source_diary_id,created_at FROM ledger_entries WHERE id=?`, id,
	).Scan(&e.ID, &e.EntryDate, &e.Type, &e.AmountCents, &e.Category, &e.Note, &e.Source, &e.SourceDiaryID, &ca)
	if err != nil {
		return nil, err
	}
	e.Amount = float64(e.AmountCents) / 100.0
	e.CreatedAt = pt2(ca)
	return e, nil
}
func (r *ledgerRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ledger_entries WHERE id=?`, id)
	return err
}
func (r *ledgerRepo) DeleteBySourceDiary(ctx context.Context, diaryID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ledger_entries WHERE source=? AND source_diary_id=?`, ledger.SourceDiary, diaryID)
	return err
}
func (r *ledgerRepo) ListByMonth(ctx context.Context, year, month int) ([]*ledger.Entry, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,entry_date,type,amount_cents,category,note,source,source_diary_id,created_at FROM ledger_entries WHERE entry_date LIKE ? ORDER BY created_at DESC`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []*ledger.Entry
	for rows.Next() {
		e := &ledger.Entry{}
		var ca sql.NullString
		if err := rows.Scan(&e.ID, &e.EntryDate, &e.Type, &e.AmountCents, &e.Category, &e.Note, &e.Source, &e.SourceDiaryID, &ca); err != nil {
			return nil, err
		}
		e.Amount = float64(e.AmountCents) / 100.0
		e.CreatedAt = pt2(ca)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
func (r *ledgerRepo) MonthlySummary(ctx context.Context, year, month int) (*ledger.MonthlySummary, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	s := &ledger.MonthlySummary{YearMonth: prefix}
	var income, expense sql.NullInt64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date LIKE ? AND type=?`, prefix+"%", ledger.TypeIncome,
	).Scan(&income); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date LIKE ? AND type=?`, prefix+"%", ledger.TypeExpense,
	).Scan(&expense); err != nil {
		return nil, err
	}
	s.Income = float64(income.Int64) / 100.0
	s.Expense = float64(expense.Int64) / 100.0
	s.Balance = s.Income - s.Expense
	return s, nil
}
func (r *ledgerRepo) CumulativeSavings(ctx context.Context, toYear, toMonth int) (int64, error) {
	lastDay := dateutil.LastDayOfMonth(toYear, toMonth)
	toDate := fmt.Sprintf("%04d-%02d-%02d", toYear, toMonth, lastDay)
	var income, expense sql.NullInt64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date <= ? AND type=?`, toDate, ledger.TypeIncome).Scan(&income); err != nil {
		return 0, err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date <= ? AND type=?`, toDate, ledger.TypeExpense).Scan(&expense); err != nil {
		return 0, err
	}
	return income.Int64 - expense.Int64, nil
}

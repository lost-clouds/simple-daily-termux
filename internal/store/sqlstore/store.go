package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"simple-daily-termux/internal/calendar"
	"simple-daily-termux/internal/countdown"
	"simple-daily-termux/internal/diary"
	"simple-daily-termux/internal/ledger"
	"simple-daily-termux/internal/pomodoro"
	"simple-daily-termux/internal/todo"
)

// Store is the aggregate store interface composited from all repository interfaces.
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

type SQLStore struct {
	db     *sql.DB
	driver string
}

func (s *SQLStore) Close() error { return s.db.Close() }

func (s *SQLStore) Todos() todo.Repository                { return &todoRepo{s.db} }
func (s *SQLStore) Countdowns() countdown.Repository       { return &countdownRepo{s.db} }
func (s *SQLStore) Pomodoros() pomodoro.Repository         { return &pomodoroRepo{s.db} }
func (s *SQLStore) Diaries() diary.Repository              { return &diaryRepo{s.db, s.driver} }
func (s *SQLStore) Ledgers() ledger.Repository             { return &ledgerRepo{s.db} }
func (s *SQLStore) Calendars() calendar.Repository         { return &calendarRepo{s.db} }
func (s *SQLStore) Settings() ledger.SettingsRepository    { return &settingsRepo{s.db, s.driver} }

// --- helpers ---

func scanTime(s string) (time.Time, error) { return time.Parse(time.RFC3339, s) }

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
	t, err := scanTime(ns.String)
	if err != nil {
		return nil
	}
	return &t
}

func pt2(ns sql.NullString) time.Time {
	if !ns.Valid || ns.String == "" {
		return time.Time{}
	}
	t, _ := scanTime(ns.String)
	return t
}

func daysUntil(target time.Time) int {
	now := time.Now()
	targetDay := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return int(targetDay.Sub(today).Hours() / 24)
}

// --- todoRepo ---

type todoRepo struct{ db *sql.DB }

func (r *todoRepo) Create(ctx context.Context, t *todo.Todo) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO todos (id,title,notes,status,priority,deadline_at,created_at,updated_at,completed_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		t.ID, t.Title, t.Notes, t.Status, t.Priority,
		nt(t.DeadlineAt), t.CreatedAt.Format(time.RFC3339),
		t.UpdatedAt.Format(time.RFC3339), nt(t.CompletedAt),
	)
	return err
}

func (r *todoRepo) Get(ctx context.Context, id string) (*todo.Todo, error) {
	t := &todo.Todo{}
	var dl, co, ca, ua sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,title,notes,status,priority,deadline_at,created_at,updated_at,completed_at FROM todos WHERE id=?`, id,
	).Scan(&t.ID, &t.Title, &t.Notes, &t.Status, &t.Priority, &dl, &ca, &ua, &co)
	if err != nil {
		return nil, err
	}
	t.DeadlineAt = pt(dl)
	t.CompletedAt = pt(co)
	t.CreatedAt = pt2(ca)
	t.UpdatedAt = pt2(ua)
	return t, nil
}

func (r *todoRepo) Update(ctx context.Context, t *todo.Todo) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE todos SET title=?,notes=?,status=?,priority=?,deadline_at=?,updated_at=?,completed_at=? WHERE id=?`,
		t.Title, t.Notes, t.Status, t.Priority, nt(t.DeadlineAt),
		t.UpdatedAt.Format(time.RFC3339), nt(t.CompletedAt), t.ID,
	)
	return err
}

func (r *todoRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE id=?`, id)
	return err
}

func (r *todoRepo) List(ctx context.Context, f todo.ListFilter) ([]*todo.Todo, error) {
	q := `SELECT id,title,notes,status,priority,deadline_at,created_at,updated_at,completed_at FROM todos WHERE 1=1`
	args := []any{}
	if f.Status != "" {
		q += " AND status=?"
		args = append(args, f.Status)
	}
	if f.HasDeadline == "true" {
		q += " AND deadline_at IS NOT NULL"
	} else if f.HasDeadline == "false" {
		q += " AND deadline_at IS NULL"
	}
	q += " ORDER BY priority DESC, created_at DESC"
	if f.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, f.Limit)
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*todo.Todo
	for rows.Next() {
		t := &todo.Todo{}
		var dl, co, ca, ua sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Notes, &t.Status, &t.Priority, &dl, &ca, &ua, &co); err != nil {
			return nil, err
		}
		t.DeadlineAt = pt(dl)
		t.CompletedAt = pt(co)
		t.CreatedAt = pt2(ca)
		t.UpdatedAt = pt2(ua)
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

// --- countdownRepo ---

type countdownRepo struct{ db *sql.DB }

func (r *countdownRepo) Create(ctx context.Context, e *countdown.Event) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO countdown_events (id,title,target_at,source,ref_id,note,created_at) VALUES (?,?,?,?,?,?,?)`,
		e.ID, e.Title, e.TargetAt.Format(time.RFC3339), e.Source, e.RefID, e.Note, e.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *countdownRepo) Get(ctx context.Context, id string) (*countdown.Event, error) {
	e := &countdown.Event{}
	var target, ca sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,title,target_at,source,ref_id,note,created_at FROM countdown_events WHERE id=?`, id,
	).Scan(&e.ID, &e.Title, &target, &e.Source, &e.RefID, &e.Note, &ca)
	if err != nil {
		return nil, err
	}
	e.TargetAt = pt2(target)
	e.CreatedAt = pt2(ca)
	e.DaysLeft = daysUntil(e.TargetAt)
	return e, nil
}

func (r *countdownRepo) Update(ctx context.Context, e *countdown.Event) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE countdown_events SET title=?,target_at=?,note=? WHERE id=?`,
		e.Title, e.TargetAt.Format(time.RFC3339), e.Note, e.ID,
	)
	return err
}

func (r *countdownRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM countdown_events WHERE id=?`, id)
	return err
}

func (r *countdownRepo) DeleteByRef(ctx context.Context, refID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM countdown_events WHERE source=? AND ref_id=?`, countdown.SourceTodo, refID)
	return err
}

func (r *countdownRepo) List(ctx context.Context) ([]*countdown.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,title,target_at,source,ref_id,note,created_at FROM countdown_events ORDER BY target_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*countdown.Event
	for rows.Next() {
		e := &countdown.Event{}
		var target, ca sql.NullString
		if err := rows.Scan(&e.ID, &e.Title, &target, &e.Source, &e.RefID, &e.Note, &ca); err != nil {
			return nil, err
		}
		e.TargetAt = pt2(target)
		e.CreatedAt = pt2(ca)
		e.DaysLeft = daysUntil(e.TargetAt)
		events = append(events, e)
	}
	return events, rows.Err()
}

// --- pomodoroRepo ---

type pomodoroRepo struct{ db *sql.DB }

func (r *pomodoroRepo) Create(ctx context.Context, s *pomodoro.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO pomodoro_sessions (id,started_at,ended_at,planned_minutes,actual_minutes,status,linked_todo_id) VALUES (?,?,?,?,?,?,?)`,
		s.ID, s.StartedAt.Format(time.RFC3339), nt(s.EndedAt),
		s.PlannedMinutes, s.ActualMinutes, s.Status, s.LinkedTodoID,
	)
	return err
}

func (r *pomodoroRepo) Update(ctx context.Context, s *pomodoro.Session) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE pomodoro_sessions SET ended_at=?,actual_minutes=?,status=? WHERE id=?`,
		nt(s.EndedAt), s.ActualMinutes, s.Status, s.ID,
	)
	return err
}

func (r *pomodoroRepo) Get(ctx context.Context, id string) (*pomodoro.Session, error) {
	s := &pomodoro.Session{}
	var sa, en sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,started_at,ended_at,planned_minutes,actual_minutes,status,linked_todo_id FROM pomodoro_sessions WHERE id=?`, id,
	).Scan(&s.ID, &sa, &en, &s.PlannedMinutes, &s.ActualMinutes, &s.Status, &s.LinkedTodoID)
	if err != nil {
		return nil, err
	}
	s.StartedAt = pt2(sa)
	s.EndedAt = pt(en)
	return s, nil
}

func (r *pomodoroRepo) GetTodayMinutes(ctx context.Context, loc string) (int, error) {
	tz, err := time.LoadLocation(loc)
	if err != nil {
		tz = time.Local
	}
	now := time.Now().In(tz)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)

	var total sql.NullInt64
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(actual_minutes),0) FROM pomodoro_sessions WHERE status=? AND started_at >= ?`,
		pomodoro.StatusCompleted, today.Format(time.RFC3339),
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return int(total.Int64), nil
}

func (r *pomodoroRepo) ListRange(ctx context.Context, from, to string) ([]*pomodoro.Session, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,started_at,ended_at,planned_minutes,actual_minutes,status,linked_todo_id
		 FROM pomodoro_sessions WHERE started_at >= ? AND started_at < ? ORDER BY started_at DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*pomodoro.Session
	for rows.Next() {
		s := &pomodoro.Session{}
		var sa, en sql.NullString
		if err := rows.Scan(&s.ID, &sa, &en, &s.PlannedMinutes, &s.ActualMinutes, &s.Status, &s.LinkedTodoID); err != nil {
			return nil, err
		}
		s.StartedAt = pt2(sa)
		s.EndedAt = pt(en)
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// --- diaryRepo ---

type diaryRepo struct{ db *sql.DB; driver string }

func (r *diaryRepo) GetByDate(ctx context.Context, date string) (*diary.Entry, error) {
	e := &diary.Entry{}
	var ca, ua sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,entry_date,content_md,mood,created_at,updated_at FROM diary_entries WHERE entry_date=?`, date,
	).Scan(&e.ID, &e.EntryDate, &e.ContentMD, &e.Mood, &ca, &ua)
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
			`INSERT INTO diary_entries (id,entry_date,content_md,mood,created_at,updated_at)
			 VALUES (?,?,?,?,?,?)
			 ON DUPLICATE KEY UPDATE content_md=VALUES(content_md),mood=VALUES(mood),updated_at=VALUES(updated_at)`,
			e.ID, e.EntryDate, e.ContentMD, e.Mood,
			e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339),
		)
	} else {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO diary_entries (id,entry_date,content_md,mood,created_at,updated_at)
			 VALUES (?,?,?,?,?,?)
			 ON CONFLICT(entry_date) DO UPDATE SET content_md=?,mood=?,updated_at=?`,
			e.ID, e.EntryDate, e.ContentMD, e.Mood,
			e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339),
			e.ContentMD, e.Mood, e.UpdatedAt.Format(time.RFC3339),
		)
	}
	return err
}

func (r *diaryRepo) ListMonth(ctx context.Context, year, month int) ([]*diary.MonthEntry, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,entry_date,mood FROM diary_entries WHERE entry_date LIKE ? ORDER BY entry_date ASC`, prefix+"%")
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

// --- ledgerRepo ---

type ledgerRepo struct{ db *sql.DB }

func (r *ledgerRepo) Create(ctx context.Context, e *ledger.Entry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ledger_entries (id,entry_date,type,amount_cents,category,note,source,source_diary_id,created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		e.ID, e.EntryDate, e.Type, e.AmountCents, e.Category, e.Note, e.Source, e.SourceDiaryID, e.CreatedAt.Format(time.RFC3339),
	)
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
		`SELECT id,entry_date,type,amount_cents,category,note,source,source_diary_id,created_at
		 FROM ledger_entries WHERE entry_date LIKE ? ORDER BY created_at DESC`, prefix+"%")
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
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date LIKE ? AND type=?`, prefix+"%", ledger.TypeIncome,
	).Scan(&income)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date LIKE ? AND type=?`, prefix+"%", ledger.TypeExpense,
	).Scan(&expense)
	if err != nil {
		return nil, err
	}

	s.Income = float64(income.Int64) / 100.0
	s.Expense = float64(expense.Int64) / 100.0
	s.Balance = s.Income - s.Expense
	return s, nil
}

func (r *ledgerRepo) CumulativeSavings(ctx context.Context, toYear, toMonth int) (int64, error) {
	toPrefix := fmt.Sprintf("%04d-%02d", toYear, toMonth)
	var income, expense sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date <= ? AND type=?`, toPrefix+"-31", ledger.TypeIncome,
	).Scan(&income)
	if err != nil {
		return 0, err
	}
	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries WHERE entry_date <= ? AND type=?`, toPrefix+"-31", ledger.TypeExpense,
	).Scan(&expense)
	if err != nil {
		return 0, err
	}
	return income.Int64 - expense.Int64, nil
}

// --- calendarRepo ---

type calendarRepo struct{ db *sql.DB }

func (r *calendarRepo) CreateEvent(ctx context.Context, e *calendar.Event) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO calendar_events (id,title,start_at,end_at,all_day,note,created_at) VALUES (?,?,?,?,?,?,?)`,
		e.ID, e.Title, e.StartAt.Format(time.RFC3339), nt(e.EndAt),
		boolToInt(e.AllDay), e.Note, e.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *calendarRepo) UpdateEvent(ctx context.Context, e *calendar.Event) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE calendar_events SET title=?,start_at=?,end_at=?,all_day=?,note=? WHERE id=?`,
		e.Title, e.StartAt.Format(time.RFC3339), nt(e.EndAt), boolToInt(e.AllDay), e.Note, e.ID,
	)
	return err
}

func (r *calendarRepo) DeleteEvent(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM calendar_events WHERE id=?`, id)
	return err
}

func (r *calendarRepo) ListEvents(ctx context.Context, year, month int) ([]*calendar.Event, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,title,start_at,end_at,all_day,note,created_at
		 FROM calendar_events WHERE start_at LIKE ? ORDER BY start_at ASC`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*calendar.Event
	for rows.Next() {
		e := &calendar.Event{}
		var sa, en, ca sql.NullString
		var ad int
		if err := rows.Scan(&e.ID, &e.Title, &sa, &en, &ad, &e.Note, &ca); err != nil {
			return nil, err
		}
		e.StartAt = pt2(sa)
		e.EndAt = pt(en)
		e.AllDay = ad == 1
		e.CreatedAt = pt2(ca)
		events = append(events, e)
	}
	return events, rows.Err()
}

// --- settingsRepo ---

type settingsRepo struct{ db *sql.DB; driver string }

func (r *settingsRepo) Get(ctx context.Context, key string) (string, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE `+"`key`"+`=?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v, err
}

func (r *settingsRepo) Set(ctx context.Context, key, value string) error {
	var err error
	if r.driver == "mysql" {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO settings (`+"`key`"+`,value) VALUES (?,?) ON DUPLICATE KEY UPDATE value=VALUES(value)`, key, value,
		)
	} else {
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO settings (key,value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=?`, key, value, value,
		)
	}
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

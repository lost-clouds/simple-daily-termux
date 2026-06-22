package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"simple-daily-termux/internal/idgen"
	"simple-daily-termux/internal/todo"
)

type todoRepo struct{ db *sql.DB }

func (r *todoRepo) Create(ctx context.Context, t *todo.Todo) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO todos (id,title,notes,status,task_type,priority,deadline_at,entry_date,created_at,updated_at,completed_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.Title, t.Notes, t.Status, t.TaskType, t.Priority, nt(t.DeadlineAt), t.EntryDate,
		t.CreatedAt.Format(time.RFC3339), t.UpdatedAt.Format(time.RFC3339), nt(t.CompletedAt))
	return err
}
func (r *todoRepo) Get(ctx context.Context, id string) (*todo.Todo, error) {
	t := &todo.Todo{}
	var dl, co, ca, ua sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,title,notes,status,task_type,priority,deadline_at,entry_date,created_at,updated_at,completed_at FROM todos WHERE id=?`, id,
	).Scan(&t.ID, &t.Title, &t.Notes, &t.Status, &t.TaskType, &t.Priority, &dl, &t.EntryDate, &ca, &ua, &co)
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
		`UPDATE todos SET title=?,notes=?,status=?,task_type=?,priority=?,deadline_at=?,entry_date=?,updated_at=?,completed_at=? WHERE id=?`,
		t.Title, t.Notes, t.Status, t.TaskType, t.Priority, nt(t.DeadlineAt), t.EntryDate,
		t.UpdatedAt.Format(time.RFC3339), nt(t.CompletedAt), t.ID)
	return err
}
func (r *todoRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE id=?`, id)
	return err
}
func (r *todoRepo) List(ctx context.Context, f todo.ListFilter) ([]*todo.Todo, error) {
	q := `SELECT id,title,notes,status,task_type,priority,deadline_at,entry_date,created_at,updated_at,completed_at FROM todos WHERE 1=1`
	args := []any{}
	if f.Status != "" {
		q += " AND status=?"
		args = append(args, f.Status)
	}
	if f.TaskType != "" {
		q += " AND task_type=?"
		args = append(args, f.TaskType)
	}
	if f.EntryDate != "" {
		q += " AND (entry_date=? OR (task_type='daily' AND status!='done'))"
		args = append(args, f.EntryDate)
	}
	if f.HasDeadline == "true" {
		q += " AND deadline_at IS NOT NULL"
	} else if f.HasDeadline == "false" {
		q += " AND deadline_at IS NULL"
	}
	q += " ORDER BY priority ASC, created_at DESC"
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
		if err := rows.Scan(&t.ID, &t.Title, &t.Notes, &t.Status, &t.TaskType, &t.Priority, &dl, &t.EntryDate, &ca, &ua, &co); err != nil {
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

// EnsureDailyTasks creates copies of active daily/long_term templates for the given date.
// Long-term tasks are backfilled: any missing days from the last copy through today
// (or the deadline, whichever is earlier) are created.
func (r *todoRepo) EnsureDailyTasks(ctx context.Context, today string) ([]*todo.Todo, error) {
	var created []*todo.Todo

	// 1. Long-term tasks — processed independently of daily tasks
	ltCreated, err := r.ensureLongTermCopies(ctx, today)
	if err != nil {
		return nil, err
	}
	created = append(created, ltCreated...)

	// 2. Daily tasks
	dailyCreated, err := r.ensureDailyCopies(ctx, today)
	if err != nil {
		return nil, err
	}
	created = append(created, dailyCreated...)

	return created, nil
}

// ensureLongTermCopies backfills daily copies of long_term tasks from the last existing
// copy date through today (capped by the deadline).
func (r *todoRepo) ensureLongTermCopies(ctx context.Context, today string) ([]*todo.Todo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, notes, deadline_at, entry_date FROM todos
		 WHERE task_type='long_term' AND status!='done' AND deadline_at IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type ltTemplate struct {
		id         string
		title      string
		notes      string
		entryDate  string
		deadlineAt *time.Time
	}
	var templates []ltTemplate
	for rows.Next() {
		var lt ltTemplate
		var dl sql.NullString
		if err := rows.Scan(&lt.id, &lt.title, &lt.notes, &dl, &lt.entryDate); err != nil {
			continue
		}
		lt.deadlineAt = pt(dl)
		if lt.deadlineAt == nil {
			continue
		}
		templates = append(templates, lt)
	}

	var created []*todo.Todo
	now := time.Now()
	todayTime, _ := time.Parse("2006-01-02", today)

	for _, tmpl := range templates {
		deadlineStr := tmpl.deadlineAt.Format("2006-01-02")

		// Find the latest copy's entry_date for this template
		var lastDate sql.NullString
		r.db.QueryRowContext(ctx,
			`SELECT MAX(entry_date) FROM todos WHERE title=? AND task_type='long_term' AND entry_date <= ?`,
			tmpl.title, today).Scan(&lastDate)

		// Determine the start date for backfill
		var startTime time.Time
		if lastDate.Valid && lastDate.String != "" {
			t, err := time.Parse("2006-01-02", lastDate.String)
			if err == nil {
				startTime = t.AddDate(0, 0, 1) // day after last copy
			} else {
				startTime, _ = time.Parse("2006-01-02", tmpl.entryDate)
			}
		} else {
			startTime, _ = time.Parse("2006-01-02", tmpl.entryDate)
		}
		if startTime.IsZero() {
			continue
		}

		// Determine the end date: min(deadline, today)
		endTime := todayTime
		deadlineTime, err := time.Parse("2006-01-02", deadlineStr)
		if err == nil && deadlineTime.Before(todayTime) {
			endTime = deadlineTime
		}

		// Create copies for each missing day
		for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
			ds := d.Format("2006-01-02")
			// Double-check no copy exists for this date
			var count int
			r.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM todos WHERE entry_date=? AND title=? AND task_type='long_term'`,
				ds, tmpl.title).Scan(&count)
			if count > 0 {
				continue
			}
			pri := todo.CalcPriority(tmpl.deadlineAt, todo.TaskLongTerm)
			deadlineCopy := *tmpl.deadlineAt
			t := &todo.Todo{
				ID:         idgen.New(),
				Title:      tmpl.title,
				Notes:      tmpl.notes,
				Status:     todo.StatusPending,
				TaskType:   todo.TaskLongTerm,
				Priority:   pri,
				DeadlineAt: &deadlineCopy,
				EntryDate:  ds,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := r.Create(ctx, t); err == nil {
				created = append(created, t)
			}
		}
	}
	return created, nil
}

// ensureDailyCopies creates today's copies of daily task templates if they don't exist yet.
func (r *todoRepo) ensureDailyCopies(ctx context.Context, today string) ([]*todo.Todo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, notes, priority, deadline_at FROM todos WHERE task_type='daily' AND status!='done'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type dailyTemplate struct {
		id         string
		title      string
		notes      string
		priority   int
		deadlineAt *time.Time
	}
	var templates []dailyTemplate
	for rows.Next() {
		var d dailyTemplate
		var dl sql.NullString
		if err := rows.Scan(&d.id, &d.title, &d.notes, &d.priority, &dl); err != nil {
			return nil, err
		}
		d.deadlineAt = pt(dl)
		templates = append(templates, d)
	}

	var created []*todo.Todo
	now := time.Now().UTC()
	for _, tmpl := range templates {
		var count int
		r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM todos WHERE entry_date=? AND title=? AND task_type='daily'`,
			today, tmpl.title).Scan(&count)
		if count > 0 {
			continue
		}

		t := &todo.Todo{
			ID:         idgen.New(),
			Title:      tmpl.title,
			Notes:      tmpl.notes,
			Status:     todo.StatusPending,
			TaskType:   todo.TaskDaily,
			Priority:   todo.PriorityIV,
			EntryDate:  today,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := r.Create(ctx, t); err == nil {
			created = append(created, t)
		}
	}
	return created, nil
}

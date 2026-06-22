package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"simple-daily-termux/internal/pomodoro"
)

type pomodoroRepo struct{ db *sql.DB }

func (r *pomodoroRepo) Create(ctx context.Context, s *pomodoro.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO pomodoro_sessions (id,started_at,ended_at,planned_minutes,actual_minutes,status,session_type,linked_todo_id) VALUES (?,?,?,?,?,?,?,?)`,
		s.ID, s.StartedAt.Format(time.RFC3339), nt(s.EndedAt), s.PlannedMinutes, s.ActualMinutes, s.Status, s.SessionType, s.LinkedTodoID)
	return err
}
func (r *pomodoroRepo) Update(ctx context.Context, s *pomodoro.Session) error {
	_, err := r.db.ExecContext(ctx, `UPDATE pomodoro_sessions SET ended_at=?,actual_minutes=?,status=?,session_type=? WHERE id=?`,
		nt(s.EndedAt), s.ActualMinutes, s.Status, s.SessionType, s.ID)
	return err
}
func (r *pomodoroRepo) Get(ctx context.Context, id string) (*pomodoro.Session, error) {
	s := &pomodoro.Session{}
	var sa, en sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,started_at,ended_at,planned_minutes,actual_minutes,status,session_type,linked_todo_id FROM pomodoro_sessions WHERE id=?`, id,
	).Scan(&s.ID, &sa, &en, &s.PlannedMinutes, &s.ActualMinutes, &s.Status, &s.SessionType, &s.LinkedTodoID)
	if err != nil {
		return nil, err
	}
	s.StartedAt = pt2(sa)
	s.EndedAt = pt(en)
	return s, nil
}
func (r *pomodoroRepo) GetTodayMinutes(ctx context.Context, loc string) (int, error) {
	return r.getTodayMinutesByType(ctx, loc, pomodoro.TypeFocus)
}
func (r *pomodoroRepo) GetTodayRestMinutes(ctx context.Context, loc string) (int, error) {
	return r.getTodayMinutesByType(ctx, loc, pomodoro.TypeRest)
}

// getTodayMinutesByType returns total completed minutes of a given session type today (UTC).
func (r *pomodoroRepo) getTodayMinutesByType(ctx context.Context, loc, stype string) (int, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var total sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(actual_minutes),0) FROM pomodoro_sessions WHERE status=? AND session_type=? AND started_at >= ?`,
		pomodoro.StatusCompleted, stype, today.Format(time.RFC3339)).Scan(&total)
	if err != nil {
		return 0, err
	}
	return int(total.Int64), nil
}
func (r *pomodoroRepo) ListRange(ctx context.Context, from, to string) ([]*pomodoro.Session, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,started_at,ended_at,planned_minutes,actual_minutes,status,session_type,linked_todo_id
		 FROM pomodoro_sessions WHERE started_at >= ? AND started_at < ? ORDER BY started_at DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*pomodoro.Session
	for rows.Next() {
		s := &pomodoro.Session{}
		var sa, en sql.NullString
		if err := rows.Scan(&s.ID, &sa, &en, &s.PlannedMinutes, &s.ActualMinutes, &s.Status, &s.SessionType, &s.LinkedTodoID); err != nil {
			return nil, err
		}
		s.StartedAt = pt2(sa)
		s.EndedAt = pt(en)
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

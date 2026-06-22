package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"simple-daily-termux/internal/calendar"
)

type calendarRepo struct{ db *sql.DB }

func (r *calendarRepo) CreateEvent(ctx context.Context, e *calendar.Event) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO calendar_events (id,title,start_at,end_at,all_day,note,created_at) VALUES (?,?,?,?,?,?,?)`,
		e.ID, e.Title, e.StartAt.Format(time.RFC3339), nt(e.EndAt), boolToInt(e.AllDay), e.Note, e.CreatedAt.Format(time.RFC3339))
	return err
}
func (r *calendarRepo) UpdateEvent(ctx context.Context, e *calendar.Event) error {
	_, err := r.db.ExecContext(ctx, `UPDATE calendar_events SET title=?,start_at=?,end_at=?,all_day=?,note=? WHERE id=?`,
		e.Title, e.StartAt.Format(time.RFC3339), nt(e.EndAt), boolToInt(e.AllDay), e.Note, e.ID)
	return err
}
func (r *calendarRepo) DeleteEvent(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM calendar_events WHERE id=?`, id)
	return err
}
func (r *calendarRepo) ListEvents(ctx context.Context, year, month int) ([]*calendar.Event, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,title,start_at,end_at,all_day,note,created_at FROM calendar_events WHERE start_at LIKE ? ORDER BY start_at ASC`, prefix+"%")
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

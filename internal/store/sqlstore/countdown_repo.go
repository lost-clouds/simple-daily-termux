package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"simple-daily-termux/internal/countdown"
)

type countdownRepo struct {
	db       *sql.DB
	timezone string
}

func (r *countdownRepo) Create(ctx context.Context, e *countdown.Event) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO countdown_events (id,title,target_at,source,ref_id,note,created_at) VALUES (?,?,?,?,?,?,?)`,
		e.ID, e.Title, e.TargetAt.Format(time.RFC3339), e.Source, e.RefID, e.Note, e.CreatedAt.Format(time.RFC3339))
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
	e.DaysLeft = daysUntil(e.TargetAt, r.timezone)
	return e, nil
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
	rows, err := r.db.QueryContext(ctx, `SELECT id,title,target_at,source,ref_id,note,created_at FROM countdown_events ORDER BY target_at ASC`)
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
		e.DaysLeft = daysUntil(e.TargetAt, r.timezone)
		events = append(events, e)
	}
	return events, rows.Err()
}

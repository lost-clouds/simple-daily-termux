package pomodoro

import (
	"context"
	"time"
)

const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusAborted   = "aborted"

	TypeFocus = "focus"
	TypeRest  = "rest"
)

type Session struct {
	ID             string     `json:"id"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
	PlannedMinutes int        `json:"planned_minutes"`
	ActualMinutes  int        `json:"actual_minutes"`
	Status         string     `json:"status"`
	SessionType    string     `json:"session_type"` // focus / rest
	LinkedTodoID   string     `json:"linked_todo_id"`
}

type Repository interface {
	Create(ctx context.Context, s *Session) error
	Update(ctx context.Context, s *Session) error
	Get(ctx context.Context, id string) (*Session, error)
	GetTodayMinutes(ctx context.Context, loc string) (int, error)
	GetTodayRestMinutes(ctx context.Context, loc string) (int, error)
	ListRange(ctx context.Context, from, to string) ([]*Session, error)
}

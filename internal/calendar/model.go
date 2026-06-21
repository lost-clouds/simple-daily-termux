package calendar

import (
	"context"
	"time"
)

type Event struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	StartAt   time.Time  `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
	AllDay    bool       `json:"all_day"`
	Note      string     `json:"note"`
	CreatedAt time.Time  `json:"created_at"`
}

type TodoDeadline struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	DeadlineAt *time.Time `json:"deadline_at"`
	Status     string     `json:"status"`
}

type CountdownTarget struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	TargetAt  time.Time `json:"target_at"`
	DaysLeft  int       `json:"days_left"`
}

type MonthView struct {
	YearMonth        string            `json:"year_month"`
	Events           []*Event          `json:"events"`
	TodoDeadlines    []*TodoDeadline   `json:"todo_deadlines"`
	CountdownTargets []*CountdownTarget `json:"countdown_targets"`
	DiaryDates       []string          `json:"diary_dates"`
}

type Repository interface {
	CreateEvent(ctx context.Context, e *Event) error
	UpdateEvent(ctx context.Context, e *Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEvents(ctx context.Context, year, month int) ([]*Event, error)
}

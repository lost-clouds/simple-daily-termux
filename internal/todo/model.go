package todo

import (
	"context"
	"time"
)

const (
	StatusPending = "pending"
	StatusDoing   = "doing"
	StatusDone    = "done"

	TaskDaily    = "daily"
	TaskLongTerm = "long_term"
	TaskOneTime  = "one_time"

	PriorityI   = 1 // red:    < 14 days
	PriorityII  = 2 // orange: 14-30 days
	PriorityIII = 3 // yellow: 30-60 days
	PriorityIV  = 4 // blue:   >= 60 days or daily tasks
)

type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Notes       string     `json:"notes"`
	Status      string     `json:"status"`
	TaskType    string     `json:"task_type"` // daily / long_term / one_time
	Priority    int        `json:"priority"`  // auto-calculated
	DeadlineAt  *time.Time `json:"deadline_at"`
	EntryDate   string     `json:"entry_date"` // YYYY-MM-DD, for daily tasks
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type ListFilter struct {
	Status      string
	TaskType    string
	HasDeadline string
	EntryDate   string // YYYY-MM-DD, for filtering today's tasks
	Limit       int
}

type Repository interface {
	Create(ctx context.Context, t *Todo) error
	Get(ctx context.Context, id string) (*Todo, error)
	Update(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, f ListFilter) ([]*Todo, error)
	EnsureDailyTasks(ctx context.Context, today string) ([]*Todo, error)
}

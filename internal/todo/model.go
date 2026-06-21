package todo

import (
	"context"
	"time"
)

const (
	StatusPending = "pending"
	StatusDoing   = "doing"
	StatusDone    = "done"
)

type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Notes       string     `json:"notes"`
	Status      string     `json:"status"`
	Priority    int        `json:"priority"`
	DeadlineAt  *time.Time `json:"deadline_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type ListFilter struct {
	Status      string
	HasDeadline string // "true" | "false" | ""
	Limit       int
}

type Repository interface {
	Create(ctx context.Context, t *Todo) error
	Get(ctx context.Context, id string) (*Todo, error)
	Update(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, f ListFilter) ([]*Todo, error)
}

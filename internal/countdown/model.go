package countdown

import (
	"context"
	"time"
)

const (
	SourceManual = "manual"
	SourceTodo   = "todo"
)

type Event struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	TargetAt  time.Time `json:"target_at"`
	Source    string    `json:"source"`
	RefID     string    `json:"ref_id"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	DaysLeft  int       `json:"days_left"`
}

type Repository interface {
	Create(ctx context.Context, e *Event) error
	Get(ctx context.Context, id string) (*Event, error)
	Delete(ctx context.Context, id string) error
	DeleteByRef(ctx context.Context, refID string) error
	List(ctx context.Context) ([]*Event, error)
}

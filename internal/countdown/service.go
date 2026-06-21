package countdown

import (
	"context"
	"time"

	"simple-daily-termux/internal/idgen"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) List(ctx context.Context) ([]*Event, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, title string, targetAt time.Time, note string) (*Event, error) {
	now := time.Now()
	e := &Event{
		ID:        idgen.New(),
		Title:     title,
		TargetAt:  targetAt,
		Source:    SourceManual,
		Note:      note,
		CreatedAt: now,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) SyncFromTodo(ctx context.Context, todoID, title string, targetAt time.Time) error {
	_ = s.repo.DeleteByRef(ctx, todoID)

	now := time.Now()
	e := &Event{
		ID:        idgen.New(),
		Title:     title,
		TargetAt:  targetAt,
		Source:    SourceTodo,
		RefID:     todoID,
		CreatedAt: now,
	}
	return s.repo.Create(ctx, e)
}

func (s *Service) RemoveByRef(ctx context.Context, todoID string) error {
	return s.repo.DeleteByRef(ctx, todoID)
}
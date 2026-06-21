package todo

import (
	"context"
	"log"
	"time"

	"simple-daily-termux/internal/countdown"
	"simple-daily-termux/internal/idgen"
)

type Service struct {
	repo     Repository
	countSvc *countdown.Service
}

func NewService(r Repository, countSvc *countdown.Service) *Service {
	return &Service{repo: r, countSvc: countSvc}
}

func (s *Service) Create(ctx context.Context, title, notes string, priority int, deadlineAt *time.Time) (*Todo, error) {
	now := time.Now()
	t := &Todo{
		ID:         idgen.New(),
		Title:      title,
		Notes:      notes,
		Status:     StatusPending,
		Priority:   priority,
		DeadlineAt: deadlineAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	if deadlineAt != nil {
		if err := s.countSvc.SyncFromTodo(ctx, t.ID, t.Title, *deadlineAt); err != nil {
			log.Printf("todo: countdown sync failed (create %s): %v", t.ID, err)
		}
	}
	return t, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Todo, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, t *Todo) error {
	t.UpdatedAt = time.Now()
	if t.Status == StatusDone && t.CompletedAt == nil {
		now := time.Now()
		t.CompletedAt = &now
	}

	old, err := s.repo.Get(ctx, t.ID)
	if err != nil {
		return err
	}

	if err := s.repo.Update(ctx, t); err != nil {
		return err
	}

	// Handle deadline linkage changes
	oldHadDeadline := old.DeadlineAt != nil
	newHasDeadline := t.DeadlineAt != nil

	if !newHasDeadline && oldHadDeadline {
		if err := s.countSvc.RemoveByRef(ctx, t.ID); err != nil {
			log.Printf("todo: countdown remove failed (update %s): %v", t.ID, err)
		}
	} else if newHasDeadline {
		if err := s.countSvc.SyncFromTodo(ctx, t.ID, t.Title, *t.DeadlineAt); err != nil {
			log.Printf("todo: countdown sync failed (update %s): %v", t.ID, err)
		}
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if t.DeadlineAt != nil {
		if err := s.countSvc.RemoveByRef(ctx, id); err != nil {
			log.Printf("todo: countdown remove failed (delete %s): %v", id, err)
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*Todo, error) {
	return s.repo.List(ctx, f)
}

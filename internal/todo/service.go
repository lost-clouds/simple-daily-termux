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

func (s *Service) Create(ctx context.Context, title, notes, taskType string, deadlineAt *time.Time) (*Todo, error) {
	now := time.Now()
	pri := CalcPriority(deadlineAt, taskType)
	t := &Todo{
		ID: idgen.New(), Title: title, Notes: notes, Status: StatusPending,
		TaskType: taskType, Priority: pri, DeadlineAt: deadlineAt,
		EntryDate: now.Format("2006-01-02"), CreatedAt: now, UpdatedAt: now,
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
		now := time.Now(); t.CompletedAt = &now
	}
	t.Priority = CalcPriority(t.DeadlineAt, t.TaskType)
	old, err := s.repo.Get(ctx, t.ID)
	if err != nil { return err }
	if err := s.repo.Update(ctx, t); err != nil { return err }
	oldHad := old.DeadlineAt != nil
	newHas := t.DeadlineAt != nil
	if !newHas && oldHad {
		if err := s.countSvc.RemoveByRef(ctx, t.ID); err != nil {
			log.Printf("todo: countdown remove failed: %v", err)
		}
	} else if newHas {
		if err := s.countSvc.SyncFromTodo(ctx, t.ID, t.Title, *t.DeadlineAt); err != nil {
			log.Printf("todo: countdown sync failed: %v", err)
		}
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	t, err := s.repo.Get(ctx, id)
	if err != nil { return err }
	if err := s.repo.Delete(ctx, id); err != nil { return err }
	if t.DeadlineAt != nil {
		if err := s.countSvc.RemoveByRef(ctx, id); err != nil {
			log.Printf("todo: countdown remove failed: %v", err)
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*Todo, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) EnsureDailyTasks(ctx context.Context, today string) ([]*Todo, error) {
	return s.repo.EnsureDailyTasks(ctx, today)
}

func CalcPriority(deadline *time.Time, taskType string) int {
	if taskType == TaskDaily { return PriorityIV }
	if deadline == nil { return PriorityIV }
	d := int(time.Until(*deadline).Hours() / 24)
	switch {
	case d < 14: return PriorityI
	case d < 30: return PriorityII
	case d < 60: return PriorityIII
	default: return PriorityIV
	}
}

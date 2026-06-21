package pomodoro

import (
	"context"
	"time"

	"simple-daily-termux/internal/idgen"
)

type Service struct{ repo Repository }

func NewService(r Repository) *Service { return &Service{repo: r} }

func (s *Service) Start(ctx context.Context, plannedMinutes int, sessionType, linkedTodoID string) (*Session, error) {
	now := time.Now()
	session := &Session{
		ID: idgen.New(), StartedAt: now, PlannedMinutes: plannedMinutes,
		Status: StatusRunning, SessionType: sessionType, LinkedTodoID: linkedTodoID,
	}
	if err := s.repo.Create(ctx, session); err != nil { return nil, err }
	return session, nil
}

func (s *Service) Finish(ctx context.Context, id, status string) (*Session, error) {
	session, err := s.repo.Get(ctx, id)
	if err != nil { return nil, err }
	now := time.Now()
	session.EndedAt = &now; session.Status = status
	session.ActualMinutes = int(now.Sub(session.StartedAt).Minutes())
	if session.ActualMinutes < 1 { session.ActualMinutes = 1 }
	if err := s.repo.Update(ctx, session); err != nil { return nil, err }
	return session, nil
}

func (s *Service) GetTodayMinutes(ctx context.Context, tz string) (int, error) {
	return s.repo.GetTodayMinutes(ctx, tz)
}

func (s *Service) GetTodayRestMinutes(ctx context.Context, tz string) (int, error) {
	return s.repo.GetTodayRestMinutes(ctx, tz)
}

func (s *Service) ListRange(ctx context.Context, from, to string) ([]*Session, error) {
	return s.repo.ListRange(ctx, from, to)
}

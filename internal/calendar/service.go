package calendar

import (
	"context"
	"fmt"
	"time"

	"simple-daily-termux/internal/countdown"
	"simple-daily-termux/internal/diary"
	"simple-daily-termux/internal/idgen"
	"simple-daily-termux/internal/todo"
)

type Service struct {
	calRepo   Repository
	todoRepo  todo.Repository
	countRepo countdown.Repository
	diaryRepo diary.Repository
}

func NewService(
	cr Repository,
	tr todo.Repository,
	ctr countdown.Repository,
	dr diary.Repository,
) *Service {
	return &Service{
		calRepo:   cr,
		todoRepo:  tr,
		countRepo: ctr,
		diaryRepo: dr,
	}
}

func (s *Service) GetMonthView(ctx context.Context, year, month int) (*MonthView, error) {
	calEvents, err := s.calRepo.ListEvents(ctx, year, month)
	if err != nil {
		return nil, err
	}
	if calEvents == nil {
		calEvents = []*Event{}
	}

	allTodos, err := s.todoRepo.List(ctx, todo.ListFilter{HasDeadline: "true"})
	if err != nil {
		return nil, err
	}

	var todoDeadlines []*TodoDeadline
	for _, t := range allTodos {
		if t.DeadlineAt == nil || t.Status == todo.StatusDone {
			continue
		}
		if t.DeadlineAt.Year() == year && int(t.DeadlineAt.Month()) == month {
			todoDeadlines = append(todoDeadlines, &TodoDeadline{
				ID:         t.ID,
				Title:      t.Title,
				DeadlineAt: t.DeadlineAt,
				Status:     t.Status,
			})
		}
	}
	if todoDeadlines == nil {
		todoDeadlines = []*TodoDeadline{}
	}

	allCountdowns, err := s.countRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	var countdownTargets []*CountdownTarget
	for _, c := range allCountdowns {
		if c.TargetAt.Year() == year && int(c.TargetAt.Month()) == month {
			countdownTargets = append(countdownTargets, &CountdownTarget{
				ID:       c.ID,
				Title:    c.Title,
				TargetAt: c.TargetAt,
				DaysLeft: c.DaysLeft,
			})
		}
	}
	if countdownTargets == nil {
		countdownTargets = []*CountdownTarget{}
	}

	diaryMonthEntries, err := s.diaryRepo.ListMonth(ctx, year, month)
	if err != nil {
		return nil, err
	}

	var diaryDates []string
	for _, e := range diaryMonthEntries {
		diaryDates = append(diaryDates, e.EntryDate)
	}
	if diaryDates == nil {
		diaryDates = []string{}
	}

	return &MonthView{
		YearMonth:        fmt.Sprintf("%04d-%02d", year, month),
		Events:           calEvents,
		TodoDeadlines:    todoDeadlines,
		CountdownTargets: countdownTargets,
		DiaryDates:       diaryDates,
	}, nil
}

func (s *Service) CreateEvent(ctx context.Context, title string, startAt time.Time, endAt *time.Time, allDay bool, note string) (*Event, error) {
	now := time.Now()
	e := &Event{
		ID:        idgen.New(),
		Title:     title,
		StartAt:   startAt,
		EndAt:     endAt,
		AllDay:    allDay,
		Note:      note,
		CreatedAt: now,
	}
	if err := s.calRepo.CreateEvent(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) UpdateEvent(ctx context.Context, id, title string, startAt time.Time, endAt *time.Time, allDay bool, note string) (*Event, error) {
	e := &Event{
		ID:        id,
		Title:     title,
		StartAt:   startAt,
		EndAt:     endAt,
		AllDay:    allDay,
		Note:      note,
	}
	if err := s.calRepo.UpdateEvent(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) DeleteEvent(ctx context.Context, id string) error {
	return s.calRepo.DeleteEvent(ctx, id)
}

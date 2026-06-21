package summary

import (
	"context"
	"time"

	"simple-daily-termux/internal/countdown"
	"simple-daily-termux/internal/ledger"
	"simple-daily-termux/internal/pomodoro"
)

type Service struct {
	ledgerSvc *ledger.Service
	countSvc  *countdown.Service
	pomoSvc   *pomodoro.Service
	timezone  string
}

func NewService(ledgerSvc *ledger.Service, countSvc *countdown.Service, pomoSvc *pomodoro.Service, timezone string) *Service {
	return &Service{
		ledgerSvc: ledgerSvc,
		countSvc:  countSvc,
		pomoSvc:   pomoSvc,
		timezone:  timezone,
	}
}

type SummaryData struct {
	Calendar         CalendarInfo       `json:"calendar"`
	Ledger           ledger.MonthlySummary `json:"ledger"`
	Countdown        []*countdown.Event    `json:"countdown"`
	FocusTodayMinutes int                `json:"focus_today_minutes"`
}

type CalendarInfo struct {
	Month string `json:"month"`
	Today string `json:"today"`
}

func (s *Service) GetSummary(ctx context.Context) (*SummaryData, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	month := now.Format("2006-01")

	// Get monthly ledger summary
	ledgerSummary, err := s.ledgerSvc.MonthlySummary(ctx, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}

	// Get active countdown events (up to 5)
	events, err := s.countSvc.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(events) > 5 {
		events = events[:5]
	}
	if events == nil {
		events = []*countdown.Event{}
	}

	// Get today's focus minutes
	focusMinutes, err := s.pomoSvc.GetTodayMinutes(ctx, s.timezone)
	if err != nil {
		focusMinutes = 0
	}

	return &SummaryData{
		Calendar: CalendarInfo{
			Month: month,
			Today: today,
		},
		Ledger:            *ledgerSummary,
		Countdown:         events,
		FocusTodayMinutes: focusMinutes,
	}, nil
}

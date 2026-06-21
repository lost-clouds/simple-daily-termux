package ledger

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"simple-daily-termux/internal/idgen"
)

type Service struct {
	repo     Repository
	settings SettingsRepository
}

func NewService(r Repository, s SettingsRepository) *Service {
	return &Service{repo: r, settings: s}
}

func (s *Service) Create(ctx context.Context, entryDate, typ string, amountCents int64, category, note string) (*Entry, error) {
	now := time.Now()
	e := &Entry{
		ID:          idgen.New(),
		EntryDate:   entryDate,
		Type:        typ,
		AmountCents: amountCents,
		Amount:      float64(amountCents) / 100.0,
		Category:    category,
		Note:        note,
		Source:      SourceManual,
		CreatedAt:   now,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Entry, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) DeleteByDiarySource(ctx context.Context, diaryID string) error {
	return s.repo.DeleteBySourceDiary(ctx, diaryID)
}

func (s *Service) CreateFromDiary(ctx context.Context, diaryID, entryDate string, items []DiaryItem) error {
	now := time.Now()
	for _, item := range items {
		e := &Entry{
			ID:            idgen.New(),
			EntryDate:     entryDate,
			Type:          item.Type,
			AmountCents:   item.AmountCents,
			Amount:        float64(item.AmountCents) / 100.0,
			Category:      item.Category,
			Note:          item.Note,
			Source:        SourceDiary,
			SourceDiaryID: diaryID,
			CreatedAt:     now,
		}
		if err := s.repo.Create(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListByMonth(ctx context.Context, year, month int) ([]*Entry, error) {
	return s.repo.ListByMonth(ctx, year, month)
}

func (s *Service) MonthlySummary(ctx context.Context, year, month int) (*MonthlySummary, error) {
	summary, err := s.repo.MonthlySummary(ctx, year, month)
	if err != nil {
		return nil, err
	}

	openingStr, err := s.settings.Get(ctx, "opening_savings_cents")
	if err != nil {
		return nil, fmt.Errorf("get opening savings: %w", err)
	}
	opening, _ := strconv.ParseInt(openingStr, 10, 64)

	cumulative, err := s.repo.CumulativeSavings(ctx, year, month)
	if err != nil {
		return nil, err
	}

	summary.Savings = float64(opening+cumulative) / 100.0
	return summary, nil
}

func (s *Service) SetOpeningSavings(ctx context.Context, amountCents int64) error {
	return s.settings.Set(ctx, "opening_savings_cents", fmt.Sprintf("%d", amountCents))
}

type DiaryItem struct {
	Type        string
	AmountCents int64
	Category    string
	Note        string
}

package diary

import (
	"context"
	"database/sql"
	"log"
	"time"

	"simple-daily-termux/internal/idgen"
	"simple-daily-termux/internal/ledger"
)

type Service struct {
	repo      Repository
	ledgerSvc *ledger.Service
}

func NewService(r Repository, ledgerSvc *ledger.Service) *Service {
	return &Service{repo: r, ledgerSvc: ledgerSvc}
}

func (s *Service) Get(ctx context.Context, date string) (*Entry, error) {
	e, err := s.repo.GetByDate(ctx, date)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return e, err
}

func (s *Service) Save(ctx context.Context, entryDate, contentMD, mood string) (*Entry, error) {
	now := time.Now()

	existing, err := s.repo.GetByDate(ctx, entryDate)
	var e *Entry
	if err == nil {
		e = existing
		e.ContentMD = contentMD
		e.Mood = mood
		e.UpdatedAt = now
	} else {
		e = &Entry{
			ID:        idgen.New(),
			EntryDate: entryDate,
			ContentMD: contentMD,
			Mood:      mood,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	if err := s.repo.Upsert(ctx, e); err != nil {
		return nil, err
	}

	parsedEntries := ParseLedgerBlocks(contentMD)
	if len(parsedEntries) > 0 {
		if err := s.ledgerSvc.DeleteByDiarySource(ctx, e.ID); err != nil {
			log.Printf("diary: ledger delete failed (diary %s): %v", e.ID, err)
		}
		items := make([]ledger.DiaryItem, len(parsedEntries))
		for i, pe := range parsedEntries {
			items[i] = ledger.DiaryItem{
				Type:        pe.Type,
				AmountCents: pe.AmountCents,
				Category:    pe.Category,
				Note:        pe.Note,
			}
		}
		if err := s.ledgerSvc.CreateFromDiary(ctx, e.ID, entryDate, items); err != nil {
			log.Printf("diary: ledger create failed (diary %s): %v", e.ID, err)
		}
	} else {
		if err := s.ledgerSvc.DeleteByDiarySource(ctx, e.ID); err != nil {
			log.Printf("diary: ledger delete failed (diary %s): %v", e.ID, err)
		}
	}

	return e, nil
}

func (s *Service) ListMonth(ctx context.Context, year, month int) ([]*MonthEntry, error) {
	return s.repo.ListMonth(ctx, year, month)
}

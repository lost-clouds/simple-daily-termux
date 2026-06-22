package diary

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
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
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

func (s *Service) Save(ctx context.Context, entryDate, contentMD, mood string) (*Entry, error) {
	now := time.Now().UTC()
	existing, err := s.repo.GetByDate(ctx, entryDate)
	if err != nil {
		return nil, err
	}

	var e *Entry
	if existing != nil {
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
	var syncErr error
	if len(parsedEntries) > 0 {
		if err := s.ledgerSvc.DeleteByDiarySource(ctx, e.ID); err != nil {
			log.Printf("diary: ledger delete failed (diary %s): %v", e.ID, err)
			syncErr = fmt.Errorf("ledger sync: delete old: %w", err)
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
			if syncErr == nil {
				syncErr = fmt.Errorf("ledger sync: create: %w", err)
			}
		}
	} else {
		if err := s.ledgerSvc.DeleteByDiarySource(ctx, e.ID); err != nil {
			log.Printf("diary: ledger delete failed (diary %s): %v", e.ID, err)
			syncErr = fmt.Errorf("ledger sync: delete: %w", err)
		}
	}

	return e, syncErr
}

func (s *Service) ListMonth(ctx context.Context, year, month int) ([]*MonthEntry, error) {
	return s.repo.ListMonth(ctx, year, month)
}

// ExportMonthMD assembles a month's diary entries into a Markdown document.
// Each entry is a section starting with "# YYYY-MM-DD", a "心情: X" line, then the content.
// Sections are separated by "\n---\n".
func (s *Service) ExportMonthMD(ctx context.Context, year, month int) (string, error) {
	entries, err := s.repo.ListMonthFull(ctx, year, month)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&b, "# %s\n", e.EntryDate)
		if e.Mood != "" {
			fmt.Fprintf(&b, "心情: %s\n", e.Mood)
		}
		b.WriteString("\n")
		b.WriteString(e.ContentMD)
		b.WriteString("\n")
	}
	return b.String(), nil
}

// importMDEntryRe matches a date header line like "# 2026-06-22".
var importMDEntryRe = regexp.MustCompile(`(?m)^# (\d{4}-\d{2}-\d{2})\s*$`)

// ImportMD parses a Markdown document and upserts each section as a diary entry.
// Returns the number of successfully imported entries.
func (s *Service) ImportMD(ctx context.Context, content string) (int, error) {
	// Find all date headers and their positions.
	matches := importMDEntryRe.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("no valid date headers found (expected '# YYYY-MM-DD')")
	}

	count := 0
	for i, match := range matches {
		date := content[match[2]:match[3]]
		// Section content starts after the header line, ends at next header or EOF.
		sectionStart := match[1] // end of the full match
		var sectionEnd int
		if i+1 < len(matches) {
			sectionEnd = matches[i+1][0]
		} else {
			sectionEnd = len(content)
		}
		section := strings.TrimSpace(content[sectionStart:sectionEnd])

		// Extract mood line if present.
		mood := ""
		moodRe := regexp.MustCompile(`^心情:\s*(.+)$`)
		if loc := moodRe.FindStringSubmatch(section); loc != nil {
			mood = strings.TrimSpace(loc[1])
			section = strings.TrimSpace(strings.Replace(section, loc[0], "", 1))
		}

		// The remaining text (after removing separator lines "---") is the content.
		section = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(section), "---"))

		_, err := s.Save(ctx, date, section, mood)
		if err != nil {
			log.Printf("diary: import failed for %s: %v", date, err)
			continue
		}
		count++
	}
	return count, nil
}

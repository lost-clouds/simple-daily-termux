package diary

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type ParsedLedgerEntry struct {
	Type        string // income | expense
	AmountCents int64
	Category    string
	Note        string
}

var ledgerBlockRe = regexp.MustCompile("(?s)```ledger\\s*\\n(.*?)```")

func ParseLedgerBlocks(content string) []ParsedLedgerEntry {
	var entries []ParsedLedgerEntry
	matches := ledgerBlockRe.FindAllStringSubmatch(content, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		entry := parseLedgerBlock(strings.TrimSpace(m[1]))
		if entry != nil {
			entries = append(entries, *entry)
		}
	}
	return entries
}

func parseLedgerBlock(block string) *ParsedLedgerEntry {
	e := &ParsedLedgerEntry{}
	hasRequired := false
	lines := strings.Split(block, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "type":
			v := strings.ToLower(val)
			if v == "income" || v == "expense" {
				e.Type = v
				hasRequired = true
			}
		case "amount":
			val = strings.ReplaceAll(val, ",", "")
			f, err := strconv.ParseFloat(val, 64)
			if err == nil && f >= 0 {
				e.AmountCents = int64(math.Round(f * 100))
			}
		case "category":
			e.Category = val
		case "note":
			e.Note = val
		}
	}

	if !hasRequired || e.AmountCents == 0 || e.Category == "" {
		log.Printf("ledgerparser: skipping invalid block (type=%q amount=%d category=%q)", e.Type, e.AmountCents, e.Category)
		return nil
	}
	return e
}

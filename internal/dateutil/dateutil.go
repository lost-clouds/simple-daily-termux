// Package dateutil provides shared date/time helpers used across handlers.
package dateutil

import (
	"strconv"
	"strings"
	"time"
)

// ParseYearMonth parses a "YYYY-MM" string into year and month ints.
// Returns (0, 0) for empty or malformed input; callers should fall back to time.Now().
func ParseYearMonth(val string) (int, int) {
	if val == "" {
		return 0, 0
	}
	parts := strings.SplitN(val, "-", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	y, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return y, m
}

// LastDayOfMonth returns the last calendar day of the given year/month.
func LastDayOfMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// FormatYearMonth formats year and month as "YYYY-MM".
func FormatYearMonth(year, month int) string {
	// Use a fixed-width format via time package to avoid manual padding.
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return t.Format("2006-01")
}

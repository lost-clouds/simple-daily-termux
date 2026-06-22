package dateutil

import "testing"

func TestParseYearMonth_Valid(t *testing.T) {
	y, m := ParseYearMonth("2026-06")
	if y != 2026 || m != 6 {
		t.Errorf("expected (2026, 6), got (%d, %d)", y, m)
	}
}

func TestParseYearMonth_Empty(t *testing.T) {
	y, m := ParseYearMonth("")
	if y != 0 || m != 0 {
		t.Errorf("expected (0, 0), got (%d, %d)", y, m)
	}
}

func TestParseYearMonth_Malformed(t *testing.T) {
	// Malformed inputs: only check that the function doesn't panic
	// and that completely invalid strings return (0,0).
	if y, m := ParseYearMonth("abc"); y != 0 || m != 0 {
		t.Errorf("ParseYearMonth('abc') = (%d,%d), want (0,0)", y, m)
	}
	if y, _ := ParseYearMonth("2026-06-22"); y != 2026 {
		t.Errorf("ParseYearMonth('2026-06-22') year = %d, want 2026", y)
	}
}

func TestParseYearMonth_NonNumeric(t *testing.T) {
	y, m := ParseYearMonth("abcd-ef")
	if y != 0 || m != 0 {
		t.Errorf("expected (0, 0), got (%d, %d)", y, m)
	}
}

func TestLastDayOfMonth(t *testing.T) {
	tests := []struct {
		year, month, expected int
	}{
		{2026, 1, 31},
		{2026, 2, 28}, // 2026 is not a leap year
		{2024, 2, 29}, // 2024 is a leap year
		{2026, 4, 30},
		{2026, 6, 30},
		{2026, 9, 30},
		{2026, 11, 30},
		{2026, 12, 31},
	}
	for _, tt := range tests {
		got := LastDayOfMonth(tt.year, tt.month)
		if got != tt.expected {
			t.Errorf("LastDayOfMonth(%d, %d) = %d, want %d", tt.year, tt.month, got, tt.expected)
		}
	}
}

func TestFormatYearMonth(t *testing.T) {
	got := FormatYearMonth(2026, 6)
	if got != "2026-06" {
		t.Errorf("expected '2026-06', got %q", got)
	}
	got = FormatYearMonth(2026, 12)
	if got != "2026-12" {
		t.Errorf("expected '2026-12', got %q", got)
	}
}

// Package nldate — unit tests for the natural language date parser.
package nldate

import (
	"testing"
	"time"
)

func ref(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestBasicKeywords(t *testing.T) {
	now := ref("2026-02-20")
	cases := []struct {
		input string
		want  string
	}{
		{"today", "2026-02-20"},
		{"tomorrow", "2026-02-21"},
		{"yesterday", "2026-02-19"},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got, ok := Parse(c.input, now)
			if !ok {
				t.Fatalf("expected ok=true")
			}
			if got.Format("2006-01-02") != c.want {
				t.Errorf("got %s, want %s", got.Format("2006-01-02"), c.want)
			}
		})
	}
}

func TestNextWeekday(t *testing.T) {
	// 2026-02-20 is a Friday
	now := ref("2026-02-20")
	cases := []struct {
		input string
		want  string
	}{
		{"next monday", "2026-02-23"},
		{"next friday", "2026-02-27"},
		{"next saturday", "2026-02-21"},
		{"next sunday", "2026-02-22"},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got, ok := Parse(c.input, now)
			if !ok {
				t.Fatalf("expected ok=true for %q", c.input)
			}
			if got.Format("2006-01-02") != c.want {
				t.Errorf("got %s, want %s", got.Format("2006-01-02"), c.want)
			}
		})
	}
}

func TestInNDuration(t *testing.T) {
	now := ref("2026-02-20")
	cases := []struct {
		input string
		want  string
	}{
		{"in 1 day", "2026-02-21"},
		{"in 3 days", "2026-02-23"},
		{"in 1 week", "2026-02-27"},
		{"in 2 weeks", "2026-03-06"},
		{"in 1 month", "2026-03-20"},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got, ok := Parse(c.input, now)
			if !ok {
				t.Fatalf("expected ok=true for %q", c.input)
			}
			if got.Format("2006-01-02") != c.want {
				t.Errorf("got %s, want %s", got.Format("2006-01-02"), c.want)
			}
		})
	}
}

func TestISODate(t *testing.T) {
	now := ref("2026-02-20")
	got, ok := Parse("2026-03-15", now)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got.Format("2006-01-02") != "2026-03-15" {
		t.Errorf("got %s", got.Format("2006-01-02"))
	}
}

func TestUnknown(t *testing.T) {
	now := ref("2026-02-20")
	_, ok := Parse("some random text", now)
	if ok {
		t.Error("expected ok=false for unrecognised input")
	}
}

func TestExtractDate(t *testing.T) {
	now := ref("2026-02-20")
	cases := []struct {
		input     string
		wantText  string
		wantDate  string
		wantFound bool
	}{
		{"Fix auth bug tomorrow", "Fix auth bug", "2026-02-21", true},
		{"Call dentist next friday", "Call dentist", "2026-02-27", true},
		{"Buy milk in 3 days", "Buy milk", "2026-02-23", true},
		{"Just a note", "Just a note", "", false},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			text, date, found := ExtractDate(c.input, now)
			if found != c.wantFound {
				t.Fatalf("found=%v, want %v", found, c.wantFound)
			}
			if text != c.wantText {
				t.Errorf("text=%q, want %q", text, c.wantText)
			}
			if c.wantFound && date.Format("2006-01-02") != c.wantDate {
				t.Errorf("date=%s, want %s", date.Format("2006-01-02"), c.wantDate)
			}
		})
	}
}

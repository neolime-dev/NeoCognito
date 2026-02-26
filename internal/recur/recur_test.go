package recur_test

import (
	"testing"
	"time"

	"github.com/neolime-dev/neocognito/internal/recur"
)

// friday is a convenient fixed reference date: 2026-02-20 (Friday) in UTC.
var friday = time.Date(2026, 2, 20, 12, 30, 0, 0, time.UTC)

// --- Parse ---

func TestParse_ValidRules(t *testing.T) {
	valid := []string{
		"daily", "weekly", "monthly", "yearly",
		"every-monday", "every-friday", "every-saturday",
		"every-mon", "every-fri", "every-sat",
		"every-2-days", "every-3-weeks", "every-6-months",
	}
	for _, s := range valid {
		if _, err := recur.Parse(s); err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", s, err)
		}
	}
}

func TestParse_InvalidRules(t *testing.T) {
	invalid := []string{"", "every-fortnight", "biweekly", "never"}
	for _, s := range invalid {
		if _, err := recur.Parse(s); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", s)
		}
	}
}

// --- Rule.NextDue builtins ---

func TestRule_NextDue_Daily(t *testing.T) {
	r, _ := recur.Parse("daily")
	got, err := r.NextDue(friday)
	if err != nil {
		t.Fatalf("NextDue: %v", err)
	}
	want := date(2026, 2, 21)
	assertDate(t, "daily", want, got)
}

func TestRule_NextDue_Weekly(t *testing.T) {
	r, _ := recur.Parse("weekly")
	got, _ := r.NextDue(friday)
	assertDate(t, "weekly", date(2026, 2, 27), got)
}

func TestRule_NextDue_Monthly(t *testing.T) {
	r, _ := recur.Parse("monthly")
	got, _ := r.NextDue(friday)
	assertDate(t, "monthly", date(2026, 3, 20), got)
}

func TestRule_NextDue_Yearly(t *testing.T) {
	r, _ := recur.Parse("yearly")
	got, _ := r.NextDue(friday)
	assertDate(t, "yearly", date(2027, 2, 20), got)
}

// --- every-N-days / weeks / months ---

func TestRule_NextDue_EveryNDays(t *testing.T) {
	cases := []struct {
		rule string
		want time.Time
	}{
		{"every-2-days", date(2026, 2, 22)},
		{"every-10-days", date(2026, 3, 2)},
	}
	for _, tc := range cases {
		r, _ := recur.Parse(tc.rule)
		got, _ := r.NextDue(friday)
		assertDate(t, tc.rule, tc.want, got)
	}
}

func TestRule_NextDue_EveryNWeeks(t *testing.T) {
	r, _ := recur.Parse("every-2-weeks")
	got, _ := r.NextDue(friday)
	assertDate(t, "every-2-weeks", date(2026, 3, 6), got)
}

func TestRule_NextDue_EveryNMonths(t *testing.T) {
	r, _ := recur.Parse("every-3-months")
	got, _ := r.NextDue(friday)
	assertDate(t, "every-3-months", date(2026, 5, 20), got)
}

// --- Weekday rules ---

func TestRule_NextDue_Weekdays(t *testing.T) {
	// from is Friday 2026-02-20
	cases := []struct {
		rule string
		want time.Time
	}{
		{"every-monday", date(2026, 2, 23)},   // next Mon
		{"every-saturday", date(2026, 2, 21)}, // next day
		{"every-friday", date(2026, 2, 27)},   // same weekday → next week
	}
	for _, tc := range cases {
		r, _ := recur.Parse(tc.rule)
		got, _ := r.NextDue(friday)
		assertDate(t, tc.rule, tc.want, got)
	}
}

func TestRule_NextDue_ShortWeekdayNames(t *testing.T) {
	pairs := [][2]string{
		{"every-monday", "every-mon"},
		{"every-friday", "every-fri"},
		{"every-saturday", "every-sat"},
	}
	for _, p := range pairs {
		long, _ := recur.Parse(p[0])
		short, _ := recur.Parse(p[1])
		gl, _ := long.NextDue(friday)
		gs, _ := short.NextDue(friday)
		if !gl.Equal(gs) {
			t.Errorf("%s vs %s: %v != %v", p[0], p[1], gl, gs)
		}
	}
}

// --- Package-level NextDue ---

func TestNextDue_PackageLevel(t *testing.T) {
	got, err := recur.NextDue("daily", friday)
	if err != nil {
		t.Fatalf("NextDue: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	assertDate(t, "daily (package-level)", date(2026, 2, 21), *got)
}

// --- StartOfDay preserves timezone ---

func TestRule_NextDue_StartOfDay(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Skip("America/Sao_Paulo timezone not available:", err)
	}
	// Use noon in São Paulo — after midnight UTC
	noon := time.Date(2026, 2, 20, 12, 0, 0, 0, loc)
	r, _ := recur.Parse("daily")
	got, _ := r.NextDue(noon)

	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
		t.Errorf("expected start of day, got %v", got)
	}
	if got.Location().String() != loc.String() {
		t.Errorf("timezone lost: want %s, got %s", loc, got.Location())
	}
}

// --- helpers ---

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func assertDate(t *testing.T, label string, want, got time.Time) {
	t.Helper()
	// Compare only the date portion (year/month/day) in the result's own timezone.
	wy, wm, wd := want.Date()
	gy, gm, gd := got.Date()
	if wy != gy || wm != gm || wd != gd {
		t.Errorf("%s: want %s, got %s", label, want.Format("2006-01-02"), got.Format("2006-01-02"))
	}
}

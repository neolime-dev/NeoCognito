// Package nldate parses natural language date expressions into time.Time.
package nldate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Parse attempts to interpret s as a natural language date relative to now.
// Returns (parsed time, true) on success, or (zero, false) if s is not a
// recognisable date expression.
//
// Supported patterns (case-insensitive):
//
//	today, tomorrow, yesterday
//	next monday … next sunday
//	in N days / in N weeks / in N months
//	YYYY-MM-DD
func Parse(s string, now time.Time) (time.Time, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return time.Time{}, false
	}

	// Absolute ISO date
	if t, err := time.ParseInLocation("2006-01-02", s, now.Location()); err == nil {
		return t, true
	}

	switch s {
	case "today":
		return startOfDay(now), true
	case "tomorrow":
		return startOfDay(now.AddDate(0, 0, 1)), true
	case "yesterday":
		return startOfDay(now.AddDate(0, 0, -1)), true
	}

	// "next <weekday>"
	if strings.HasPrefix(s, "next ") {
		day := strings.TrimPrefix(s, "next ")
		if t, ok := nextWeekday(now, day); ok {
			return t, true
		}
	}

	// "in N days/weeks/months"
	if strings.HasPrefix(s, "in ") {
		rest := strings.TrimPrefix(s, "in ")
		parts := strings.Fields(rest)
		if len(parts) == 2 {
			n, err := strconv.Atoi(parts[0])
			if err == nil {
				unit := strings.TrimRight(parts[1], "s") // normalise plural
				switch unit {
				case "day":
					return startOfDay(now.AddDate(0, 0, n)), true
				case "week":
					return startOfDay(now.AddDate(0, 0, n*7)), true
				case "month":
					return startOfDay(now.AddDate(0, n, 0)), true
				}
			}
		}
	}

	// "N days" / "N weeks" / "N months" (without "in")
	parts := strings.Fields(s)
	if len(parts) == 2 {
		n, err := strconv.Atoi(parts[0])
		if err == nil {
			unit := strings.TrimRight(parts[1], "s")
			switch unit {
			case "day":
				return startOfDay(now.AddDate(0, 0, n)), true
			case "week":
				return startOfDay(now.AddDate(0, 0, n*7)), true
			case "month":
				return startOfDay(now.AddDate(0, n, 0)), true
			}
		}
	}

	return time.Time{}, false
}

// ExtractDate scans text for a trailing date expression and returns the
// cleaned text (without the date) and the parsed date.
// E.g.: "Fix auth bug tomorrow" → ("Fix auth bug", <tomorrow>, true)
func ExtractDate(text string, now time.Time) (string, *time.Time, bool) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text, nil, false
	}

	// Try the last word
	last := words[len(words)-1]
	if t, ok := Parse(last, now); ok {
		clean := strings.Join(words[:len(words)-1], " ")
		return clean, &t, true
	}

	// Try last two words (e.g. "next friday", "in 3")
	if len(words) >= 2 {
		twoLast := words[len(words)-2] + " " + words[len(words)-1]
		if t, ok := Parse(twoLast, now); ok {
			clean := strings.Join(words[:len(words)-2], " ")
			return clean, &t, true
		}
	}

	// Try last three words ("in 3 days", "in 2 weeks")
	if len(words) >= 3 {
		threeLast := words[len(words)-3] + " " + words[len(words)-2] + " " + words[len(words)-1]
		if t, ok := Parse(threeLast, now); ok {
			clean := strings.Join(words[:len(words)-3], " ")
			return clean, &t, true
		}
	}

	return text, nil, false
}

// Format returns a human-readable string for a date relative to now.
func Format(t time.Time, now time.Time) string {
	sod := startOfDay(now)
	diff := int(startOfDay(t).Sub(sod).Hours() / 24)

	switch diff {
	case -1:
		return "yesterday"
	case 0:
		return "today"
	case 1:
		return "tomorrow"
	}
	if diff > 1 && diff < 7 {
		return fmt.Sprintf("in %d days", diff)
	}
	return t.Format("02 Jan 2006")
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func nextWeekday(now time.Time, name string) (time.Time, bool) {
	name = strings.TrimFunc(name, unicode.IsSpace)
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
		// Short forms
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}
	target, ok := days[name]
	if !ok {
		return time.Time{}, false
	}
	d := int(target - now.Weekday())
	if d <= 0 {
		d += 7
	}
	return startOfDay(now.AddDate(0, 0, d)), true
}

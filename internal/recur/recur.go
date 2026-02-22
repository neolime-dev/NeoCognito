// Package recur implements the recurrence engine for NeoCognito blocks.
// It parses recurrence rules stored in a block's `recur` frontmatter field
// and computes the next due date when a task is completed.
package recur

import (
	"fmt"
	"strings"
	"time"
)

// Rule describes a parsed recurrence specification.
type Rule struct {
	raw string
}

// Parse parses a recurrence string such as:
//
//	daily, weekly, monthly, yearly
//	every-monday, every-tuesday … every-sunday
//	every-2-days, every-3-weeks, every-6-months
func Parse(s string) (Rule, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return Rule{}, fmt.Errorf("empty recurrence rule")
	}
	// Validate by trying to compute next
	_, err := next(s, time.Now())
	if err != nil {
		return Rule{}, fmt.Errorf("invalid recurrence rule %q: %w", s, err)
	}
	return Rule{raw: s}, nil
}

// NextDue returns the next occurrence after the given time.
func (r Rule) NextDue(from time.Time) (time.Time, error) {
	return next(r.raw, from)
}

// NextDue is a package-level convenience that parses on the fly.
func NextDue(rule string, from time.Time) (*time.Time, error) {
	t, err := next(strings.TrimSpace(strings.ToLower(rule)), from)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func next(rule string, from time.Time) (time.Time, error) {
	sod := startOfDay(from)

	switch rule {
	case "daily":
		return sod.AddDate(0, 0, 1), nil
	case "weekly":
		return sod.AddDate(0, 0, 7), nil
	case "monthly":
		return sod.AddDate(0, 1, 0), nil
	case "yearly":
		return sod.AddDate(1, 0, 0), nil
	}

	// every-<weekday>
	if strings.HasPrefix(rule, "every-") {
		rest := strings.TrimPrefix(rule, "every-")

		// every-N-days / every-N-weeks / every-N-months
		parts := strings.SplitN(rest, "-", 2)
		if len(parts) == 2 {
			var n int
			if _, err := fmt.Sscanf(parts[0], "%d", &n); err == nil {
				unit := parts[1]
				// strip trailing s
				unit = strings.TrimRight(unit, "s")
				switch unit {
				case "day":
					return sod.AddDate(0, 0, n), nil
				case "week":
					return sod.AddDate(0, 0, n*7), nil
				case "month":
					return sod.AddDate(0, n, 0), nil
				}
			}
		}

		// every-weekday
		if t, ok := nextWeekday(from, rest); ok {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognised rule: %q", rule)
}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday,
	"tuesday": time.Tuesday, "wednesday": time.Wednesday,
	"thursday": time.Thursday, "friday": time.Friday,
	"saturday": time.Saturday,
	"sun":      time.Sunday, "mon": time.Monday,
	"tue": time.Tuesday, "wed": time.Wednesday,
	"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
}

func nextWeekday(from time.Time, name string) (time.Time, bool) {
	target, ok := weekdays[name]
	if !ok {
		return time.Time{}, false
	}
	d := int(target - from.Weekday())
	if d <= 0 {
		d += 7
	}
	return startOfDay(from.AddDate(0, 0, d)), true
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

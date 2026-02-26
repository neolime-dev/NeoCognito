// Package pomodoro implements the Pomodoro focus timer TUI component.
package pomodoro

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/block"
)

// Phase represents the current Pomodoro phase.
type Phase int

const (
	PhaseIdle       Phase = iota
	PhaseWork             // 25 min
	PhaseShortBreak       // 5 min
	PhaseLongBreak        // 15 min
)

var (
	workDuration       = 25 * time.Minute
	shortBreakDuration = 5 * time.Minute
	longBreakDuration  = 15 * time.Minute
)

// TickMsg is sent every second when the timer is running.
type TickMsg time.Time

// DoneMsg is sent when a work session completes.
type DoneMsg struct{ Block *block.Block }

// Model is the Pomodoro state.
type Model struct {
	phase     Phase
	remaining time.Duration
	target    *block.Block
	sessions  int // completed work sessions since last long break
	active    bool
}

// New creates an idle Pomodoro model.
func New() Model {
	return Model{phase: PhaseIdle}
}

// Start begins a work session for the given block.
func (m *Model) Start(b *block.Block) {
	m.target = b
	m.phase = PhaseWork
	m.remaining = workDuration
	m.active = true
}

// Pause toggles the timer on/off.
func (m *Model) Pause() { m.active = !m.active }

// Reset returns the timer to idle.
func (m *Model) Reset() {
	m.phase = PhaseIdle
	m.active = false
	m.remaining = 0
	m.target = nil
}

// IsRunning reports whether the timer is counting down.
func (m Model) IsRunning() bool { return m.active && m.phase != PhaseIdle }

// Phase returns the current phase.
func (m Model) CurrentPhase() Phase { return m.phase }

// Target returns the block being focused on.
func (m Model) Target() *block.Block { return m.target }

// Init satisfies tea.Model (unused here — ticking is driven by the root app).
func (m Model) Init() tea.Cmd { return nil }

// Tick advances the timer by one second. Should be called from the root app's
// Update when a TickMsg arrives.
func (m Model) Tick() (Model, tea.Cmd) {
	if !m.active || m.phase == PhaseIdle {
		return m, nil
	}

	m.remaining -= time.Second
	if m.remaining > 0 {
		return m, tickCmd()
	}

	// Phase just ended
	switch m.phase {
	case PhaseWork:
		m.sessions++
		m.active = false
		b := m.target
		if b != nil {
			b.Pomodoros++
			b.TimeSpent += int(workDuration.Minutes())
		}
		// Decide break type
		if m.sessions%4 == 0 {
			m.phase = PhaseLongBreak
			m.remaining = longBreakDuration
		} else {
			m.phase = PhaseShortBreak
			m.remaining = shortBreakDuration
		}
		m.active = true
		return m, tea.Batch(tickCmd(), func() tea.Msg { return DoneMsg{Block: b} })

	case PhaseShortBreak, PhaseLongBreak:
		m.phase = PhaseIdle
		m.active = false
		m.remaining = 0
	}

	return m, nil
}

// StatusBarSegment returns the formatted timer string for the status bar.
// Returns "" when idle.
func (m Model) StatusBarSegment() string {
	if m.phase == PhaseIdle {
		return ""
	}
	icon := phaseIcon(m.phase)
	mins := int(m.remaining.Minutes())
	secs := int(m.remaining.Seconds()) % 60
	label := fmt.Sprintf("%s %02d:%02d", icon, mins, secs)
	if !m.active {
		label += " 󰏤"
	}
	return label
}

// ShortStatus returns a one-line description for the status bar.
func (m Model) ShortStatus() string {
	if m.phase == PhaseIdle || m.target == nil {
		return ""
	}
	title := m.target.Title
	const maxLen = 20
	if len(title) > maxLen {
		title = title[:maxLen] + "…"
	}
	return fmt.Sprintf(" — %s", title)
}

// View returns a panel view (used when the user wants to see the timer detail).
func (m Model) View() string {
	if m.phase == PhaseIdle {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %s  %s\n", phaseLabel(m.phase), m.StatusBarSegment()))
	if m.target != nil {
		sb.WriteString(fmt.Sprintf("  Focus: %s\n", m.target.Title))
		sb.WriteString(fmt.Sprintf("  Sessions: %d  󰔠×%d  󰔛 %dm\n", m.sessions, m.target.Pomodoros, m.target.TimeSpent))
	}
	sb.WriteString("\n  [p] pause  [.] reset")
	return sb.String()
}

// tickCmd returns a command that fires a TickMsg after one second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func phaseIcon(p Phase) string {
	switch p {
	case PhaseWork:
		return "󰔠"
	case PhaseShortBreak:
		return "󱌏"
	case PhaseLongBreak:
		return "󰌯"
	}
	return ""
}

func phaseLabel(p Phase) string {
	switch p {
	case PhaseWork:
		return "FOCUS"
	case PhaseShortBreak:
		return "SHORT BREAK"
	case PhaseLongBreak:
		return "LONG BREAK"
	}
	return ""
}

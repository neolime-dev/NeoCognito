// Package statusbar provides the dynamic keybinding footer.
package statusbar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// Binding represents a single keybinding hint.
type Binding struct {
	Key  string
	Desc string
}

// Model is the statusbar state.
type Model struct {
	Bindings        []Binding
	Width           int
	Mode            string // "NORMAL" or "INSERT"
	PomodoroSegment string // e.g. "🍅 12:43", set by app each frame
}

// New creates a new statusbar.
func New() Model {
	return Model{
		Mode: "NORMAL",
	}
}

// SetWidth updates the status bar width.
func (m *Model) SetWidth(w int) {
	m.Width = w
}

// SetBindings replaces the current keybinding hints.
func (m *Model) SetBindings(bindings []Binding) {
	m.Bindings = bindings
}

// View renders the status bar.
func (m Model) View() string {
	// Left side: keybindings
	var parts []string
	for _, b := range m.Bindings {
		part := styles.KeyStyle.Render(b.Key) + " " + styles.DescStyle.Render(b.Desc)
		parts = append(parts, part)
	}
	left := strings.Join(parts, "  │  ")

	// Calculate available space for left portion
	pomodo := ""
	if m.PomodoroSegment != "" {
		pomodo = "  " + styles.KeyStyle.Render(m.PomodoroSegment) + "  "
	}

	gap := m.Width - lipgloss.Width(left) - lipgloss.Width(pomodo) - 2
	if gap < 0 {
		gap = 0
	}

	bar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" "+left,
		strings.Repeat(" ", gap),
		pomodo,
	)

	return styles.StatusBarStyle.Width(m.Width).Render(bar)
}

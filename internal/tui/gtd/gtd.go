// Package gtd provides an "Inbox Zero" processing wizard overlay.
package gtd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// Model represents the GTD processor state.
type Model struct {
	blocks []*block.Block
	cursor int // points to the current block being processed

	Visible bool
	Width   int
	Height  int

	UpdateFn func(b *block.Block) error
	DeleteFn func(b *block.Block) error
}

// New creates a new Model.
func New() Model {
	return Model{}
}

// Start begins processing the given list of inbox blocks.
func (m *Model) Start(inbox []*block.Block) {
	m.blocks = inbox
	m.cursor = 0
	m.Visible = true
}

// Stop ends processing.
func (m *Model) Stop() {
	m.Visible = false
	m.blocks = nil
}

// Update handles input for processing.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	if len(m.blocks) == 0 || m.cursor >= len(m.blocks) {
		m.Stop()
		return m, nil
	}

	b := m.blocks[m.cursor]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.Stop()
			return m, nil

		case "x": // Delete
			if m.DeleteFn != nil {
				if err := m.DeleteFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
			}
			m.next()

		case "a": // Archive (No status)
			if m.UpdateFn != nil {
				b.Status = ""
				if err := m.UpdateFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
			}
			m.next()

		case "t", "enter": // Taskify
			if m.UpdateFn != nil {
				b.Status = block.StatusTodo
				if err := m.UpdateFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
			}
			m.next()

		case "y": // Completed
			if m.UpdateFn != nil {
				b.Status = block.StatusDone
				if err := m.UpdateFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
			}
			m.next()

		case "s": // skip for now
			m.next()
		}
	}
	return m, nil
}

func (m *Model) next() {
	m.cursor++
	if m.cursor >= len(m.blocks) {
		m.Stop()
	}
}

// View renders the GTD processing overlay.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	if len(m.blocks) == 0 || m.cursor >= len(m.blocks) {
		return ""
	}

	b := m.blocks[m.cursor]

	var sb strings.Builder
	header := styles.TitleStyle.Render("⚡ Inbox Zero Wizard")
	progress := styles.DimItemStyle.Render(fmt.Sprintf(" [%d/%d]", m.cursor+1, len(m.blocks)))
	sb.WriteString(header + progress + "\n\n")

	// Current block
	sb.WriteString("  " + styles.SelectedItemStyle.Render(b.Title) + "\n")
	if b.Area != "" {
		sb.WriteString("  " + styles.TagStyle.Render("@"+b.Area) + "\n")
	}
	sb.WriteString("\n")

	// Question
	sb.WriteString(styles.NormalModeStyle.Bold(true).Render("  Is this actionable?") + "\n\n")

	// Options
	options := []struct{ key, desc string }{
		{"[t/enter]", "Yes — Move to Tasks (Todo)"},
		{"[y]", "Yes — Done right now"},
		{"[a]", "No  — Archive as Reference"},
		{"[x]", "No  — Delete"},
		{"[s]", "Skip for later"},
	}

	for _, opt := range options {
		sb.WriteString(fmt.Sprintf("  %-10s %s\n",
			styles.PrimaryStyle.Render(opt.key),
			styles.NormalItemStyle.Render(opt.desc)))
	}

	sb.WriteString("\n" + styles.DimItemStyle.Render("  [esc/q] Exit Wizard"))

	return styles.ActiveBorder.
		BorderForeground(styles.Accent).
		Width(m.Width-4).
		Height(m.Height-4).
		Padding(1, 2).
		Render(sb.String())
}

// SetSize updates dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

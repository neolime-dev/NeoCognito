// Package related provides the Related Blocks side panel using TF-IDF / FTS matching.
package related

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// OpenRelatedMsg signals to open the related blocks overlay for a specific block.
type OpenRelatedMsg struct {
	Target *block.Block
}

// CloseRelatedMsg signals to close the overlay.
type CloseRelatedMsg struct{}

// SelectResultMsg signals that the user selected one of the related blocks.
type SelectResultMsg struct {
	Block *block.Block
}

// Model represents the related blocks overlay.
type Model struct {
	Target  *block.Block
	blocks  []*block.Block
	cursor  int
	Visible bool
	Width   int
	Height  int

	// Engine connection
	FindRelatedFn func(target *block.Block, limit int) ([]*block.Block, error)
}

// New creates a new related blocks model.
func New() Model {
	return Model{}
}

// Open initializes the overlay for a specific block.
func (m *Model) Open(target *block.Block) {
	m.Visible = true
	m.Target = target
	m.cursor = 0
	if m.FindRelatedFn != nil && target != nil {
		m.blocks, _ = m.FindRelatedFn(target, 10)
	}
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages when the overlay is active.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			m.Visible = false
			return m, func() tea.Msg { return CloseRelatedMsg{} }
		case "j", "down":
			if m.cursor < len(m.blocks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if len(m.blocks) > 0 && m.cursor < len(m.blocks) {
				m.Visible = false
				selected := m.blocks[m.cursor]
				return m, func() tea.Msg { return SelectResultMsg{Block: selected} }
			}
		}
	}

	return m, nil
}

// View renders the floating panel.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	var sb strings.Builder

	// Header
	sb.WriteString(styles.TitleStyle.Render("󰈎 Related Blocks"))
	if m.Target != nil {
		sb.WriteString(styles.DimItemStyle.Render(" for " + m.Target.Title))
	}
	sb.WriteString("\n\n")

	if len(m.blocks) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("  No related blocks found."))
	} else {
		for i, b := range m.blocks {
			cursor := "  "
			style := styles.NormalItemStyle
			if i == m.cursor {
				cursor = "▸ "
				style = styles.SelectedItemStyle
			}

			line := cursor + style.Render(b.Title)
			if len(b.Tags) > 0 {
				line += " " + styles.TagStyle.Render(strings.Join(b.Tags, " "))
			}
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString("\n" + styles.DimItemStyle.Render("  [esc] close  [enter] open"))

	content := lipgloss.NewStyle().
		Margin(1, 2).
		Render(sb.String())

	return styles.ActiveBorder.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}

// SetSize updates the overlay dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

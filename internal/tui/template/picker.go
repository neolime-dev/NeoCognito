// Package template provides the TUI picker for instantiation.
package template

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/template"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
	"github.com/sahilm/fuzzy"
)

// SelectedMsg is sent when a template is chosen to be instantiated.
type SelectedMsg struct {
	Template template.Template
}

// CloseMsg is sent when the picker is closed.
type CloseMsg struct{}

// Model is the template picker state.
type Model struct {
	Visible   bool
	templates []template.Template
	filtered  []template.Template
	cursor    int
	input     textinput.Model
	Width     int
	Height    int
}

// New creates a new template picker model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Search templates..."
	ti.Prompt = "  🔍 "
	ti.PromptStyle = styles.SearchPromptStyle
	ti.TextStyle = styles.NormalItemStyle
	ti.CharLimit = 64

	return Model{
		input: ti,
	}
}

// Open shows the picker and populates it with templates.
func (m *Model) Open(templates []template.Template) tea.Cmd {
	m.Visible = true
	m.templates = templates
	m.filtered = templates
	m.cursor = 0
	m.input.SetValue("")
	m.input.Focus()
	return textinput.Blink
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.Visible = false
			m.input.Blur()
			return m, func() tea.Msg { return CloseMsg{} }

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.Visible = false
				m.input.Blur()
				selected := m.filtered[m.cursor]
				return m, func() tea.Msg { return SelectedMsg{Template: selected} }
			}

		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.filtered) - 1
			}

		case "down", "ctrl+j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		default:
			// Let textinput handle normal typing
			prevValue := m.input.Value()
			m.input, cmd = m.input.Update(msg)
			if m.input.Value() != prevValue {
				m.filter()
			}
			return m, cmd
		}
	}

	return m, cmd
}

func (m *Model) filter() {
	query := strings.TrimSpace(m.input.Value())
	if query == "" {
		m.filtered = m.templates
		m.cursor = 0
		return
	}

	var names []string
	for _, t := range m.templates {
		names = append(names, t.Name)
	}

	matches := fuzzy.Find(query, names)
	m.filtered = nil
	for _, match := range matches {
		m.filtered = append(m.filtered, m.templates[match.Index])
	}
	m.cursor = 0
}

// View renders the template picker overlay.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	var sb strings.Builder

	header := styles.TitleStyle.Render("✨ New from Template")
	sb.WriteString(header + "\n\n")
	sb.WriteString(m.input.View() + "\n\n")

	maxVisible := m.Height - 6
	if maxVisible < 1 {
		maxVisible = 5
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	if len(m.filtered) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("  No matching templates."))
	}

	for i := start; i < end; i++ {
		t := m.filtered[i]
		cursor := "  "
		style := styles.NormalItemStyle

		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedItemStyle
		}

		line := cursor + style.Render(t.Name)
		if len(t.Block.Tags) > 0 {
			line += " " + styles.TagStyle.Render(strings.Join(t.Block.Tags, " "))
		}

		sb.WriteString(line + "\n")
	}

	if len(m.filtered) > maxVisible {
		info := fmt.Sprintf("\n  %d/%d", m.cursor+1, len(m.filtered))
		sb.WriteString(styles.DimItemStyle.Render(info))
	}

	return styles.ActiveBorder.
		Width(m.Width).
		Height(m.Height).
		Render(sb.String())
}

// SetSize updates panel dimensions
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.input.Width = w - 8
}

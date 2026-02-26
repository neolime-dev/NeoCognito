// Package search provides the Omni-Search overlay panel.
package search

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
	"github.com/sahilm/fuzzy"
)

// OpenSearchMsg signals the app to open the search overlay.
type OpenSearchMsg struct{}

// CloseSearchMsg signals the app to close the search overlay.
type CloseSearchMsg struct{}

// SelectResultMsg indicates a search result was selected.
type SelectResultMsg struct {
	Block *block.Block
}

// Model is the search overlay state.
type Model struct {
	input    textinput.Model
	allItems []*block.Block
	results  []*block.Block
	cursor   int
	Active   bool
	Width    int
	Height   int

	// Callback for DB search — set by parent.
	SearchFn func(query string) ([]*block.Block, error)
}

// New creates a new search model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Search blocks..."
	ti.CharLimit = 128
	ti.Prompt = "  🔍 "
	ti.PromptStyle = styles.SearchPromptStyle
	ti.TextStyle = styles.NormalItemStyle

	return Model{
		input: ti,
	}
}

// SetItems provides the full set of blocks for fuzzy matching.
func (m *Model) SetItems(blocks []*block.Block) {
	m.allItems = blocks
}

// Init shows the search view and focuses the input.
func (m *Model) Init() tea.Cmd {
	m.cursor = 0
	if m.results == nil {
		m.results = m.allItems
	}
	m.input.SetValue("")
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var cmds []tea.Cmd

	if m.Active && !m.input.Focused() {
		m.input.Focus()
		cmds = append(cmds, textinput.Blink)
	} else if !m.Active && m.input.Focused() {
		m.input.Blur()
	}

	if !m.Active {
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, nil
		case "enter":
			if len(m.results) > 0 && m.cursor < len(m.results) {
				selected := m.results[m.cursor]
				m.input.Blur()
				return m, func() tea.Msg { return SelectResultMsg{Block: selected} }
			}
		case "down", "ctrl+n":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
			return m, nil
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.results) > 0 {
				m.cursor = len(m.results) - 1
			}
			return m, nil
		}
	}

	// Update text input
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	// Perform fuzzy search on the input value
	query := strings.TrimSpace(m.input.Value())
	if query == "" {
		m.results = m.allItems
		m.cursor = 0
	} else {
		m.performSearch(query)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) performSearch(query string) {
	// First try DB search if available
	if m.SearchFn != nil {
		results, err := m.SearchFn(query)
		if err == nil && len(results) > 0 {
			m.results = results
			if m.cursor >= len(m.results) {
				m.cursor = 0
			}
			return
		}
	}

	// Fallback: fuzzy match on titles from allItems
	if len(m.allItems) == 0 {
		m.results = nil
		return
	}

	titles := make([]string, len(m.allItems))
	for i, b := range m.allItems {
		titles[i] = b.Title + " " + strings.Join(b.Tags, " ")
	}

	matches := fuzzy.Find(query, titles)
	m.results = nil
	for _, match := range matches {
		if match.Index < len(m.allItems) {
			m.results = append(m.results, m.allItems[match.Index])
		}
	}
	if m.cursor >= len(m.results) {
		m.cursor = 0
	}
}

// View renders the search panel.
func (m Model) View() string {

	var sb strings.Builder

	header := styles.TitleStyle.Render("󰍉 Omni-Search")
	sb.WriteString(header + "\n\n")
	sb.WriteString(m.input.View() + "\n\n")

	if len(m.results) == 0 && m.input.Value() != "" {
		sb.WriteString(styles.DimItemStyle.Render("  No results found.") + "\n")
	}

	maxResults := m.Height - 8
	if maxResults < 5 {
		maxResults = 5
	}
	end := len(m.results)
	if end > maxResults {
		end = maxResults
	}

	for i := 0; i < end; i++ {
		b := m.results[i]
		cursor := "  "
		style := styles.NormalItemStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedItemStyle
		}

		line := cursor + style.Render(b.Title)
		if b.Status != "" {
			line += " " + styles.StatusBadge(b.Status)
		}
		if len(b.Tags) > 0 {
			line += " " + styles.TagStyle.Render(strings.Join(b.Tags, " "))
		}
		sb.WriteString(line + "\n")

		// Display FTS snippet if present (loaded into BodyPreviewText from DB)
		if strings.Contains(b.BodyPreviewText, "»") {
			// clean up line breaks
			snippet := strings.ReplaceAll(b.BodyPreviewText, "\n", " ")
			sb.WriteString(styles.DimItemStyle.Render("      ..."+snippet) + "\n")
		}
	}

	if len(m.results) > maxResults {
		sb.WriteString(styles.DimItemStyle.Render(
			strings.Repeat(" ", 4)+"... and more",
		) + "\n")
	}

	content := sb.String()

	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}

	return borderStyle.Width(m.Width).Height(m.Height).Render(content)
}

// SetSize updates the overlay dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.input.Width = w - 10
}

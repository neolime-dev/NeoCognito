// Package waiting provides the "⏳ Waiting" TUI panel for delegated blocks.
package waiting

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// SelectedMsg is sent when the user selects a waiting block to view.
type SelectedMsg struct{ Block *block.Block }

// Model is the Waiting panel state.
type Model struct {
	blocks []waitingEntry
	cursor int
	Active bool
	Width  int
	Height int
}

type waitingEntry struct {
	block     *block.Block
	daysSince int
}

// New creates an empty waiting model.
func New() Model { return Model{} }

// SetBlocks filters and loads delegated blocks.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.blocks = nil
	for _, b := range blocks {
		if b.Status == block.StatusDone || b.Status == block.StatusArchived {
			continue
		}

		isWaiting := b.DelegatedTo != ""
		for _, t := range b.Tags {
			if strings.ToLower(t) == "#waiting" || strings.ToLower(t) == "waiting" || strings.ToLower(t) == "@waiting" {
				isWaiting = true
				break
			}
		}

		if isWaiting {
			days := int(time.Since(b.Modified).Hours() / 24)
			m.blocks = append(m.blocks, waitingEntry{block: b, daysSince: days})
		}
	}
	if m.cursor >= len(m.blocks) {
		m.cursor = max(0, len(m.blocks)-1)
	}
}

// SelectedBlock returns the currently selected block, or nil.
func (m Model) SelectedBlock() *block.Block {
	if len(m.blocks) == 0 || m.cursor >= len(m.blocks) {
		return nil
	}
	return m.blocks[m.cursor].block
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.blocks)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.blocks) > 0 {
				m.cursor = len(m.blocks) - 1
			}
		case "g":
			m.cursor = 0
		case "G":
			if len(m.blocks) > 0 {
				m.cursor = len(m.blocks) - 1
			}
		case "enter":
			if b := m.SelectedBlock(); b != nil {
				return m, func() tea.Msg { return SelectedMsg{Block: b} }
			}
		}
	}
	return m, nil
}

// View renders the waiting panel.
func (m Model) View() string {
	border := styles.InactiveBorder
	if m.Active {
		border = styles.ActiveBorder
	}

	var sb strings.Builder
	title := fmt.Sprintf("󰔛 Waiting (%d)", len(m.blocks))
	sb.WriteString(styles.TitleStyle.Render(title) + "\n")
	sb.WriteString(styles.DimItemStyle.Render("  Shows tasks delegated to others or tagged with #waiting.") + "\n\n")

	if len(m.blocks) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("  No delegated items — nice!"))
	}

	for i, e := range m.blocks {
		cursor := "  "
		itemStyle := styles.NormalItemStyle
		if i == m.cursor {
			cursor = "▸ "
			itemStyle = styles.SelectedItemStyle
		}

		// Urgency badge based on days waiting
		urgency := urgencyBadge(e.daysSince)
		delegate := styles.DimItemStyle.Render("→ " + e.block.DelegatedTo)
		waiting := styles.DimItemStyle.Render(fmt.Sprintf("(%dd)", e.daysSince))

		line := fmt.Sprintf("%s%s %s %s %s", cursor, urgency, itemStyle.Render(e.block.Title), delegate, waiting)
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n" + styles.DimItemStyle.Render("  [j/k] navigate  [Enter] view  [e] edit"))

	return border.Width(m.Width).Height(m.Height).Render(sb.String())
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) { m.Width = w; m.Height = h }

func urgencyBadge(days int) string {
	switch {
	case days >= 7:
		return "🔴"
	case days >= 3:
		return "🟡"
	default:
		return "🟢"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

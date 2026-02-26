// Package journal provides a chronological journal view grouped by day.
package journal

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// dayGroup groups blocks by calendar date.
type dayGroup struct {
	date   time.Time
	blocks []*block.Block
}

// OpenEditorMsg is sent when the user wants to open a block in the editor.
type OpenEditorMsg struct {
	FilePath string
}

// GoBackMsg signals navigation back to the previous screen.
type GoBackMsg struct{}

// Model is the Journal View state.
type Model struct {
	days   []dayGroup
	dayCur int // which day group is selected
	blkCur int // which block within that day

	Width  int
	Height int
	Active bool
}

// New creates a new Journal model.
func New() Model {
	return Model{}
}

// SetBlocks updates the journal with a fresh set of blocks.
func (m *Model) SetBlocks(blocks []*block.Block) {
	// Group blocks by calendar date (using Created)
	grouped := map[string]*dayGroup{}
	for _, b := range blocks {
		key := b.Created.Format("2006-01-02")
		if _, ok := grouped[key]; !ok {
			grouped[key] = &dayGroup{date: b.Created.Truncate(24 * time.Hour)}
		}
		grouped[key].blocks = append(grouped[key].blocks, b)
	}

	// Sort days newest first
	m.days = make([]dayGroup, 0, len(grouped))
	for _, g := range grouped {
		m.days = append(m.days, *g)
	}
	sort.Slice(m.days, func(i, j int) bool {
		return m.days[i].date.After(m.days[j].date)
	})

	// Clamp cursors
	if m.dayCur >= len(m.days) {
		m.dayCur = 0
	}
	m.blkCur = 0
}

// SelectedBlock returns the currently highlighted block.
func (m Model) SelectedBlock() *block.Block {
	if len(m.days) == 0 || m.dayCur >= len(m.days) {
		return nil
	}
	day := m.days[m.dayCur]
	if len(day.blocks) == 0 || m.blkCur >= len(day.blocks) {
		return nil
	}
	return day.blocks[m.blkCur]
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles input.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if len(m.days) == 0 {
				break
			}
			day := m.days[m.dayCur]
			if m.blkCur < len(day.blocks)-1 {
				m.blkCur++
			} else if m.dayCur < len(m.days)-1 {
				m.dayCur++
				m.blkCur = 0
			} else {
				// Wrap to absolute start
				m.dayCur = 0
				m.blkCur = 0
			}
		case "k", "up":
			if m.blkCur > 0 {
				m.blkCur--
			} else if m.dayCur > 0 {
				m.dayCur--
				m.blkCur = len(m.days[m.dayCur].blocks) - 1
				if m.blkCur < 0 {
					m.blkCur = 0
				}
			} else {
				// Wrap to absolute end
				if len(m.days) > 0 {
					m.dayCur = len(m.days) - 1
					m.blkCur = len(m.days[m.dayCur].blocks) - 1
					if m.blkCur < 0 {
						m.blkCur = 0
					}
				}
			}
		case "g":
			m.dayCur = 0
			m.blkCur = 0
		case "G":
			if len(m.days) > 0 {
				m.dayCur = len(m.days) - 1
				day := m.days[m.dayCur]
				m.blkCur = len(day.blocks) - 1
				if m.blkCur < 0 {
					m.blkCur = 0
				}
			}
		case "enter":
			if b := m.SelectedBlock(); b != nil && b.FilePath != "" {
				return m, func() tea.Msg { return OpenEditorMsg{FilePath: b.FilePath} }
			}
		case "esc":
			return m, func() tea.Msg { return GoBackMsg{} }
		}
	}
	return m, nil
}

// View renders the journal panel.
func (m Model) View() string {
	header := styles.TitleStyle.Render("󰏫 Journal")
	subtitle := styles.DimItemStyle.Render("  All blocks, newest first")

	if len(m.days) == 0 {
		empty := styles.DimItemStyle.Render("  No entries yet.")
		content := lipgloss.JoinVertical(lipgloss.Left, header+"\n", subtitle+"\n", "\n"+empty)
		return m.border().Render(content)
	}

	var sb strings.Builder
	sb.WriteString(header + "\n")
	sb.WriteString(subtitle + "\n")

	linesWritten := lipgloss.Height(header) + 1
	avail := m.Height - 4
	if avail < 1 {
		avail = 10
	}

	for di, day := range m.days {
		if linesWritten >= avail {
			break
		}
		// Day section header
		dayStr := day.date.Format("2006-01-02 (Monday)")
		dayHeader := styles.PrimaryStyle.Bold(true).Render("── " + dayStr + " ──")
		sb.WriteString("\n  " + dayHeader + "\n")
		linesWritten += 2

		for bi, b := range day.blocks {
			if linesWritten >= avail {
				sb.WriteString("  " + styles.DimItemStyle.Render("  …") + "\n")
				break
			}
			cursor := "  "
			style := styles.NormalItemStyle
			if di == m.dayCur && bi == m.blkCur && m.Active {
				cursor = "▸ "
				style = styles.SelectedItemStyle
			}

			title := b.Title
			if title == "" {
				title = "(untitled)"
			}
			maxLen := m.Width - 10
			if maxLen > 0 && len(title) > maxLen {
				title = title[:maxLen-1] + "…"
			}

			timeStr := styles.DimItemStyle.Render(b.Created.Format("15:04"))
			line := fmt.Sprintf("  %s%s  %s", cursor, style.Render(title), timeStr)
			sb.WriteString(line + "\n")
			linesWritten++
		}
	}

	sb.WriteString("\n  " + styles.DimItemStyle.Render("[Enter] edit  [j/k] navigate  [Esc] go back"))

	return m.border().Render(sb.String())
}

func (m Model) border() lipgloss.Style {
	if m.Active {
		return styles.ActiveBorder.Width(m.Width).Height(m.Height)
	}
	return styles.InactiveBorder.Width(m.Width).Height(m.Height)
}

// SetSize updates dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

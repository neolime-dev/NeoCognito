// Package kanban provides a 3-column TUI board for managing tasks.
package kanban

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// StatusCycledMsg is sent when a block moves columns (updates status).
type StatusCycledMsg struct {
	Block *block.Block
}

const (
	colTodo  = 0
	colDoing = 1
	colDone  = 2
)

// Model is the Kanban board state.
type Model struct {
	todo  []*block.Block
	doing []*block.Block
	done  []*block.Block

	colFocus int    // 0, 1, or 2
	cursors  [3]int // the cursor position within each column

	Width  int
	Height int
	Active bool

	visualMode  bool
	visualStart int

	// Callback for updating blocks — set by parent.
	UpdateFn func(b *block.Block) error
	DeleteFn func(b *block.Block) error
}

// New creates a new Kanban model.
func New() Model {
	return Model{}
}

// SetBlocks distributes blocks into the 3 columns based on status.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.todo = nil
	m.doing = nil
	m.done = nil

	for _, b := range blocks {
		if !b.IsTask() {
			continue
		}
		switch b.Status {
		case block.StatusTodo:
			m.todo = append(m.todo, b)
		case block.StatusDoing:
			m.doing = append(m.doing, b)
		case block.StatusDone:
			m.done = append(m.done, b)
		}
	}

	// Clamp cursors
	m.clampCursor(colTodo, len(m.todo))
	m.clampCursor(colDoing, len(m.doing))
	m.clampCursor(colDone, len(m.done))
}

func (m *Model) clampCursor(col, length int) {
	if length == 0 {
		m.cursors[col] = 0
	} else if m.cursors[col] >= length {
		m.cursors[col] = length - 1
	}
}

// SelectedBlocks returns all currently selected blocks from the focused column.
func (m Model) SelectedBlocks() []*block.Block {
	col := m.currentColumn()
	if len(col) == 0 {
		return nil
	}
	if !m.visualMode {
		return []*block.Block{col[m.cursors[m.colFocus]]}
	}

	start, end := m.visualStart, m.cursors[m.colFocus]
	if start > end {
		start, end = end, start
	}
	end++ // inclusive slice bound
	if end > len(col) {
		end = len(col)
	}
	return col[start:end]
}

// SelectedBlock returns the currently hovered block, or nil if none.
func (m Model) SelectedBlock() *block.Block {
	col := m.currentColumn()
	if len(col) == 0 {
		return nil
	}
	return col[m.cursors[m.colFocus]]
}

func (m Model) currentColumn() []*block.Block {
	switch m.colFocus {
	case colTodo:
		return m.todo
	case colDoing:
		return m.doing
	case colDone:
		return m.done
	}
	return nil
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if m.colFocus > 0 {
				m.colFocus--
				m.visualMode = false
			}
		case "l", "right":
			if m.colFocus < 2 {
				m.colFocus++
				m.visualMode = false
			}
		case "j", "down":
			col := m.currentColumn()
			if m.cursors[m.colFocus] < len(col)-1 {
				m.cursors[m.colFocus]++
			} else {
				m.cursors[m.colFocus] = 0
			}
		case "k", "up":
			if m.cursors[m.colFocus] > 0 {
				m.cursors[m.colFocus]--
			} else {
				col := m.currentColumn()
				if len(col) > 0 {
					m.cursors[m.colFocus] = len(col) - 1
				}
			}
		case "g":
			m.cursors[m.colFocus] = 0
		case "G":
			col := m.currentColumn()
			if len(col) > 0 {
				m.cursors[m.colFocus] = len(col) - 1
			}
		case "v":
			m.visualMode = !m.visualMode
			if m.visualMode {
				m.visualStart = m.cursors[m.colFocus]
			}
		case "esc":
			if m.visualMode {
				m.visualMode = false
			}
			return m, nil
		case " ": // Space to cycle status
			blocks := m.SelectedBlocks()
			for _, b := range blocks {
				b.NextStatus()
				if m.UpdateFn != nil {
					if err := m.UpdateFn(b); err != nil {
						return m, func() tea.Msg { return err }
					}
				}
			}
			if len(blocks) > 0 {
				return m, func() tea.Msg { return StatusCycledMsg{Block: blocks[0]} }
			}
		case "x": // Delete
			if b := m.SelectedBlock(); b != nil && m.DeleteFn != nil {
				if err := m.DeleteFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
				return m, func() tea.Msg { return StatusCycledMsg{Block: b} }
			}
		}
	}
	return m, nil
}

// View renders the Kanban board.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	borderColor := styles.Muted
	if m.Active {
		borderColor = styles.Primary
	}
	outerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.Width).
		Height(m.Height)

	// Inner content area: width = m.Width (outer border adds 2 on each side,
	// but lipgloss Width() here sets the INNER width).
	innerW := m.Width
	headerStr := styles.TitleStyle.Render("󰟀 Kanban")
	headerH := lipgloss.Height(headerStr)

	availH := m.Height - headerH - 2 // 2 = top+bottom interior padding
	if availH < 2 {
		availH = 2
	}

	// Each column box uses RoundedBorder (2 chars on each side horizontally).
	// 3 columns share innerW minus 2 gap chars → each column outer = (innerW - 2) / 3
	// We give each column an inner width = outerCol - 2 (for its own border).
	outerColW := (innerW - 2) / 3
	if outerColW < 8 {
		outerColW = 8
	}
	colW := outerColW - 2 // inner content width per column
	if colW < 4 {
		colW = 4
	}

	todoStr := m.renderColumn("📋 Todo", m.todo, colTodo, colW, availH)
	doingStr := m.renderColumn("⚡ Doing", m.doing, colDoing, colW, availH)
	doneStr := m.renderColumn("✅ Done", m.done, colDone, colW, availH)

	gap := lipgloss.NewStyle().Width(1).Render("")
	board := lipgloss.JoinHorizontal(lipgloss.Top, todoStr, gap, doingStr, gap, doneStr)
	content := lipgloss.JoinVertical(lipgloss.Left, headerStr, board)

	return outerStyle.Render(content)
}

func (m Model) renderColumn(title string, blocks []*block.Block, colIdx, width, height int) string {
	isActiveCol := m.Active && m.colFocus == colIdx

	var sb strings.Builder

	// Column Header
	headerStyle := styles.DimItemStyle
	if isActiveCol {
		headerStyle = styles.PrimaryStyle.Bold(true)
	}
	sb.WriteString("  " + headerStyle.Render(fmt.Sprintf("%s (%d)", title, len(blocks))) + "\n\n")

	// Cards
	if len(blocks) == 0 {
		sb.WriteString("  " + styles.DimItemStyle.Render("Empty"))
	}

	for i, b := range blocks {
		if i >= height-4 { // simple truncation rather than full scrolling for now
			sb.WriteString("  " + styles.DimItemStyle.Render("..."))
			break
		}

		cursor := "  "
		style := styles.NormalItemStyle

		isSelected := false
		if m.visualMode && isActiveCol {
			start, end := m.visualStart, m.cursors[colIdx]
			if start > end {
				start, end = end, start
			}
			if i >= start && i <= end {
				isSelected = true
			}
		}

		if isActiveCol && i == m.cursors[colIdx] {
			cursor = "▸ "
			if !m.visualMode {
				style = styles.SelectedItemStyle
			} else {
				style = styles.PrimaryStyle.Bold(true)
			}
		} else if isSelected {
			style = styles.SelectedItemStyle
		}

		badge := styles.StatusBadge(b.Status)
		titleStr := style.Render(b.Title)

		// Truncate title string if too long to fit column width
		visibleTitle := titleStr
		// rudimentary truncation for plain string representation of lipgloss,
		// assuming title doesn't contain escape sequences before Render
		rawTitle := b.Title
		maxLen := width - 8 // minus cursor(2), space(1), badge(2), space(1), padding(2)
		if maxLen > 0 && len(rawTitle) > maxLen {
			rawTitle = rawTitle[:maxLen-1] + "…"
			visibleTitle = style.Render(rawTitle)
		}

		line := fmt.Sprintf("  %s%s %s", cursor, badge, visibleTitle)
		sb.WriteString(line + "\n")
	}

	// Outline box for column
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Width(width).
		Height(height)

	if isActiveCol {
		boxStyle = boxStyle.BorderForeground(styles.Primary)
	}

	return boxStyle.Render(sb.String())
}

// SetSize updates panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

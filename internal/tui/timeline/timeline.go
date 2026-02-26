// Package timeline provides the task timeline panel.
package timeline

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// StatusCycledMsg is sent when a block's status is changed.
type StatusCycledMsg struct {
	Block *block.Block
}

// Model is the timeline panel state.
type Model struct {
	rawBlocks   []*block.Block // full unfiltered list from parent
	overdue     []*block.Block
	today       []*block.Block
	upcoming    []*block.Block
	noDue       []*block.Block
	allBlocks   []*block.Block // flat ordered list for navigation
	cursor      int
	visualMode  bool
	visualStart int
	filtering   bool
	filterQuery string
	Width       int
	Height      int
	Active      bool

	// Callbacks — set by parent.
	UpdateFn func(b *block.Block) error
	DeleteFn func(b *block.Block) error
}

// New creates a new timeline model.
func New() Model {
	return Model{}
}

// SetBlocks categorizes blocks into timeline groups.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.rawBlocks = blocks
	m.applyFilter()
}

// applyFilter narrows rawBlocks using m.filterQuery and then categorizes.
func (m *Model) applyFilter() {
	var filtered []*block.Block
	if m.filterQuery == "" {
		filtered = m.rawBlocks
	} else {
		q := strings.ToLower(m.filterQuery)
		for _, b := range m.rawBlocks {
			if strings.Contains(strings.ToLower(b.Title), q) {
				filtered = append(filtered, b)
				continue
			}
			for _, t := range b.Tags {
				if strings.Contains(strings.ToLower(t), q) {
					filtered = append(filtered, b)
					break
				}
			}
		}
	}

	m.overdue = nil
	m.today = nil
	m.upcoming = nil
	m.noDue = nil
	m.allBlocks = nil

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	for _, b := range filtered {
		if b.Due == nil {
			m.noDue = append(m.noDue, b)
		} else if b.Due.Before(todayStart) {
			m.overdue = append(m.overdue, b)
		} else if b.Due.Before(todayEnd) {
			m.today = append(m.today, b)
		} else {
			m.upcoming = append(m.upcoming, b)
		}
	}

	m.allBlocks = append(m.allBlocks, m.overdue...)
	m.allBlocks = append(m.allBlocks, m.today...)
	m.allBlocks = append(m.allBlocks, m.upcoming...)
	m.allBlocks = append(m.allBlocks, m.noDue...)

	if m.cursor >= len(m.allBlocks) {
		m.cursor = 0
	}
}

// SelectedBlocks returns all currently selected blocks.
func (m Model) SelectedBlocks() []*block.Block {
	if len(m.allBlocks) == 0 {
		return nil
	}
	if !m.visualMode {
		if m.cursor >= len(m.allBlocks) {
			return nil
		}
		return []*block.Block{m.allBlocks[m.cursor]}
	}

	start, end := m.visualStart, m.cursor
	if start > end {
		start, end = end, start
	}
	end++ // inclusive slice bound
	if end > len(m.allBlocks) {
		end = len(m.allBlocks)
	}
	return m.allBlocks[start:end]
}

// SelectedBlock returns the currently hovered block, or nil.
func (m Model) SelectedBlock() *block.Block {
	if len(m.allBlocks) == 0 || m.cursor >= len(m.allBlocks) {
		return nil
	}
	return m.allBlocks[m.cursor]
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
		case "j", "down":
			if m.cursor < len(m.allBlocks)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.allBlocks) > 0 {
				m.cursor = len(m.allBlocks) - 1
			}
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
		case "G":
			if len(m.allBlocks) > 0 {
				m.cursor = len(m.allBlocks) - 1
			}
		case "g":
			m.cursor = 0
		case "v":
			m.visualMode = !m.visualMode
			if m.visualMode {
				m.visualStart = m.cursor
			}
		case "f":
			m.filtering = true
			m.filterQuery = ""
			m.applyFilter()
			return m, nil
		case "esc":
			if m.visualMode {
				m.visualMode = false
			}
			if m.filtering {
				m.filtering = false
				m.filterQuery = ""
				m.applyFilter()
			}
			return m, nil
		case "x":
			if b := m.SelectedBlock(); b != nil && m.DeleteFn != nil {
				if err := m.DeleteFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
				return m, func() tea.Msg { return StatusCycledMsg{Block: b} }
			}
		}
	}

	// Filter mode: capture single characters
	if m.filtering {
		if kMsg, ok := msg.(tea.KeyMsg); ok {
			switch kMsg.String() {
			case "backspace":
				if len(m.filterQuery) > 0 {
					m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
					m.applyFilter()
				}
			case "enter", "esc":
				m.filtering = false
			default:
				if r := kMsg.String(); len(r) == 1 {
					m.filterQuery += r
					m.applyFilter()
				}
			}
		}
	}

	return m, nil
}

// View renders the timeline panel.
func (m Model) View() string {
	var sb strings.Builder

	header := styles.TitleStyle.Render("󰄱 Tasks")
	sb.WriteString(header + "\n")

	// Filter indicator
	if m.filterQuery != "" || m.filtering {
		cursor := ""
		if m.filtering {
			cursor = "█"
		}
		indicator := styles.TagStyle.Render(fmt.Sprintf(" filter: %s%s ", m.filterQuery, cursor))
		sb.WriteString(indicator + "\n")
	}

	if len(m.allBlocks) == 0 {
		sb.WriteString("\n" + styles.DimItemStyle.Render("  No tasks found.") + "\n")
	}

	globalIdx := 0
	globalIdx = m.renderGroup(&sb, "⚠ Overdue", m.overdue, globalIdx)
	globalIdx = m.renderGroup(&sb, "📌 Today", m.today, globalIdx)
	globalIdx = m.renderGroup(&sb, "📅 Upcoming", m.upcoming, globalIdx)
	_ = m.renderGroup(&sb, "📝 No Date", m.noDue, globalIdx)

	content := sb.String()
	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}

	return borderStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}

func (m Model) renderGroup(sb *strings.Builder, title string, blocks []*block.Block, startIdx int) int {
	if len(blocks) == 0 {
		return startIdx
	}

	sb.WriteString("\n  " + styles.DimItemStyle.Render(title) + "\n")

	for i, b := range blocks {
		idx := startIdx + i
		cursor := "  "
		style := styles.NormalItemStyle

		isSelected := false
		if m.visualMode {
			start, end := m.visualStart, m.cursor
			if start > end {
				start, end = end, start
			}
			if idx >= start && idx <= end {
				isSelected = true
			}
		}

		if idx == m.cursor && m.Active {
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

		line := fmt.Sprintf("  %s%s  %s", cursor, badge, titleStr)

		if b.Due != nil {
			dueStr := b.Due.Format("02 Jan")
			line += " " + styles.DimItemStyle.Render("("+dueStr+")")
		}

		if len(b.Tags) > 0 {
			line += " " + styles.TagStyle.Render(strings.Join(b.Tags, " "))
		}

		sb.WriteString(line + "\n")
	}

	return startIdx + len(blocks)
}

// SetSize updates panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

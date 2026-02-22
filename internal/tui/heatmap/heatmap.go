// Package heatmap provides a GitHub-style contribution graph for NeoCognito.
package heatmap

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// ViewDayMsg is sent when the user presses Enter on a specific date.
type ViewDayMsg struct {
	Date time.Time
}

// Model is the heatmap calendar state.
type Model struct {
	blocks []*block.Block
	counts map[string]int // "YYYY-MM-DD" -> count
	maxVol int

	cursorDate time.Time
	endDate    time.Time // usually today
	weeks      int       // how many weeks to show (depends on width)

	Width  int
	Height int
	Active bool
}

// New creates a new Heatmap model.
func New() Model {
	now := time.Now()
	// Strip time to just the day
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return Model{
		counts:     make(map[string]int),
		cursorDate: today,
		endDate:    today,
		weeks:      26, // default half-year
	}
}

// SetBlocks populates the heatmap data.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.blocks = blocks
	m.counts = make(map[string]int)
	m.maxVol = 0

	for _, b := range blocks {
		dateStr := b.Created.Format("2006-01-02")
		m.counts[dateStr]++
		if m.counts[dateStr] > m.maxVol {
			m.maxVol = m.counts[dateStr]
		}
	}
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input for navigating the heatmap.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			m.cursorDate = m.cursorDate.AddDate(0, 0, -7)
		case "l", "right":
			m.cursorDate = m.cursorDate.AddDate(0, 0, 7)
		case "k", "up":
			m.cursorDate = m.cursorDate.AddDate(0, 0, -1)
		case "j", "down":
			m.cursorDate = m.cursorDate.AddDate(0, 0, 1)
		case "enter":
			return m, func() tea.Msg { return ViewDayMsg{Date: m.cursorDate} }
		}

		// Clamp cursor to the visible range or reasonable bounds
		minDate := m.endDate.AddDate(0, 0, -(m.weeks*7)+1)
		if m.cursorDate.Before(minDate) {
			m.cursorDate = minDate
		}
		if m.cursorDate.After(m.endDate) {
			m.cursorDate = m.endDate
		}
	}

	return m, nil
}

// View renders the heatmap.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	title := styles.TitleStyle.Render("󰸗 Activity Heatmap")

	// Recalculate weeks based on available width (each week is 2 chars wide + 1 space = 3)
	// We leave some room for margins.
	availCols := (m.Width - 10) / 2
	if availCols > 52 {
		availCols = 52
	}
	if availCols < 4 {
		availCols = 4
	}
	// We don't override m.weeks on the Model itself here since View doesn't mutate,
	// but we use it for local rendering.
	weeks := availCols

	// Find the start date (the Sunday of the first week)
	// endDate's weekday: 0=Sunday, 6=Saturday
	offsetToSunday := int(m.endDate.Weekday())
	startOfLastWeek := m.endDate.AddDate(0, 0, -offsetToSunday)
	startDate := startOfLastWeek.AddDate(0, 0, -(weeks-1)*7)

	var sb strings.Builder
	sb.WriteString(title + "\n\n")

	// 1. Render Month Header
	sb.WriteString("    ") // Margin for day labels
	var lastMonth time.Month
	for col := 0; col < weeks; {
		cellDate := startDate.AddDate(0, 0, col*7)
		if cellDate.Month() != lastMonth && col < weeks-1 {
			monthStr := cellDate.Format("Jan")
			sb.WriteString(styles.DimItemStyle.Render(monthStr + " ")) // 4 chars wide (takes 2 columns exactly)
			col += 2
			lastMonth = cellDate.Month()
		} else {
			sb.WriteString("  ") // 2 chars wide (1 column)
			col++
		}
	}
	sb.WriteString("\n")

	// Render the 7 days of the week rows
	dayLabels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	for row := 0; row < 7; row++ {
		// Label for the day (only show Mon, Wed, Fri)
		if row == 1 || row == 3 || row == 5 {
			sb.WriteString(styles.DimItemStyle.Render(dayLabels[row] + " "))
		} else {
			sb.WriteString("    ")
		}

		// Render columns (weeks)
		for col := 0; col < weeks; col++ {
			cellDate := startDate.AddDate(0, 0, col*7+row)

			// Don't render future days
			if cellDate.After(m.endDate) {
				sb.WriteString("  ")
				continue
			}

			dateStr := cellDate.Format("2006-01-02")
			count := m.counts[dateStr]

			// Determine color intensity
			cell := "■ "
			var cellStyle lipgloss.Style

			if count == 0 {
				cellStyle = lipgloss.NewStyle().Foreground(styles.Surface) // empty
			} else {
				// 1-4 scale
				scale := 1
				if m.maxVol > 1 {
					ratio := float64(count) / float64(m.maxVol)
					if ratio > 0.75 {
						scale = 4
					} else if ratio > 0.5 {
						scale = 3
					} else if ratio > 0.25 {
						scale = 2
					}
				}
				switch scale {
				case 1:
					cellStyle = lipgloss.NewStyle().Foreground(styles.Muted)
				case 2:
					cellStyle = lipgloss.NewStyle().Foreground(styles.Success)
				case 3:
					cellStyle = lipgloss.NewStyle().Foreground(styles.Primary)
				case 4:
					cellStyle = lipgloss.NewStyle().Foreground(styles.Accent)
				}
			}

			// Highlight cursor cell
			if cellDate.Equal(m.cursorDate) {
				if m.Active {
					cellStyle = cellStyle.Background(styles.Primary).Foreground(lipgloss.Color("#FFFFFF"))
				} else {
					cellStyle = cellStyle.Background(styles.Muted).Foreground(lipgloss.Color("#FFFFFF"))
				}
			}

			sb.WriteString(cellStyle.Render(cell))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	// Footer info
	cursorStr := m.cursorDate.Format("Jan 02, 2006")
	count := m.counts[m.cursorDate.Format("2006-01-02")]
	footer := fmt.Sprintf("  %s: %d blocks created/modified", cursorStr, count)
	sb.WriteString(styles.PrimaryStyle.Render(footer) + "\n")

	if m.Active {
		sb.WriteString(styles.DimItemStyle.Render("\n  [Enter] view log for day  [h/j/k/l] navigate"))
	}

	content := sb.String()

	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}

	return borderStyle.
		Width(m.Width).
		Height(m.Height - 2). // adjust for weird sizing if needed
		Render(content)
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

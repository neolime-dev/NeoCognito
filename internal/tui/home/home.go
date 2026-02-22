// Package home provides the dashboard overlay for NeoCognito.
package home

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// OpenEditorMsg signals the main app to open an editor for a block.
type OpenEditorMsg struct {
	FilePath string
}

// Model represents the home dashboard state.
type Model struct {
	today      []*block.Block
	upcoming   []*block.Block
	monthDates map[int]bool // Days of current month with due tasks

	inboxCount int
	completed  int
	pomodoros  int
	totalPomos int

	cursor       int
	focusSection int // 0 = today, 1 = upcoming

	Width  int
	Height int
	Active bool

	// Callbacks
	GetBlocksFn func() ([]*block.Block, error)
}

// New creates a new Home dashboard model.
func New() Model {
	return Model{
		monthDates: make(map[int]bool),
	}
}

// SetBlocks ingests the global block list and extracts relevant dashboard metrics.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.today = nil
	m.upcoming = nil
	m.monthDates = make(map[int]bool)
	m.inboxCount = 0
	m.completed = 0
	m.pomodoros = 0
	m.totalPomos = 5 // Arbitrary daily target for now

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	sevenDays := todayEnd.Add(6 * 24 * time.Hour)

	for _, b := range blocks {
		// Inbox Count
		if b.Status == "" && !strings.Contains(b.Title, "Daily —") {
			m.inboxCount++
		}

		// Today's Stats
		if b.Modified.After(todayStart) && b.Modified.Before(todayEnd) {
			m.pomodoros += b.Pomodoros
			if b.Status == block.StatusDone {
				m.completed++
			}
		}

		if b.Due != nil {
			// Populate calendar
			if b.Due.Month() == now.Month() && b.Due.Year() == now.Year() {
				m.monthDates[b.Due.Day()] = true
			}

			if b.Status == block.StatusDone || b.Status == block.StatusArchived {
				continue
			}

			if !b.Due.Before(todayStart) && b.Due.Before(todayEnd) {
				m.today = append(m.today, b)
			} else if !b.Due.Before(todayEnd) && b.Due.Before(sevenDays) {
				m.upcoming = append(m.upcoming, b)
			}
		}
	}

	m.clampCursor()
}

func (m *Model) clampCursor() {
	list := m.currentList()
	if len(list) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(list) {
		m.cursor = len(list) - 1
	}
}

func (m Model) currentList() []*block.Block {
	if m.focusSection == 0 {
		return m.today
	}
	return m.upcoming
}

// Init satisfies tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles UI interactions.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.currentList())-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "tab", "right", "l":
			if m.focusSection == 0 {
				m.focusSection = 1
				m.clampCursor()
			}
		case "shift+tab", "left", "h":
			if m.focusSection == 1 {
				m.focusSection = 0
				m.clampCursor()
			}
		case "enter":
			list := m.currentList()
			if len(list) > 0 && m.cursor < len(list) {
				b := list[m.cursor]
				if b.FilePath != "" {
					return m, func() tea.Msg { return OpenEditorMsg{FilePath: b.FilePath} }
				}
			}
		}
	}
	return m, nil
}

// View returns the fully rendered Dashboard.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	// 1. Header Row
	now := time.Now()
	dateStr := now.Format("Monday, 02 January 2006")
	welcome := styles.TitleStyle.Render("󰋜 Welcome back!")
	dateFormatted := styles.NormalModeStyle.Render(fmt.Sprintf(" Today is %s", dateStr))
	header := lipgloss.JoinHorizontal(lipgloss.Center, welcome, dateFormatted)

	innerW := m.Width - 2
	if innerW < 10 {
		innerW = 10
	}

	headerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Width(innerW).
		Padding(0, 1). // Minimal padding for tight header
		Render(header)

	headerH := lipgloss.Height(headerBox)

	innerH := m.Height - 2
	if innerH < 10 {
		innerH = 10
	}

	availBodyH := innerH - headerH
	if availBodyH < 4 {
		availBodyH = 4
	}

	availColsW := innerW - 4 // We have two columns, each with a border (adds 2+2=4 width)
	if availColsW < 10 {
		availColsW = 10
	}

	col1W := availColsW * 3 / 5
	col2W := availColsW - col1W

	availWidgetsH := availBodyH - 4 // Two widgets stacked in left column, minus borders (2+2=4 height)
	if availWidgetsH < 2 {
		availWidgetsH = 2
	}

	col1H1 := availWidgetsH / 2
	col1H2 := availWidgetsH - col1H1

	// 2. Col 1: Today + Upcoming
	todayWidget := m.renderTaskList("🎯 Daily Focus", m.today, 0, col1W, col1H1)
	upcomingWidget := m.renderTaskList("📅 Upcoming", m.upcoming, 1, col1W, col1H2)
	leftCol := lipgloss.JoinVertical(lipgloss.Left, todayWidget, upcomingWidget)

	// 3. Col 2: Calendar + Summary + Shortcuts
	var rightStack []string
	usedRightH := 0

	calendarWidget := m.renderCalendar(col2W)
	calH := lipgloss.Height(calendarWidget)
	if usedRightH+calH <= availBodyH {
		rightStack = append(rightStack, calendarWidget)
		usedRightH += calH
	}

	statsWidget := m.renderStats(col2W)
	statH := lipgloss.Height(statsWidget)
	if usedRightH+statH <= availBodyH {
		rightStack = append(rightStack, statsWidget)
		usedRightH += statH
	}

	shortcutsWidget := m.renderShortcuts(col2W)
	shortH := lipgloss.Height(shortcutsWidget)
	if usedRightH+shortH <= availBodyH {
		rightStack = append(rightStack, shortcutsWidget)
		usedRightH += shortH
	}

	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightStack...)

	// 4. Combine
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	borderFg := styles.Muted
	if m.Active {
		borderFg = styles.Primary
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderFg).
		Width(m.Width).
		Height(m.Height).
		Render(lipgloss.JoinVertical(lipgloss.Left, headerBox, body))
}

func (m Model) renderTaskList(title string, items []*block.Block, sectionID, w, h int) string {
	var sb strings.Builder

	headerStyle := styles.DimItemStyle
	if m.Active && m.focusSection == sectionID {
		headerStyle = styles.PrimaryStyle.Bold(true)
	}
	sb.WriteString("  " + headerStyle.Render(title) + "\n")

	if len(items) == 0 {
		sb.WriteString("  " + styles.DimItemStyle.Render("Nothing scheduled."))
	} else {
		for i, b := range items {
			if i >= h-3 {
				sb.WriteString("  " + styles.DimItemStyle.Render("..."))
				break
			}
			cursor := "  "
			style := styles.NormalItemStyle
			if m.Active && m.focusSection == sectionID && i == m.cursor {
				cursor = "▸ "
				style = styles.SelectedItemStyle
			}

			badge := styles.StatusBadge(b.Status)

			// Optional due date tag for upcoming
			tag := ""
			if sectionID == 1 && b.Due != nil {
				tag = fmt.Sprintf(" (%s)", b.Due.Format("Mon"))
				tag = styles.DimItemStyle.Render(tag)
			}

			titleStr := style.Render(b.Title)
			maxLen := w - 8
			if len(b.Title) > maxLen && maxLen > 0 {
				titleStr = style.Render(b.Title[:maxLen-1] + "…")
			}

			sb.WriteString(fmt.Sprintf("  %s%s %s%s\n", cursor, badge, titleStr, tag))
		}
	}

	borderColor := styles.Muted
	if m.Active && m.focusSection == sectionID {
		borderColor = styles.Primary
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(w).
		Height(h).
		Render(sb.String())
}

func (m Model) renderStats(w int) string {
	var sb strings.Builder
	sb.WriteString("  " + styles.TitleStyle.Render("󰋜 Stats") + "\n")

	// Inbox
	inbColor := styles.NormalModeStyle
	if m.inboxCount > 5 {
		inbColor = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Redish if piled up
	}
	sb.WriteString(fmt.Sprintf("  Inbox items : %s\n", inbColor.Render(fmt.Sprintf("%d", m.inboxCount))))

	// Completed
	sb.WriteString(fmt.Sprintf("  Tasks Done  : %s\n", styles.SelectedItemStyle.Render(fmt.Sprintf("%d", m.completed))))

	// Pomodoros
	pomoStr := "░░░░░"
	if m.pomodoros > 0 {
		fill := m.pomodoros
		if fill > 5 {
			fill = 5
		}
		pomoStr = strings.Repeat("█", fill) + strings.Repeat("░", 5-fill)
	}
	sb.WriteString(fmt.Sprintf("  Pomodoros   : %s %d\n", styles.AccentStyle.Render(pomoStr), m.pomodoros))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Width(w).
		Render(sb.String())
}

func (m Model) renderShortcuts(w int) string {
	var sb strings.Builder
	sb.WriteString("  " + styles.TitleStyle.Render("󰋜 Quick Actions") + "\n")

	shortcuts := []struct{ key, desc string }{
		{"[a]", "Capture to Inbox"},
		{"[d]", "Open Daily Note"},
		{"[/]", "Global Search"},
		{"[ctrl+p]", "Command Palette"},
		{"[Z]", "Inbox Zero Wizard"},
	}

	for _, s := range shortcuts {
		sb.WriteString(fmt.Sprintf("  %-10s %s\n", styles.PrimaryStyle.Render(s.key), styles.DimItemStyle.Render(s.desc)))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Width(w).
		Render(sb.String())
}

func (m Model) renderCalendar(w int) string {
	now := time.Now()
	month := now.Month()
	year := now.Year()

	t1 := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	startWeekday := int(t1.Weekday()) // 0 = Sunday

	// next month 0th day = last day of this month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, now.Location()).Day()

	var sb strings.Builder
	title := fmt.Sprintf("%s %d", month.String(), year)

	// Center title
	padLen := (w - len(title) - 2) / 2
	if padLen < 0 {
		padLen = 0
	}
	sb.WriteString(strings.Repeat(" ", padLen) + styles.TitleStyle.Render(title) + "\n")

	sb.WriteString("  Su Mo Tu We Th Fr Sa\n")

	// Print leading spaces
	sb.WriteString("  ")
	for i := 0; i < startWeekday; i++ {
		sb.WriteString("   ")
	}

	for day := 1; day <= lastDay; day++ {
		var dayStr string

		if day == now.Day() {
			dayStr = styles.PrimaryStyle.Reverse(true).Render(fmt.Sprintf("%2d", day)) + " "
		} else if m.monthDates[day] {
			dayStr = styles.PrimaryStyle.Bold(true).Render(fmt.Sprintf("%2d", day)) + " "
		} else {
			dayStr = styles.NormalItemStyle.Render(fmt.Sprintf("%2d", day)) + " "
		}

		sb.WriteString(dayStr)

		if (day+startWeekday)%7 == 0 {
			sb.WriteString("\n  ")
		}
	}
	sb.WriteString("\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Width(w).
		Render(sb.String())
}

// SetSize updates panel dimensions
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

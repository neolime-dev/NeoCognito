// Package home provides the dashboard overlay for NeoCognito.
package home

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// OpenEditorMsg signals the main app to open an editor for a block.
type OpenEditorMsg struct {
	FilePath string
}

// Model represents the home dashboard state.
type Model struct {
	today      []*block.Block
	upcoming   []*block.Block
	recent     []*block.Block
	monthDates map[int]bool // Days of current month with due tasks

	inboxCount int
	completed  int
	pomodoros  int

	cursor       int
	focusSection int // 0 = today, 1 = upcoming, 2 = recent

	viewing       bool
	viewport      viewport.Model
	rendered      string
	renderer      *glamour.TermRenderer
	rendererWidth int

	Width  int
	Height int
	Active bool

	// Callbacks
	GetBlocksFn func() ([]*block.Block, error)
	LookupFn    func(title string) (*block.Block, error)
}

// New creates a new Home dashboard model.
func New() Model {
	vp := viewport.New(0, 0)
	return Model{
		monthDates: make(map[int]bool),
		viewport:   vp,
	}
}

// SetBlocks ingests the global block list and extracts relevant dashboard metrics.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.today = nil
	m.upcoming = nil
	m.monthDates = make(map[int]bool)
	m.inboxCount = 0
	m.completed = 0

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

	// Determine Recent Notes (those with content, sorted by modified)
	m.recent = nil
	sortable := make([]*block.Block, len(blocks))
	copy(sortable, blocks)
	// Sort by modified DESC
	for i := 0; i < len(sortable); i++ {
		for j := i + 1; j < len(sortable); j++ {
			if sortable[i].Modified.Before(sortable[j].Modified) {
				sortable[i], sortable[j] = sortable[j], sortable[i]
			}
		}
	}
	for _, b := range sortable {
		if len(strings.TrimSpace(b.BodyPreviewText)) > 0 && !strings.Contains(b.Title, "Daily —") {
			m.recent = append(m.recent, b)
			if len(m.recent) >= 10 {
				break
			}
		}
	}

	m.clampCursor()
}

// SetTodayPomodoros sets the number of pomodoro sessions completed in the current app session.
func (m *Model) SetTodayPomodoros(n int) {
	m.pomodoros = n
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
	switch m.focusSection {
	case 0:
		return m.today
	case 1:
		return m.upcoming
	case 2:
		return m.recent
	}
	return m.today
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

	if m.viewing {
		return m.updateViewing(msg)
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
		case "right", "l":
			m.focusSection = (m.focusSection + 1) % 3
			m.clampCursor()
		case "left", "h":
			m.focusSection--
			if m.focusSection < 0 {
				m.focusSection = 2
			}
			m.clampCursor()
		case "enter":
			list := m.currentList()
			if len(list) > 0 && m.cursor < len(list) {
				b := list[m.cursor]
				if len(strings.TrimSpace(b.BodyPreviewText)) > 0 {
					m.renderBlock(b)
					m.viewing = true
				} else if b.FilePath != "" {
					return m, func() tea.Msg { return OpenEditorMsg{FilePath: b.FilePath} }
				}
			}
		}
	}
	return m, nil
}

func (m Model) updateViewing(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "q":
			m.viewing = false
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) renderBlock(b *block.Block) {
	var content strings.Builder
	content.WriteString("# " + b.Title + "\n\n")
	if len(b.Tags) > 0 {
		content.WriteString("Tags: " + strings.Join(b.Tags, ", ") + "\n\n")
	}
	content.WriteString("---\n\n")

	body := m.expandTransclusions(b.Body)
	content.WriteString(body)

	wrapWidth := m.Width - 6
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	if m.renderer == nil || m.rendererWidth != wrapWidth {
		r, err := glamour.NewTermRenderer(
			glamour.WithStyles(styles.MarkdownStyle()),
			glamour.WithWordWrap(wrapWidth),
		)
		if err == nil {
			m.renderer = r
			m.rendererWidth = wrapWidth
		}
	}

	if m.renderer != nil {
		if rendered, err := m.renderer.Render(content.String()); err == nil {
			m.rendered = rendered
		} else {
			m.rendered = content.String()
		}
	} else {
		m.rendered = content.String()
	}

	m.viewport.SetContent(m.rendered)
	m.viewport.GotoTop()
}

func (m *Model) expandTransclusions(text string) string {
	if m.LookupFn == nil {
		return text
	}
	for {
		start := strings.Index(text, "![[")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "]]")
		if end == -1 {
			break
		}
		end += start
		reftitle := text[start+3 : end]
		replacement := "*[could not resolve: " + reftitle + "]*"
		if ref, err := m.LookupFn(reftitle); err == nil && ref != nil {
			replacement = "> **" + ref.Title + "**\n>\n"
			for _, line := range strings.Split(ref.Body, "\n") {
				replacement += "> " + line + "\n"
			}
		}
		text = text[:start] + replacement + text[end+2:]
	}
	return text
}

// View returns the fully rendered Dashboard.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	if m.viewing {
		content := lipgloss.NewStyle().
			Padding(1, 2).
			Render(m.viewport.View())

		return styles.ActiveBorder.
			Width(m.Width).
			Height(m.Height).
			Render(lipgloss.JoinVertical(lipgloss.Left,
				styles.TitleStyle.Render("󰋜 Preview"),
				content,
				"\n "+styles.DimItemStyle.Render("[Esc] back  [j/k] scroll"),
			))
	}

	// 1. Header Row
	now := time.Now()
	dateStr := now.Format("Monday, 02 January 2006")
	welcome := styles.TitleStyle.Render("󰋜 Welcome back!")
	dateFormatted := styles.PrimaryStyle.Render(fmt.Sprintf(" Today is %s", dateStr))
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

	// Adjust for three widgets in col1
	availWidgetsH := availBodyH - 6 // Three widgets stacked in left column, minus borders (3*2=6 height)
	if availWidgetsH < 3 {
		availWidgetsH = 3
	}

	col1H1 := availWidgetsH / 3
	col1H2 := availWidgetsH / 3
	col1H3 := availWidgetsH - col1H1 - col1H2

	// 2. Col 1: Today + Upcoming + Recent
	todayWidget := m.renderTaskList("󰓾 Daily Focus", m.today, 0, col1W, col1H1)
	upcomingWidget := m.renderTaskList("󰃭 Upcoming", m.upcoming, 1, col1W, col1H2)
	recentWidget := m.renderTaskList("󰏫 Recent Notes", m.recent, 2, col1W, col1H3)
	leftCol := lipgloss.JoinVertical(lipgloss.Left, todayWidget, upcomingWidget, recentWidget)

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
		emptyMsg := "Nothing scheduled."
		switch sectionID {
		case 0:
			emptyMsg = "No tasks due today."
		case 1:
			emptyMsg = "None in the next 7 days."
		case 2:
			emptyMsg = "No recent notes."
		}
		sb.WriteString("  " + styles.DimItemStyle.Render(emptyMsg))
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
				tag = fmt.Sprintf(" (%s)", b.Due.Format("02 Jan"))
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
	sb.WriteString("  " + styles.TitleStyle.Render("󰄶 Stats") + "\n")

	// Inbox
	inbColor := styles.SelectedItemStyle
	if m.inboxCount > 5 {
		inbColor = styles.AccentStyle.Bold(true)
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
	sb.WriteString("  " + styles.TitleStyle.Render("󰌌 Quick Actions") + "\n")

	shortcuts := []struct{ key, desc string }{
		{"[a]", "Capture to Inbox"},
		{"[d]", "Open Daily Note"},
		{"[/]", "Global Search"},
		{"[ctrl+p]", "Command Palette"},
		{"[Z]", "Inbox Zero Wizard"},
	}

	for _, s := range shortcuts {
		sb.WriteString(fmt.Sprintf("  %s %s\n", styles.PrimaryStyle.Render(fmt.Sprintf("%-10s", s.key)), styles.DimItemStyle.Render(s.desc)))
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

	m.viewport.Width = w - 4
	m.viewport.Height = h - 6
}

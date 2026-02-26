// Package sidebar provides the navigation sidebar panel.
package sidebar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// View identifiers for sidebar navigation items.
const (
	ViewHome     = "home"
	ViewInbox    = "inbox"
	ViewTasks    = "tasks"
	ViewWiki     = "wiki"
	ViewKanban   = "kanban"
	ViewWaiting  = "waiting"
	ViewJournal  = "journal"
	ViewReview   = "review"
	ViewHeatmap  = "heatmap"
	ViewSearch   = "search"
	ViewProjects = "projects"
	ViewGraphVis = "graphvis"
	ViewTagCloud = "tagcloud"
)

type item struct {
	id       string
	label    string
	icon     string
	isHeader bool
}

var items = []item{
	{label: "DASHBOARD", isHeader: true},
	{id: ViewHome, label: "Home", icon: "󰋜 "},
	{id: ViewInbox, label: "Inbox", icon: "󰙴 "},
	{id: ViewTasks, label: "Tasks", icon: "󰄱 "},
	{id: ViewKanban, label: "Kanban", icon: "󰟀 "},
	{id: ViewWaiting, label: "Waiting", icon: "󰔛 "},
	{label: "AREAS", isHeader: true},
	{id: ViewWiki, label: "Wiki", icon: "󰈙 "},
	{id: ViewProjects, label: "Projects", icon: "󰏗 "},
	{label: "SYSTEM", isHeader: true},
	{id: ViewGraphVis, label: "Graph", icon: "󱎹 "},
	{id: ViewJournal, label: "Journal", icon: "󰏫 "},
	{id: ViewReview, label: "Review", icon: "󰄲 "},
	{id: ViewHeatmap, label: "Heatmap", icon: "󰸗 "},
	{id: ViewSearch, label: "Search", icon: "󰍉 "},
	{id: ViewTagCloud, label: "Tags", icon: "󰓹 "},
}

// Model is the sidebar state.
type Model struct {
	cursor int
	Width  int
	Height int
	Active bool // Whether this panel has focus
}

// ViewChangedMsg is sent when the user selects a different view.
type ViewChangedMsg struct {
	ViewID string
}

// New creates a new sidebar model.
func New() Model {
	// find first non-header
	c := 0
	for i, it := range items {
		if !it.isHeader {
			c = i
			break
		}
	}
	return Model{
		cursor: c,
		Width:  20,
		Active: true,
	}
}

// CurrentView returns the ID of the currently selected view.
func (m Model) CurrentView() string {
	return items[m.cursor].id
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the sidebar.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			target := -1
			for i := m.cursor + 1; i < len(items); i++ {
				if !items[i].isHeader {
					target = i
					break
				}
			}
			if target == -1 { // Wrap to start
				for i := 0; i < m.cursor; i++ {
					if !items[i].isHeader {
						target = i
						break
					}
				}
			}
			if target != -1 {
				m.cursor = target
				return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[m.cursor].id} }
			}
		case "k", "up":
			target := -1
			for i := m.cursor - 1; i >= 0; i-- {
				if !items[i].isHeader {
					target = i
					break
				}
			}
			if target == -1 { // Wrap to end
				for i := len(items) - 1; i > m.cursor; i-- {
					if !items[i].isHeader {
						target = i
						break
					}
				}
			}
			if target != -1 {
				m.cursor = target
				return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[m.cursor].id} }
			}
		case "1":
			m.cursor = 1
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[1].id} }
		case "2":
			m.cursor = 2
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[2].id} }
		case "3":
			m.cursor = 3
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[3].id} }
		case "4":
			m.cursor = 4
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[4].id} }
		case "5":
			m.cursor = 5
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[5].id} }
		case "6":
			m.cursor = 7
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[7].id} }
		case "7":
			m.cursor = 8
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[8].id} }
		case "8":
			m.cursor = 10
			return m, func() tea.Msg { return ViewChangedMsg{ViewID: items[10].id} }
		}
	}
	return m, nil
}

// View renders the sidebar.
func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString("\n")

	for i, it := range items {
		if it.isHeader {
			line := "\n  " + styles.DimItemStyle.Render("▸ "+it.label)
			sb.WriteString(line + "\n")
			continue
		}

		cursor := "  "
		style := styles.NormalItemStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedItemStyle
		}

		label := it.icon + " " + it.label
		line := cursor + style.Render(label)
		sb.WriteString("  " + line + "\n")
	}

	// Add spacing to fill height
	rendered := sb.String()
	lines := strings.Count(rendered, "\n")
	for i := lines; i < m.Height-2; i++ {
		rendered += "\n"
	}

	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}

	return borderStyle.
		Width(m.Width).
		Height(m.Height).
		Render(rendered)
}

// SetSize updates the sidebar dimensions.
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

// borderColor returns the appropriate border color based on focus.
func borderColor(active bool) lipgloss.TerminalColor {
	if active {
		return styles.Primary
	}
	return styles.Muted
}

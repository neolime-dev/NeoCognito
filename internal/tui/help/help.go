// Package help provides the keyboard shortcut help overlay.
package help

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/tui/sidebar"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// Model is the help overlay state.
type Model struct {
	Visible bool
	Width   int
	Height  int
	viewID  string
}

// New creates a new help model.
func New() Model { return Model{} }

// Open shows the help overlay for the given view.
func (m *Model) Open(viewID string) {
	m.Visible = true
	m.viewID = viewID
}

// SetSize updates the overlay dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "?":
			m.Visible = false
		}
	}
	return m, nil
}

// View renders the help overlay.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	title := styles.TitleStyle.Render("? Keybindings") + "  " + styles.DimItemStyle.Render("[esc/q/?] close")

	globalLines := buildSection(globalBindings())
	viewName, viewBnds := viewHelp(m.viewID)
	viewLines := buildSection(viewBnds)

	var body string
	if m.Width >= 70 && viewName != "" {
		halfW := (m.Width - 6) / 2
		leftCol := lipgloss.NewStyle().Width(halfW).Render(
			styles.PrimaryStyle.Bold(true).Render("Global") + "\n" + globalLines,
		)
		rightCol := lipgloss.NewStyle().Width(halfW).Render(
			styles.PrimaryStyle.Bold(true).Render(viewName) + "\n" + viewLines,
		)
		body = lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	} else {
		body = styles.PrimaryStyle.Bold(true).Render("Global") + "\n" + globalLines
		if viewName != "" {
			body += "\n" + styles.PrimaryStyle.Bold(true).Render(viewName) + "\n" + viewLines
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body)
	return styles.ActiveBorder.Width(m.Width).Height(m.Height).
		Render(lipgloss.NewStyle().Padding(0, 1).Render(content))
}

type binding struct{ key, desc string }

func buildSection(bindings []binding) string {
	var sb strings.Builder
	for _, b := range bindings {
		key := fmt.Sprintf("%-12s", b.key) // pad before styling to avoid ANSI length issues
		sb.WriteString("  " + styles.KeyStyle.Render(key) + " " + styles.DescStyle.Render(b.desc) + "\n")
	}
	return sb.String()
}

func globalBindings() []binding {
	return []binding{
		{"Tab", "switch focus"},
		{"d", "daily note"},
		{"/", "search"},
		{"a", "capture to inbox"},
		{"Z", "inbox zero"},
		{"u", "undo"},
		{"Ctrl+R", "redo"},
		{"n", "new from template"},
		{"e", "edit"},
		{"x", "delete"},
		{"H", "history"},
		{"C", "connections"},
		{"r", "related blocks"},
		{"p", "pomodoro"},
		{".", "reset pomodoro"},
		{"Ctrl+P", "command palette"},
		{"q", "quit"},
		{"?", "toggle help"},
	}
}

func viewHelp(viewID string) (string, []binding) {
	switch viewID {
	case sidebar.ViewHome:
		return "Home", []binding{
			{"j/k", "navigate"},
			{"→/l", "next section"},
			{"←/h", "prev section"},
			{"Enter", "view block"},
		}
	case sidebar.ViewInbox:
		return "Inbox", []binding{
			{"j/k", "navigate"},
			{"a", "add item"},
			{"e", "edit"},
			{"x", "delete"},
			{"Space", "multi-select"},
			{"H", "history"},
		}
	case sidebar.ViewTasks:
		return "Tasks", []binding{
			{"j/k", "navigate"},
			{"Space", "cycle status"},
			{"e", "edit"},
			{"x", "delete"},
			{"p", "pomodoro"},
			{"H", "history"},
		}
	case sidebar.ViewKanban:
		return "Kanban", []binding{
			{"j/k", "navigate"},
			{"h/l", "change column"},
			{"Space", "cycle status"},
			{"e", "edit"},
			{"x", "delete"},
		}
	case sidebar.ViewWiki:
		return "Wiki", []binding{
			{"j/k", "navigate"},
			{"Enter", "view"},
			{"e", "edit"},
			{"/", "filter"},
		}
	case sidebar.ViewWaiting:
		return "Waiting", []binding{
			{"j/k", "navigate"},
			{"Enter", "open"},
		}
	case sidebar.ViewJournal:
		return "Journal", []binding{
			{"j/k", "navigate"},
			{"g/G", "first/last"},
			{"Enter", "edit"},
			{"Esc", "go back"},
		}
	case sidebar.ViewHeatmap:
		return "Heatmap", []binding{
			{"j/k/h/l", "navigate"},
			{"Enter", "view day"},
		}
	case sidebar.ViewReview:
		return "Review", []binding{
			{"j/k", "navigate"},
			{"Enter", "open/create"},
		}
	case sidebar.ViewGraphVis:
		return "Graph", []binding{
			{"j/k/h/l", "navigate"},
			{"Enter", "select node"},
		}
	case sidebar.ViewTagCloud:
		return "Tags", []binding{
			{"j/k", "navigate"},
			{"Enter", "filter by tag"},
		}
	case sidebar.ViewSearch:
		return "Search", []binding{
			{"type", "search query"},
			{"Enter", "open result"},
			{"Esc", "close"},
		}
	case sidebar.ViewProjects:
		return "Projects", []binding{
			{"j/k", "navigate"},
			{"Enter", "view"},
			{"e", "edit"},
		}
	default:
		return "", nil
	}
}

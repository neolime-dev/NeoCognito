// Package palette provides a fuzzy-searchable global command palette.
package palette

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// ClosePaletteMsg hides the palette overlay.
type ClosePaletteMsg struct{}

// ExecuteMsg indicates a palette command was selected.
type ExecuteMsg struct {
	Command string
}

// Action represents an entry in the command palette.
type Action struct {
	id          string
	title       string
	description string
}

func (a Action) Title() string       { return a.title }
func (a Action) Description() string { return a.description }
func (a Action) FilterValue() string { return a.title + " " + a.description }

// Model is the palette overlay state.
type Model struct {
	list         list.Model
	mainActions  []list.Item
	themeActions []list.Item
	inSubmenu    bool

	Visible bool
	Width   int
	Height  int
}

// New creates a new command palette model.
func New() Model {
	mainActions := []list.Item{
		Action{id: "new-daily", title: "New Daily Note", description: "Create or open today's daily note"},
		Action{id: "new-weekly", title: "New Weekly Review", description: "Create a weekly review note"},
		Action{id: "menu-themes", title: "Themes...", description: "Change the UI color palette"},
		Action{id: "sync", title: "Run Sync", description: "Force a resync with the filesystem"},
		Action{id: "pomodoro-start", title: "Pomodoro: Start", description: "Start or resume focus timer"},
		Action{id: "pomodoro-stop", title: "Pomodoro: Stop", description: "Pause focus timer"},
		Action{id: "pomodoro-reset", title: "Pomodoro: Reset", description: "Reset focus timer back to 25m"},
		Action{id: "export-html", title: "Export to HTML", description: "Generate a static website from your notes"},
		Action{id: "about", title: " About", description: "NeoCognito v1.0 - Created by Lemon"},
		Action{id: "quit", title: "Quit", description: "Exit NeoCognito"},
	}

	themeActions := []list.Item{
		Action{id: "menu-back", title: "← Back", description: "Return to main menu"},
		Action{id: "theme-tokyo", title: "Tokyo Night", description: "Dark, vibrant blue/purple theme (Default)"},
		Action{id: "theme-catppuccin", title: "Catppuccin", description: "Soft pastel theme"},
		Action{id: "theme-nord", title: "Nord", description: "Arctic, cool cyan-focused theme"},
		Action{id: "theme-gruvbox", title: "Gruvbox", description: "Warm, retro retro groove theme"},
		Action{id: "theme-omarchy", title: "System (Omarchy)", description: "Sync with OS theme dynamically"},
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = styles.SelectedItemStyle.BorderLeftForeground(styles.Primary)
	d.Styles.SelectedDesc = styles.TagStyle.BorderLeftForeground(styles.Primary)

	l := list.New(mainActions, d, 20, 10)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Title = "Command Palette"
	l.Styles.Title = styles.NormalModeStyle

	// Start with filtering active so the user can just type immediately
	l.FilterInput.Focus()

	return Model{
		list:         l,
		mainActions:  mainActions,
		themeActions: themeActions,
		inSubmenu:    false,
		Visible:      false,
	}
}

// Open shows the palette and resets the filter state.
func (m *Model) Open() {
	m.Visible = true
	m.inSubmenu = false
	m.list.Title = "Command Palette"
	m.list.SetItems(m.mainActions)
	m.list.ResetSelected()
	m.list.ResetFilter()
	m.list.FilterInput.Focus()
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages inside the palette overlay.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case styles.ThemeChangedMsg:
		d := list.NewDefaultDelegate()
		d.Styles.SelectedTitle = styles.SelectedItemStyle.BorderLeftForeground(styles.Primary)
		d.Styles.SelectedDesc = styles.TagStyle.BorderLeftForeground(styles.Primary)
		m.list.SetDelegate(d)
		m.list.Styles.Title = styles.NormalModeStyle
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.list.Index() == 0 && m.list.FilterState() != list.Filtering {
				m.list.Select(len(m.list.VisibleItems()) - 1)
				return m, nil
			}
		case "down", "j":
			if m.list.Index() == len(m.list.VisibleItems())-1 && m.list.FilterState() != list.Filtering {
				m.list.Select(0)
				return m, nil
			}
		case "esc", "ctrl+c":
			if m.list.FilterState() == list.Filtering {
				break
			}
			m.Visible = false
			return m, func() tea.Msg { return ClosePaletteMsg{} }

		case "enter":
			if i, ok := m.list.SelectedItem().(Action); ok {
				if i.id == "menu-themes" {
					m.inSubmenu = true
					m.list.Title = "Themes"
					m.list.SetItems(m.themeActions)
					m.list.ResetSelected()
					m.list.ResetFilter()
					return m, nil
				} else if i.id == "menu-back" {
					m.inSubmenu = false
					m.list.Title = "Command Palette"
					m.list.SetItems(m.mainActions)
					m.list.ResetSelected()
					m.list.ResetFilter()
					return m, nil
				}

				m.Visible = false
				return m, tea.Batch(
					func() tea.Msg { return ClosePaletteMsg{} },
					func() tea.Msg { return ExecuteMsg{Command: i.id} },
				)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the palette as a floating window.
func (m Model) View() string {
	if !m.Visible || m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	m.list.SetSize(m.Width-4, m.Height-4)

	content := lipgloss.NewStyle().
		Margin(1, 2).
		Render(m.list.View())

	return styles.ActiveBorder.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}

// SetSize updates the overlay dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

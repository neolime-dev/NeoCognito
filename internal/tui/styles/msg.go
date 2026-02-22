package styles

import tea "github.com/charmbracelet/bubbletea"

// ThemeChangedMsg is fired when the user selects a new color theme.
// TUI components that cache styles (like list.Model) must intercept this
// message and recreate their delegates/styles.
type ThemeChangedMsg struct {
	Theme string
}

// EmitsThemeChanged creates a tea.Cmd that dispatches the ThemeChangedMsg.
func EmitsThemeChanged(theme string) tea.Cmd {
	return func() tea.Msg {
		return ThemeChangedMsg{Theme: theme}
	}
}

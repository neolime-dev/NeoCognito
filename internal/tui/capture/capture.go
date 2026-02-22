package capture

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

type Msg struct {
	Title string
}

type Model struct {
	textInput textinput.Model
	err       error
	Done      bool
	Value     string
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "What's on your mind?"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	return Model{
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.Value = m.textInput.Value()
			m.Done = true
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(styles.Accent)
		return errStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	header := styles.TitleStyle.Render("⚡ Quick Capture")
	input := m.textInput.View()
	help := styles.DimItemStyle.Render(" [enter] save  [esc] cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, header, "\n"+input, "\n"+help))

	return "\n" + box + "\n"
}

func Run() (string, error) {
	m := New()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	return finalModel.(Model).Value, nil
}

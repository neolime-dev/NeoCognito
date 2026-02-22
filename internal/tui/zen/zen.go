package zen

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

type Model struct {
	textarea textarea.Model
	block    *block.Block
	err      error
	width    int
	height   int
	saved    bool
}

func New(b *block.Block) Model {
	ta := textarea.New()
	ta.Placeholder = "Write your thoughts..."
	ta.Focus()
	ta.SetValue(b.Body)

	// Custom styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	return Model{
		textarea: ta,
		block:    b,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+s":
			m.block.Body = m.textarea.Value()
			if err := block.WriteFile(m.block); err != nil {
				m.err = err
			} else {
				m.saved = true
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 6)

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(styles.Accent)
		return errStyle.Render(fmt.Sprintf("\n  Error: %v\n\n  Press esc to quit", m.err))
	}

	header := styles.TitleStyle.Render("🧘 Zen Mode: " + m.block.Title)
	if m.saved {
		successStyle := lipgloss.NewStyle().Foreground(styles.Success)
		header += successStyle.Render(" [Saved]")
	}

	footer := styles.DimItemStyle.Render("  [ctrl+s] save  [esc] quit")

	content := lipgloss.NewStyle().
		Margin(1, 2).
		Render(m.textarea.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		footer,
	)
}

// Run launches the Zen mode as a standalone program
func Run(b *block.Block) error {
	p := tea.NewProgram(New(b), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

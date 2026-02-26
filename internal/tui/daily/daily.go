// Package daily provides the Daily Note panel and creation logic.
package daily

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// OpenEditorMsg signals the app to open this block in $EDITOR.
type OpenEditorMsg struct{ FilePath string }

// dailyTemplate is the default content for a new daily note.
const dailyTemplate = `## 🎯 Top 3 Priorities

1. 
2. 
3. 

## 📝 Notes


## 💭 Reflections

`

// EnsureDailyNote creates today's daily note block if it doesn't exist.
// Returns the block (existing or newly created).
func EnsureDailyNote(blocksDir string, dataDir string) (*block.Block, error) {
	today := time.Now().Format("2006-01-02")
	filename := today + "-daily.md"
	fp := filepath.Join(blocksDir, filename)

	// Already exists — parse and return
	if _, err := os.Stat(fp); err == nil {
		return block.ParseFile(fp)
	}

	// Create new daily note
	b := block.NewBlock(fmt.Sprintf("Daily — %s", today))
	b.Tags = []string{"@daily"}
	b.Body = dailyTemplate
	b.FilePath = fp

	if err := block.WriteFile(b); err != nil {
		return nil, fmt.Errorf("creating daily note: %w", err)
	}
	return b, nil
}

// Model is the daily note panel.
type Model struct {
	block    *block.Block
	viewport viewport.Model
	Visible  bool
	Width    int
	Height   int
}

// New creates a new daily model.
func New() Model {
	vp := viewport.New(0, 0)
	return Model{viewport: vp}
}

// SetBlock loads a block into the panel and renders it.
func (m *Model) SetBlock(b *block.Block) {
	m.block = b
	m.renderContent()
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
		case "esc", "q":
			m.Visible = false
			return m, nil
		case "e":
			if m.block != nil {
				return m, func() tea.Msg { return OpenEditorMsg{FilePath: m.block.FilePath} }
			}
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the daily note panel.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	var sb strings.Builder
	title := "📅 Daily Note"
	if m.block != nil {
		title += " — " + time.Now().Format("02 Jan 2006")
	}
	sb.WriteString(styles.TitleStyle.Render(title) + "\n\n")
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n" + styles.DimItemStyle.Render("  [e] edit  [j/k] scroll  [Esc] close"))

	return styles.ActiveBorder.Width(m.Width).Height(m.Height).Render(sb.String())
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.viewport.Width = w - 4
	m.viewport.Height = h - 6
}

func (m *Model) renderContent() {
	if m.block == nil {
		return
	}
	content := "# " + m.block.Title + "\n\n" + m.block.Body

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(styles.MarkdownStyle()),
		glamour.WithWordWrap(m.Width-6),
	)
	if err != nil {
		m.viewport.SetContent(content)
		return
	}
	rendered, err := renderer.Render(content)
	if err != nil {
		m.viewport.SetContent(content)
		return
	}
	m.viewport.SetContent(rendered)
	m.viewport.GotoTop()
}

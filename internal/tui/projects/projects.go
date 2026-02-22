// Package projects provides a focused view for active projects.
package projects

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// Model is the projects panel state.
type Model struct {
	blocks        []*block.Block
	cursor        int
	viewing       bool
	viewport      viewport.Model
	rendered      string
	renderer      *glamour.TermRenderer
	rendererWidth int
	Width         int
	Height        int
	Active        bool

	LookupFn func(title string) (*block.Block, error)
}

// New creates a new projects model.
func New() Model {
	vp := viewport.New(0, 0)
	return Model{
		viewport: vp,
	}
}

// SetBlocks filters incoming blocks for project tags.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.blocks = nil
	for _, b := range blocks {
		isProject := false
		for _, tag := range b.Tags {
			if strings.ToLower(tag) == "project" || strings.ToLower(tag) == "projects" {
				isProject = true
				break
			}
		}
		if isProject {
			m.blocks = append(m.blocks, b)
		}
	}

	if m.cursor >= len(m.blocks) && len(m.blocks) > 0 {
		m.cursor = len(m.blocks) - 1
	}
}

func (m Model) SelectedBlock() *block.Block {
	if len(m.blocks) == 0 || m.cursor >= len(m.blocks) {
		return nil
	}
	return m.blocks[m.cursor]
}

func (m Model) Init() tea.Cmd { return nil }

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
			if m.cursor < len(m.blocks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if b := m.SelectedBlock(); b != nil {
				m.renderBlock(b)
				m.viewing = true
			}
		case "G":
			if len(m.blocks) > 0 {
				m.cursor = len(m.blocks) - 1
			}
		case "g":
			m.cursor = 0
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
			glamour.WithStandardStyle("dark"),
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

func (m Model) View() string {
	var sb strings.Builder

	header := styles.TitleStyle.Render("󰏗 Projects")
	sb.WriteString(header + "\n")

	if m.viewing {
		sb.WriteString(m.viewport.View())
		sb.WriteString("\n" + styles.DimItemStyle.Render("  [Esc] back  [j/k] scroll"))
	} else {
		if len(m.blocks) == 0 {
			sb.WriteString("\n" + styles.DimItemStyle.Render("  No active projects (tag with #project).") + "\n")
		}

		for i, b := range m.blocks {
			cursor := "  "
			style := styles.NormalItemStyle
			if i == m.cursor && m.Active {
				cursor = "▸ "
				style = styles.SelectedItemStyle
			}

			title := b.Title
			if title == "" {
				title = styles.DimItemStyle.Render("(untitled)")
			}

			line := cursor + style.Render(title)
			if len(b.Tags) > 0 {
				line += " " + styles.TagStyle.Render(strings.Join(b.Tags, " "))
			}
			if len(b.Links) > 0 {
				line += " " + styles.DimItemStyle.Render(fmt.Sprintf("[%d links]", len(b.Links)))
			}
			sb.WriteString(line + "\n")
		}
	}

	content := sb.String()
	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}

	return borderStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}

func (m Model) IsViewing() bool {
	return m.viewing
}

func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.viewport.Width = w - 4
	m.viewport.Height = h - 4
}

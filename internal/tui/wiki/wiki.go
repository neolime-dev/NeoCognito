// Package wiki provides the knowledge base panel with Markdown rendering.
package wiki

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// Model is the wiki panel state.
type Model struct {
	blocks        []*block.Block
	cursor        int
	viewing       bool // true when viewing rendered content
	viewport      viewport.Model
	rendered      string                // rendered markdown
	renderer      *glamour.TermRenderer // cached renderer, nil if stale
	rendererWidth int                   // width at which renderer was last built
	Width         int
	Height        int
	Active        bool

	// LookupFn resolves a block by title (for transclusion). Set by parent.
	LookupFn func(title string) (*block.Block, error)
}

// New creates a new wiki model.
func New() Model {
	vp := viewport.New(0, 0)
	return Model{
		viewport: vp,
	}
}

// SetBlocks replaces the wiki entry list.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.blocks = blocks
	if m.cursor >= len(blocks) && len(blocks) > 0 {
		m.cursor = len(blocks) - 1
	}
}

// SelectedBlock returns the currently highlighted block.
func (m Model) SelectedBlock() *block.Block {
	if len(m.blocks) == 0 || m.cursor >= len(m.blocks) {
		return nil
	}
	return m.blocks[m.cursor]
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
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
	// Build full markdown content
	var content strings.Builder
	content.WriteString("# " + b.Title + "\n\n")
	if len(b.Tags) > 0 {
		content.WriteString("Tags: " + strings.Join(b.Tags, ", ") + "\n\n")
	}
	if len(b.Links) > 0 {
		content.WriteString("Links: " + strings.Join(b.Links, ", ") + "\n\n")
	}
	content.WriteString("---\n\n")

	// Expand ![[title]] transclusions
	body := m.expandTransclusions(b.Body)
	content.WriteString(body)

	// Use cached renderer; rebuild only when width changes.
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

// expandTransclusions replaces ![[Title]] with the referenced block's body.
func (m *Model) expandTransclusions(text string) string {
	if m.LookupFn == nil {
		return text
	}
	// Simple scan for ![[...]]
	for {
		start := strings.Index(text, "![[")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], "]]")
		if end == -1 {
			break
		}
		end += start // absolute end position
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

// View renders the wiki panel.
func (m Model) View() string {
	var sb strings.Builder

	header := styles.TitleStyle.Render("󰈙 Wiki")
	sb.WriteString(header + "\n")

	if m.viewing {
		sb.WriteString(m.viewport.View())
		sb.WriteString("\n" + styles.DimItemStyle.Render("  [Esc] back  [j/k] scroll"))
	} else {
		if len(m.blocks) == 0 {
			sb.WriteString("\n" + styles.DimItemStyle.Render("  No wiki entries.") + "\n")
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

// IsViewing returns whether the wiki is in content viewing mode.
func (m Model) IsViewing() bool {
	return m.viewing
}

// SetSize updates panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.viewport.Width = w - 4
	m.viewport.Height = h - 4
}

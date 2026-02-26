// Package inbox provides the Inbox panel for unprocessed captures.
package inbox

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// BlockCreatedMsg is sent when a new block is successfully captured.
type BlockCreatedMsg struct {
	Block *block.Block
}

// RefreshMsg signals that the block list should be refreshed.
type RefreshMsg struct{}

// Model is the inbox panel state.
type Model struct {
	allBlocks     []*block.Block // full unfiltered list
	blocks        []*block.Block // filtered view
	cursor        int
	visualMode    bool
	visualStart   int
	adding        bool
	filtering     bool
	filterQuery   string
	input         textinput.Model
	viewing       bool
	viewport      viewport.Model
	rendered      string
	renderer      *glamour.TermRenderer
	rendererWidth int
	Width         int
	Height        int
	Active        bool
	Focused       bool

	// Callbacks — set by the parent.
	CreateFn func(title string) (*block.Block, error)
	DeleteFn func(b *block.Block) error
	LookupFn func(title string) (*block.Block, error)
}

// New creates a new inbox model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Capture a thought..."
	ti.CharLimit = 256
	ti.Prompt = "  ✦ "
	ti.PromptStyle = styles.SearchPromptStyle
	ti.TextStyle = styles.NormalItemStyle

	ti.TextStyle = styles.NormalItemStyle

	vp := viewport.New(0, 0)

	return Model{
		input:    ti,
		viewport: vp,
	}
}

// SetBlocks replaces the current block list.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.allBlocks = blocks
	m.applyFilter()
}

// applyFilter narrows m.blocks using m.filterQuery.
func (m *Model) applyFilter() {
	if m.filterQuery == "" {
		m.blocks = m.allBlocks
	} else {
		q := strings.ToLower(m.filterQuery)
		var out []*block.Block
		for _, b := range m.allBlocks {
			if strings.Contains(strings.ToLower(b.Title), q) {
				out = append(out, b)
				continue
			}
			for _, t := range b.Tags {
				if strings.Contains(strings.ToLower(t), q) {
					out = append(out, b)
					break
				}
			}
		}
		m.blocks = out
	}
	if m.cursor >= len(m.blocks) {
		m.cursor = 0
	}
}

// SelectedBlocks returns all currently selected blocks.
func (m Model) SelectedBlocks() []*block.Block {
	if len(m.blocks) == 0 {
		return nil
	}
	if !m.visualMode {
		if m.cursor >= len(m.blocks) {
			return nil
		}
		return []*block.Block{m.blocks[m.cursor]}
	}

	start, end := m.visualStart, m.cursor
	if start > end {
		start, end = end, start
	}
	end++ // inclusive slice bound
	if end > len(m.blocks) {
		end = len(m.blocks)
	}
	return m.blocks[start:end]
}

// SelectedBlock returns the currently hovered block, or nil.
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

	// If adding, delegate to input
	if m.adding {
		return m.updateAdding(msg)
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
			} else {
				m.cursor = 0
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.blocks) > 0 {
				m.cursor = len(m.blocks) - 1
			}
		case "a":
			m.adding = true
			m.input.Focus()
			m.input.SetValue("")
			return m, textinput.Blink
		case "enter":
			if b := m.SelectedBlock(); b != nil {
				if len(strings.TrimSpace(b.BodyPreviewText)) > 0 {
					m.renderBlock(b)
					m.viewing = true
				}
			}
		case "G":
			if len(m.blocks) > 0 {
				m.cursor = len(m.blocks) - 1
			}
		case "g":
			m.cursor = 0
		case "x":
			if b := m.SelectedBlock(); b != nil && m.DeleteFn != nil {
				if err := m.DeleteFn(b); err != nil {
					return m, func() tea.Msg { return err }
				}
				return m, func() tea.Msg { return RefreshMsg{} }
			}
		case "v":
			m.visualMode = !m.visualMode
			if m.visualMode {
				m.visualStart = m.cursor
			}
		case "f":
			m.filtering = true
			m.filterQuery = ""
			m.applyFilter()
			return m, nil
		case "esc":
			if m.visualMode {
				m.visualMode = false
			}
			if m.filtering {
				m.filtering = false
				m.filterQuery = ""
				m.applyFilter()
			}
			return m, nil
		}
	}

	// Filter mode: capture single characters
	if m.filtering {
		if kMsg, ok := msg.(tea.KeyMsg); ok {
			switch kMsg.String() {
			case "backspace":
				if len(m.filterQuery) > 0 {
					m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
					m.applyFilter()
				}
			case "enter", "esc":
				m.filtering = false
			default:
				if r := kMsg.String(); len(r) == 1 {
					m.filterQuery += r
					m.applyFilter()
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

	body := m.expandTransclusions(b.BodyPreviewText)
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

func (m Model) updateAdding(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			title := strings.TrimSpace(m.input.Value())
			if title != "" && m.CreateFn != nil {
				b, err := m.CreateFn(title)
				if err == nil {
					m.adding = false
					m.input.Blur()
					return m, func() tea.Msg { return BlockCreatedMsg{Block: b} }
				}
			}
			m.adding = false
			m.input.Blur()
			return m, nil
		case "esc":
			m.adding = false
			m.input.Blur()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the inbox panel.
func (m Model) View() string {
	var sb strings.Builder

	header := styles.TitleStyle.Render("󰙴 Inbox")
	sb.WriteString(header + "\n")

	if m.viewing {
		sb.WriteString(m.viewport.View())
		sb.WriteString("\n" + styles.DimItemStyle.Render("  [Esc] back  [j/k/up/down] scroll"))
	} else {
		if len(m.blocks) == 0 && !m.adding {
			emptyMsg := styles.DimItemStyle.Render("  No items. Press 'a' to capture.")
			sb.WriteString("\n" + emptyMsg + "\n")
		}

		// Render the add input if active
		if m.adding {
			sb.WriteString("\n" + m.input.View() + "\n\n")
		}

		// Filter indicator
		if m.filterQuery != "" || m.filtering {
			cursor := ""
			if m.filtering {
				cursor = "█"
			}
			indicator := styles.TagStyle.Render(fmt.Sprintf(" filter: %s%s ", m.filterQuery, cursor))
			sb.WriteString(indicator + "\n")
		}

		// Render block list
		maxVisible := m.Height - 6
		if maxVisible < 1 {
			maxVisible = 10
		}
		start := 0
		if m.cursor >= maxVisible {
			start = m.cursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(m.blocks) {
			end = len(m.blocks)
		}

		for i := start; i < end; i++ {
			b := m.blocks[i]
			cursor := "  "
			style := styles.NormalItemStyle

			isSelected := false
			if m.visualMode {
				start, end := m.visualStart, m.cursor
				if start > end {
					start, end = end, start
				}
				if i >= start && i <= end {
					isSelected = true
				}
			}

			if i == m.cursor && m.Active {
				cursor = "▸ "
				if !m.visualMode {
					style = styles.SelectedItemStyle
				} else {
					// distinct style to signify head of selection
					style = styles.PrimaryStyle.Bold(true)
				}
			} else if isSelected {
				style = styles.SelectedItemStyle
			}

			title := b.Title
			if title == "" {
				title = styles.DimItemStyle.Render("(untitled)")
			}

			line := cursor + style.Render(title)
			if len(b.Tags) > 0 {
				tagStr := styles.TagStyle.Render(strings.Join(b.Tags, " "))
				line += " " + tagStr
			}
			sb.WriteString(line + "\n")
		}

		// Scroll indicator
		if len(m.blocks) > maxVisible {
			info := fmt.Sprintf("  %d/%d", m.cursor+1, len(m.blocks))
			sb.WriteString(styles.DimItemStyle.Render(info) + "\n")
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

// IsAdding returns whether the user is in add mode.
func (m Model) IsAdding() bool {
	return m.adding
}

// SetSize updates panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.input.Width = w - 8

	m.viewport.Width = w - 4
	m.viewport.Height = h - 4
}

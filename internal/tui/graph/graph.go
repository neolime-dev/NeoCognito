// Package graph provides the Connections panel showing backlinks and forward-links.
package graph

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// NavigateMsg is sent when the user picks a linked block to view.
type NavigateMsg struct {
	Block *block.Block
}

// Model is the Graph / Connections panel state.
type Model struct {
	subject   *block.Block
	backlinks []*block.Block
	fwdlinks  []*block.Block
	cursor    int // indexes into the combined list (backlinks first)
	Visible   bool
	Width     int
	Height    int

	// Callbacks — set by parent.
	BacklinksFn func(blockID string) ([]*block.Block, error)
	FwdlinksFn  func(blockID string) ([]*block.Block, error)
}

// New creates a new Graph model.
func New() Model {
	return Model{}
}

// Open loads the connections for the given block.
func (m *Model) Open(b *block.Block) {
	m.subject = b
	m.backlinks = nil
	m.fwdlinks = nil
	m.cursor = 0

	if b == nil {
		return
	}
	if m.BacklinksFn != nil {
		m.backlinks, _ = m.BacklinksFn(b.ID)
	}
	if m.FwdlinksFn != nil {
		m.fwdlinks, _ = m.FwdlinksFn(b.ID)
	}
	m.Visible = true
}

// allLinks returns backlinks followed by fwdlinks (combined for cursor nav).
func (m Model) allLinks() []*block.Block {
	var out []*block.Block
	out = append(out, m.backlinks...)
	out = append(out, m.fwdlinks...)
	return out
}

// SelectedBlock returns the currently highlighted linked block.
func (m Model) SelectedBlock() *block.Block {
	all := m.allLinks()
	if len(all) == 0 || m.cursor >= len(all) {
		return nil
	}
	return all[m.cursor]
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles input.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}
	all := m.allLinks()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "C":
			m.Visible = false
			return m, nil
		case "j", "down":
			if m.cursor < len(all)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if b := m.SelectedBlock(); b != nil {
				m.Visible = false
				return m, func() tea.Msg { return NavigateMsg{Block: b} }
			}
		}
	}
	return m, nil
}

// View renders the Connections panel as a popup.
func (m Model) View() string {
	if !m.Visible || m.subject == nil {
		return ""
	}

	var sb strings.Builder

	title := styles.TitleStyle.Render("󱎹 Connections")
	subject := styles.PrimaryStyle.Bold(true).Render(m.subject.Title)
	sb.WriteString(title + "\n")
	sb.WriteString("  " + subject + "\n\n")

	all := m.allLinks()

	// Backlinks section
	sb.WriteString(styles.DimItemStyle.Render("  ← Backlinks") + "\n")
	if len(m.backlinks) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("    (none)") + "\n")
	}
	for i, b := range m.backlinks {
		cursor := "  "
		style := styles.NormalItemStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedItemStyle
		}
		sb.WriteString(fmt.Sprintf("  %s%s\n", cursor, style.Render(b.Title)))
	}

	sb.WriteString("\n" + styles.DimItemStyle.Render("  → Forward Links") + "\n")
	offset := len(m.backlinks)
	if len(m.fwdlinks) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("    (none)") + "\n")
	}
	for i, b := range m.fwdlinks {
		idx := offset + i
		cursor := "  "
		style := styles.NormalItemStyle
		if idx == m.cursor {
			cursor = "▸ "
			style = styles.SelectedItemStyle
		}
		sb.WriteString(fmt.Sprintf("  %s%s\n", cursor, style.Render(b.Title)))
	}

	if len(all) > 0 {
		sb.WriteString("\n" + styles.DimItemStyle.Render("  [Enter] navigate  [Esc/C] close"))
	}

	return styles.ActiveBorder.
		Width(m.Width).
		Height(m.Height).
		Render(sb.String())
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

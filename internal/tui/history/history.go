// Package history provides TUI navigation of a block's version history.
package history

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// RestoredMsg is sent when the user restores an old version.
type RestoredMsg struct{ Block *block.Block }

// CloseMsg is sent when the history panel is closed.
type CloseMsg struct{}

// Version represents a single historical snapshot.
type Version struct {
	Timestamp time.Time
	FilePath  string
}

// Model is the history panel state.
type Model struct {
	Visible    bool
	block      *block.Block
	versions   []Version
	cursor     int
	confirming bool
	Width      int
	Height     int
}

// New creates a new history model.
func New() Model { return Model{} }

// Open shows the history panel for the given block.
func (m *Model) Open(b *block.Block, dataDir string) {
	m.block = b
	m.cursor = 0
	m.Visible = true
	m.versions = loadVersions(b, dataDir)
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
		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				if len(m.versions) > 0 && m.cursor < len(m.versions) {
					b, err := restoreVersion(m.block, m.versions[m.cursor])
					if err == nil {
						m.Visible = false
						m.confirming = false
						return m, func() tea.Msg { return RestoredMsg{Block: b} }
					}
				}
				m.confirming = false
			case "n", "N", "esc":
				m.confirming = false
			}
			return m, nil
		}
		switch msg.String() {
		case "esc", "q":
			m.Visible = false
			return m, func() tea.Msg { return CloseMsg{} }
		case "j", "down":
			if m.cursor < len(m.versions)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if len(m.versions) > 0 && m.cursor < len(m.versions) {
				m.confirming = true
			}
		}
	}
	return m, nil
}

// View renders the history panel.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.TitleStyle.Render("󰋚 Version History") + "\n")
	if m.block != nil {
		sb.WriteString(styles.DimItemStyle.Render("  Block: "+m.block.Title) + "\n\n")
	}

	if len(m.versions) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("  No history yet (enable git_commit in config).") + "\n")
	}

	for i, v := range m.versions {
		cursor := "  "
		style := styles.NormalItemStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedItemStyle
		}

		label := v.Timestamp.Format("2006-01-02  15:04:05")
		if i == 0 {
			label += styles.DimItemStyle.Render("  (current)")
		}
		sb.WriteString(cursor + style.Render(label) + "\n")
	}

	if m.confirming {
		sb.WriteString("\n" + styles.AccentStyle.Render("  Restore this version? [y] yes  [n] cancel"))
	} else {
		sb.WriteString("\n" + styles.DimItemStyle.Render("  [Enter] restore  [Esc] close"))
	}

	return styles.ActiveBorder.Width(m.Width).Height(m.Height).Render(sb.String())
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) { m.Width = w; m.Height = h }

// --- helpers ---

func loadVersions(b *block.Block, dataDir string) []Version {
	versionsDir := filepath.Join(dataDir, "versions", b.ID)
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil
	}

	var versions []Version
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		// Filename encodes timestamp: 2006-01-02T15-04-05.md
		ts, err := time.Parse("2006-01-02T15-04-05", strings.TrimSuffix(e.Name(), ".md"))
		if err != nil {
			continue
		}
		versions = append(versions, Version{
			Timestamp: ts,
			FilePath:  filepath.Join(versionsDir, e.Name()),
		})
	}

	// Most recent first
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Timestamp.After(versions[j].Timestamp)
	})
	return versions
}

func restoreVersion(b *block.Block, v Version) (*block.Block, error) {
	data, err := os.ReadFile(v.FilePath)
	if err != nil {
		return nil, fmt.Errorf("reading version: %w", err)
	}
	// Overwrite the current file with the historical content
	if err := os.WriteFile(b.FilePath, data, 0644); err != nil {
		return nil, fmt.Errorf("restoring version: %w", err)
	}
	return block.ParseFile(b.FilePath)
}

package tagcloud

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lemondesk/neocognito/internal/store"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// Model is the tag cloud state.
type Model struct {
	store  *store.Store
	Active bool
	Width  int
	Height int

	counts map[string]int
}

// New creates a new tag cloud view.
func New(st *store.Store) Model {
	return Model{store: st}
}

// LoadData refreshes tag counts from the store.
func (m *Model) LoadData() {
	if m.store == nil {
		return
	}
	counts, err := m.store.GetTagCounts()
	if err == nil {
		m.counts = counts
	} else {
		m.counts = make(map[string]int)
	}
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
	return m, nil
}

// View renders the tag cloud panel.
func (m Model) View() string {
	if m.Width == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.TitleStyle.Render("󰓹 Tag Cloud") + "\n\n")

	if len(m.counts) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("  No tags found."))
		borderStyle := styles.InactiveBorder
		if m.Active {
			borderStyle = styles.ActiveBorder
		}
		return borderStyle.Width(m.Width).Height(m.Height).Render(sb.String())
	}

	// Extract and sort tags by frequency descending
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range m.counts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Value == sorted[j].Value {
			return sorted[i].Key < sorted[j].Key
		}
		return sorted[i].Value > sorted[j].Value
	})

	maxCount := sorted[0].Value
	if maxCount == 0 {
		maxCount = 1
	}

	// Simple visual cloud
	for _, kv := range sorted {
		// Calculate color intensity or font-weight analogy based on freq
		ratio := float64(kv.Value) / float64(maxCount)
		tagStr := fmt.Sprintf("#%s (%d)", kv.Key, kv.Value)
		var rendered string

		if ratio > 0.8 {
			rendered = styles.TitleStyle.Render(tagStr)
		} else if ratio > 0.4 {
			rendered = styles.SuccessStyle.Render(tagStr)
		} else if ratio > 0.1 {
			rendered = styles.NormalModeStyle.Render(tagStr)
		} else {
			rendered = styles.DimItemStyle.Render(tagStr)
		}

		sb.WriteString("  " + rendered + "\n")
	}

	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}
	return borderStyle.Width(m.Width).Height(m.Height).Render(sb.String())
}

// SetSize updates dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

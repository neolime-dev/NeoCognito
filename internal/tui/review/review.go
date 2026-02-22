// Package review provides the Weekly Review dashboard.
package review

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

// OpenReviewTemplateMsg triggers the creation of a weekly review note.
type OpenReviewTemplateMsg struct{}

// Model is the Weekly Review dashboard state.
type Model struct {
	createdThisWeek    int
	completedThisWeek  int
	tasksAddedThisWeek int
	activeAreas        map[string]bool
	staleCount         int

	Width  int
	Height int
	Active bool
}

// New creates a new Review model.
func New() Model {
	return Model{
		activeAreas: make(map[string]bool),
	}
}

// SetBlocks processes all blocks and calculates weekly stats.
func (m *Model) SetBlocks(blocks []*block.Block) {
	m.createdThisWeek = 0
	m.completedThisWeek = 0
	m.tasksAddedThisWeek = 0
	m.activeAreas = make(map[string]bool)
	m.staleCount = 0

	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	for _, b := range blocks {
		// Stale block count (assuming 14 days default, hardcoded here for stats)
		if b.IsStale(14) {
			m.staleCount++
		}

		createdRecent := b.Created.After(oneWeekAgo)
		modifiedRecent := b.Modified.After(oneWeekAgo)

		if createdRecent {
			m.createdThisWeek++
			if b.IsTask() {
				m.tasksAddedThisWeek++
			}
		}

		if modifiedRecent && b.Status == block.StatusDone {
			m.completedThisWeek++
		}

		if (createdRecent || modifiedRecent) && b.Area != "" {
			m.activeAreas[b.Area] = true
		}
	}
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r", "enter":
			return m, func() tea.Msg { return OpenReviewTemplateMsg{} }
		}
	}
	return m, nil
}

// View renders the Weekly Review dashboard.
func (m Model) View() string {
	if m.Width <= 0 || m.Height <= 0 {
		return ""
	}

	title := styles.TitleStyle.Render("󰄲 Weekly Review Dashboard")

	var sb strings.Builder
	sb.WriteString(title + "\n\n")

	sb.WriteString(styles.PrimaryStyle.Bold(true).Render("  ACTIVITY LAST 7 DAYS") + "\n\n")

	sb.WriteString(fmt.Sprintf("  %s %d\n", styles.DimItemStyle.Render("Blocks Created:"), m.createdThisWeek))
	sb.WriteString(fmt.Sprintf("  %s  %d\n", styles.DimItemStyle.Render("Tasks Added:   "), m.tasksAddedThisWeek))
	sb.WriteString(fmt.Sprintf("  %s %d\n", lipgloss.NewStyle().Foreground(styles.Success).Render("Tasks Completed:"), m.completedThisWeek))

	sb.WriteString("\n" + styles.PrimaryStyle.Bold(true).Render("  ACTIVE AREAS") + "\n\n")
	if len(m.activeAreas) == 0 {
		sb.WriteString(styles.DimItemStyle.Render("  None") + "\n")
	} else {
		for a := range m.activeAreas {
			sb.WriteString(fmt.Sprintf("  %s %s\n", styles.TagStyle.Render("■"), a))
		}
	}

	sb.WriteString("\n" + styles.PrimaryStyle.Bold(true).Render("  MAINTENANCE") + "\n\n")
	if m.staleCount > 0 {
		sb.WriteString(fmt.Sprintf("  %s %s\n", lipgloss.NewStyle().Foreground(styles.Warning).Render("⚠"),
			styles.NormalItemStyle.Render(fmt.Sprintf("You have %d stale blocks needing attention.", m.staleCount))))
	} else {
		sb.WriteString(fmt.Sprintf("  %s %s\n", lipgloss.NewStyle().Foreground(styles.Success).Render("✔"),
			styles.NormalItemStyle.Render("No stale blocks! Your vault is clean.")))
	}

	sb.WriteString("\n\n")

	if m.Active {
		action := styles.NormalModeStyle.Render(" r / Enter ") + " " + styles.PrimaryStyle.Render("Start Weekly Review Note")
		sb.WriteString("  " + action + "\n")
	}

	content := sb.String()

	borderStyle := styles.InactiveBorder
	if m.Active {
		borderStyle = styles.ActiveBorder
	}

	return borderStyle.
		Width(m.Width).
		Height(m.Height - 2). // inner height
		Render(content)
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// Package tui provides the root Bubbletea application model.
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/cmd"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/config"
	"github.com/neolime-dev/neocognito/internal/export"
	"github.com/neolime-dev/neocognito/internal/info"
	"github.com/neolime-dev/neocognito/internal/recur"
	"github.com/neolime-dev/neocognito/internal/store"
	sy "github.com/neolime-dev/neocognito/internal/sync"
	systemplate "github.com/neolime-dev/neocognito/internal/template"
	"github.com/neolime-dev/neocognito/internal/tui/daily"
	"github.com/neolime-dev/neocognito/internal/tui/graph"
	"github.com/neolime-dev/neocognito/internal/tui/graphvis"
	"github.com/neolime-dev/neocognito/internal/tui/gtd"
	"github.com/neolime-dev/neocognito/internal/tui/heatmap"
	"github.com/neolime-dev/neocognito/internal/tui/history"
	"github.com/neolime-dev/neocognito/internal/tui/home"
	"github.com/neolime-dev/neocognito/internal/tui/inbox"
	"github.com/neolime-dev/neocognito/internal/tui/journal"
	"github.com/neolime-dev/neocognito/internal/tui/kanban"
	"github.com/neolime-dev/neocognito/internal/tui/palette"
	"github.com/neolime-dev/neocognito/internal/tui/pomodoro"
	"github.com/neolime-dev/neocognito/internal/tui/projects"
	"github.com/neolime-dev/neocognito/internal/tui/related"
	"github.com/neolime-dev/neocognito/internal/tui/review"
	"github.com/neolime-dev/neocognito/internal/tui/search"
	"github.com/neolime-dev/neocognito/internal/tui/sidebar"
	"github.com/neolime-dev/neocognito/internal/tui/statusbar"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
	"github.com/neolime-dev/neocognito/internal/tui/tagcloud"
	tpl "github.com/neolime-dev/neocognito/internal/tui/template"
	"github.com/neolime-dev/neocognito/internal/tui/timeline"
	"github.com/neolime-dev/neocognito/internal/tui/waiting"
	"github.com/neolime-dev/neocognito/internal/tui/wiki"
	"github.com/neolime-dev/neocognito/internal/undo"
)

const (
	focusSidebar = 0
	focusMain    = 1
)

// App is the root Bubbletea model.
type App struct {
	sidebar   sidebar.Model
	statusbar statusbar.Model
	home      home.Model
	inbox     inbox.Model
	timeline  timeline.Model
	kanban    kanban.Model
	wiki      wiki.Model
	search    search.Model
	gtd       gtd.Model
	daily     daily.Model
	history   history.Model
	waiting   waiting.Model
	pomodoro  pomodoro.Model
	picker    tpl.Model
	journal   journal.Model
	graph     graph.Model
	graphVis  graphvis.Model
	heatmap   heatmap.Model
	review    review.Model
	palette   palette.Model
	related   related.Model
	projects  projects.Model
	tagCloud  tagcloud.Model

	activeView   string
	previousView string
	focus        int
	mode         string
	width        int
	height       int

	cfg       *config.Config
	store     store.Storer
	engine    *sy.Engine
	undoStack *undo.Stack
	flashMsg  string // brief status message (e.g. "Undone: status change")

	confirmDelete []*block.Block // if non-nil, showing delete confirmation modal
}

// NewApp creates and initialises the root application model.
func NewApp(st store.Storer, engine *sy.Engine, cfg *config.Config) App {
	if cfg == nil {
		cfg = config.Default()
	}

	app := App{
		sidebar:    sidebar.New(),
		statusbar:  statusbar.New(),
		home:       home.New(),
		inbox:      inbox.New(),
		timeline:   timeline.New(),
		kanban:     kanban.New(),
		wiki:       wiki.New(),
		search:     search.New(),
		gtd:        gtd.New(),
		daily:      daily.New(),
		history:    history.New(),
		waiting:    waiting.New(),
		pomodoro:   pomodoro.New(),
		picker:     tpl.New(),
		journal:    journal.New(),
		graph:      graph.New(),
		graphVis:   graphvis.New(),
		heatmap:    heatmap.New(),
		review:     review.New(),
		palette:    palette.New(),
		related:    related.New(),
		projects:   projects.New(),
		activeView: sidebar.ViewHome,
		focus:      focusSidebar,
		mode:       "NORMAL",
		cfg:        cfg,
		store:      st,
		engine:     engine,
		undoStack:  undo.New(50),
	}

	app.inbox.CreateFn = func(title string) (*block.Block, error) {
		return engine.CreateBlock(title)
	}

	// Timeline status cycling — wrap with undo and recurrence
	app.timeline.UpdateFn = func(b *block.Block) error {
		prevStatus := ""
		if bl, err := st.GetBlock(b.ID); err == nil && bl != nil {
			prevStatus = bl.Status
		}
		newStatus := b.Status
		blockID := b.ID

		if err := engine.UpdateBlock(b); err != nil {
			return err
		}

		// If just completed AND has recurrence — spawn next occurrence
		if newStatus == block.StatusDone && b.Recur != "" && b.Due != nil {
			if nextDue, err := recur.NextDue(b.Recur, *b.Due); err == nil {
				next := block.NewBlock(b.Title)
				next.Status = block.StatusTodo
				next.Tags = append([]string{}, b.Tags...)
				next.Recur = b.Recur
				next.DelegatedTo = b.DelegatedTo
				next.Area = b.Area
				next.Due = nextDue
				next.FilePath = filepath.Join(engine.BlocksDir(), next.ID+".md")
				versionsDir := filepath.Join(engine.DataDir(), "versions")
				if err := block.WriteFileVersioned(next, versionsDir); err == nil {
					_ = engine.IndexFile(next.FilePath)
				}
			}
		}

		// Push undo operation
		app.undoStack.Push(undo.Operation{
			Description: fmt.Sprintf("status %s→%s on '%s'", prevStatus, newStatus, b.Title),
			Undo: func() error {
				if bl, err := st.GetBlock(blockID); err == nil && bl != nil {
					bl.Status = prevStatus
					return engine.UpdateBlock(bl)
				}
				return nil
			},
			Redo: func() error {
				if bl, err := st.GetBlock(blockID); err == nil && bl != nil {
					bl.Status = newStatus
					return engine.UpdateBlock(bl)
				}
				return nil
			},
		})
		return nil
	}

	app.kanban.UpdateFn = app.timeline.UpdateFn
	app.gtd.UpdateFn = app.timeline.UpdateFn
	app.gtd.DeleteFn = func(b *block.Block) error {
		return engine.DeleteBlock(b)
	}

	app.search.SearchFn = func(query string) ([]*block.Block, error) {
		return st.Search(query)
	}

	app.related.FindRelatedFn = func(target *block.Block, limit int) ([]*block.Block, error) {
		return st.FindRelatedBlocks(target, limit)
	}

	// Wiki and Projects transclusion
	app.wiki.LookupFn = func(title string) (*block.Block, error) {
		b, err := st.GetBlockByTitle(title)
		if err != nil {
			return nil, err
		}
		// Load full body from disk
		if b != nil && b.FilePath != "" {
			if full, err := block.ParseFile(b.FilePath); err == nil {
				return full, nil
			}
		}
		return b, nil
	}
	app.projects.LookupFn = app.wiki.LookupFn
	app.inbox.LookupFn = app.wiki.LookupFn
	app.home.LookupFn = app.wiki.LookupFn

	// Graph connections
	app.graph.BacklinksFn = func(blockID string) ([]*block.Block, error) {
		ids, err := st.GetLinksTo(blockID)
		if err != nil {
			return nil, err
		}
		var out []*block.Block
		for _, id := range ids {
			if b, err := st.GetBlock(id); err == nil && b != nil {
				out = append(out, b)
			}
		}
		return out, nil
	}
	app.graph.FwdlinksFn = func(blockID string) ([]*block.Block, error) {
		ids, err := st.GetLinksFrom(blockID)
		if err != nil {
			return nil, err
		}
		var out []*block.Block
		for _, id := range ids {
			if b, err := st.GetBlock(id); err == nil && b != nil {
				out = append(out, b)
			}
		}
		return out, nil
	}

	app.graphVis.GetBlocksFn = func() ([]*block.Block, error) {
		return st.ListBlocks(store.Filter{})
	}
	app.graphVis.GetEdgesFn = func() ([]store.GraphEdge, error) {
		return st.GetGraphEdges()
	}
	app.graphVis.OnSelect = func(b *block.Block) tea.Cmd {
		if b != nil && b.FilePath != "" {
			return app.openEditor(b.FilePath)
		}
		return nil
	}

	return app
}

// Init satisfies tea.Model.
func (a App) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return refreshDataMsg{} },
		styles.EmitsThemeChanged(a.cfg.Theme),
	)
}

type refreshDataMsg struct{}
type editorFinishedMsg struct{ err error }
type clearFlashMsg struct{}

// errMsg wraps errors that occur during UI interactions to display them to the user
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// Update handles all incoming messages.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateLayout()
		return a, nil

	case pomodoro.TickMsg:
		var cmd tea.Cmd
		a.pomodoro, cmd = a.pomodoro.Tick()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

	case pomodoro.DoneMsg:
		// Persist incremented pomodoro count
		if msg.Block != nil {
			_ = a.engine.UpdateBlock(msg.Block)
		}
		a.refreshPanelData()
		return a, nil

	case clearFlashMsg:
		a.flashMsg = ""
		return a, nil

	case error:
		a.flashMsg = "✗ Error: " + msg.Error()
		cmds = append(cmds, tea.Tick(4*time.Second, func(_ time.Time) tea.Msg {
			return clearFlashMsg{}
		}))
		return a, tea.Batch(cmds...)

	case tea.KeyMsg:
		// Global shortcuts — always active
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "/", "ctrl+k":
			if a.activeView != sidebar.ViewSearch && !a.inbox.IsAdding() {
				a.previousView = a.activeView
				a.activeView = sidebar.ViewSearch
				a.updateFocusStates()
				return a, a.search.Init()
			}
		case "a", "A":
			// If no overlays or input modes are active, jump to inbox and trigger add mode
			if !a.gtd.Visible && !a.daily.Visible && !a.history.Visible && !a.palette.Visible && !a.related.Visible && a.activeView != sidebar.ViewSearch && !a.picker.Visible && !a.inbox.IsAdding() && a.confirmDelete == nil {
				a.activeView = sidebar.ViewInbox
				a.focus = focusMain
				a.updateFocusStates()
				a.refreshPanelData()

				var cmd tea.Cmd
				a.inbox, cmd = a.inbox.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
				return a, cmd
			}
		case "d", "D":
			if !a.gtd.Visible && !a.daily.Visible && !a.history.Visible && !a.palette.Visible && !a.related.Visible && a.activeView != sidebar.ViewSearch && !a.picker.Visible && !a.inbox.IsAdding() && a.confirmDelete == nil {
				return a, a.openDailyNote()
			}
		case "u":
			if !a.inbox.IsAdding() && a.activeView != sidebar.ViewSearch {
				label, err := a.undoStack.Undo()
				if err == nil && label != "" {
					a.flashMsg = "↩ Undone: " + label
					a.refreshPanelData()
					cmds = append(cmds, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
						return clearFlashMsg{}
					}))
				}
				return a, tea.Batch(cmds...)
			}
		case "ctrl+r":
			if !a.inbox.IsAdding() && a.activeView != sidebar.ViewSearch {
				label, err := a.undoStack.Redo()
				if err == nil && label != "" {
					a.flashMsg = "↪ Redone: " + label
					a.refreshPanelData()
				}
				return a, nil
			}
		}

		if a.confirmDelete != nil {
			switch msg.String() {
			case "y", "Y":
				count := len(a.confirmDelete)
				var lastErr error
				for _, b := range a.confirmDelete {
					if err := a.engine.DeleteBlock(b); err != nil {
						lastErr = err
					}
				}
				if lastErr != nil {
					a.flashMsg = "✗ Error deleting: " + lastErr.Error()
					cmds = append(cmds, tea.Tick(4*time.Second, func(_ time.Time) tea.Msg {
						return clearFlashMsg{}
					}))
				} else {
					if count == 1 {
						a.flashMsg = "🗑 Deleted: " + a.confirmDelete[0].Title
					} else {
						a.flashMsg = fmt.Sprintf("🗑 Deleted %d blocks", count)
					}
					cmds = append(cmds, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
						return clearFlashMsg{}
					}))
				}
				a.refreshPanelData()
				a.confirmDelete = nil
				return a, tea.Batch(cmds...)
			case "n", "N", "esc", "q":
				a.confirmDelete = nil
				return a, nil
			}
			return a, nil // block other inputs while confirming
		}

		// Delegate to overlays first
		if a.gtd.Visible {
			var cmd tea.Cmd
			a.gtd, cmd = a.gtd.Update(msg)
			if !a.gtd.Visible {
				a.refreshPanelData()
			}
			return a, cmd
		}
		if a.daily.Visible {
			var cmd tea.Cmd
			a.daily, cmd = a.daily.Update(msg)
			if !a.daily.Visible {
				a.refreshPanelData()
			}
			return a, cmd
		}
		if a.history.Visible {
			var cmd tea.Cmd
			a.history, cmd = a.history.Update(msg)
			if !a.history.Visible {
				a.refreshPanelData()
			}
			return a, cmd
		}
		if a.palette.Visible {
			var cmd tea.Cmd
			a.palette, cmd = a.palette.Update(msg)
			return a, cmd
		}
		if a.related.Visible {
			var cmd tea.Cmd
			a.related, cmd = a.related.Update(msg)
			return a, cmd
		}

		if a.inbox.IsAdding() {
			var cmd tea.Cmd
			a.inbox, cmd = a.inbox.Update(msg)
			return a, cmd
		}
		if a.picker.Visible {
			var cmd tea.Cmd
			a.picker, cmd = a.picker.Update(msg)
			return a, cmd
		}

		// Normal keybindings
		switch msg.String() {
		case "x":
			blocks := a.selectedBlocks()
			if len(blocks) > 0 {
				a.confirmDelete = blocks
				return a, nil
			}
		case "n":
			if a.focus == focusMain {
				// Open template picker
				tMgr, err := systemplate.NewManager(filepath.Join(a.engine.DataDir(), "templates"))
				if err == nil {
					if templates, err := tMgr.List(); err == nil && len(templates) > 0 {
						return a, a.picker.Open(templates)
					}
				}
			}
		case "q":
			return a, tea.Quit
		case "r":
			if b := a.selectedBlock(); b != nil {
				a.related.Open(b)
				return a, nil
			}
		case "Z": // Start Inbox Zero Flow
			noStatus := ""
			blocks, _ := a.store.ListBlocks(store.Filter{Status: &noStatus})
			if len(blocks) > 0 {
				a.gtd.Start(blocks)
			} else {
				a.flashMsg = "Inbox is already empty!"
				a.refreshPanelData()
			}
			return a, nil
		case "ctrl+p":
			a.palette.Open()
			return a, nil

		case "tab":
			a.cycleFocus()
			return a, nil
		case "e":
			if b := a.selectedBlock(); b != nil && b.FilePath != "" {
				return a, a.openEditor(b.FilePath)
			}
		case "H":
			if b := a.selectedBlock(); b != nil {
				a.history.Open(b, a.engine.DataDir())
				return a, nil
			}
		case "p":
			// Start / pause Pomodoro on selected block
			if a.pomodoro.IsRunning() {
				a.pomodoro.Pause()
				return a, nil
			}
			if b := a.selectedBlock(); b != nil {
				a.pomodoro.Start(b)
				return a, tea.Tick(time.Millisecond, func(_ time.Time) tea.Msg {
					return pomodoro.TickMsg(time.Now())
				})
			}
		case ".":
			// Reset pomodoro
			a.pomodoro.Reset()
			return a, nil
		case "C":
			// Open Graph / Connections panel for selected block
			if b := a.selectedBlock(); b != nil {
				a.graph.Open(b)
				return a, nil
			}
		}

		if a.focus == focusSidebar {
			var cmd tea.Cmd
			a.sidebar, cmd = a.sidebar.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			cmds = append(cmds, a.updateActivePanel(msg))
		}

	case sidebar.ViewChangedMsg:
		a.previousView = a.activeView
		a.activeView = msg.ViewID
		if a.activeView == sidebar.ViewSearch {
			cmds = append(cmds, a.search.Init())
		}
		a.updateFocusStates()
		a.refreshPanelData()

	case styles.ThemeChangedMsg:
		// Forward ThemeChangedMsg to panels that cache styles
		a.palette, _ = a.palette.Update(msg)

		// Some Bubbletea inputs or view caching might need a dirty state trigger,
		// but refreshing panel data is a good catch-all for custom UI layouts
		a.refreshPanelData()

	case home.OpenEditorMsg:
		if msg.FilePath != "" {
			return a, a.openEditor(msg.FilePath)
		}

	case inbox.BlockCreatedMsg:
		a.refreshPanelData()

	case timeline.StatusCycledMsg:
		a.refreshPanelData()

	case kanban.StatusCycledMsg:
		a.refreshPanelData()

	case history.RestoredMsg:
		if msg.Block != nil {
			_ = a.engine.IndexFile(msg.Block.FilePath)
		}
		a.refreshPanelData()

	case daily.OpenEditorMsg:
		return a, a.openEditor(msg.FilePath)

	case waiting.SelectedMsg:
		if msg.Block != nil && msg.Block.FilePath != "" {
			return a, a.openEditor(msg.Block.FilePath)
		}

	case search.CloseSearchMsg:
		a.activeView = a.previousView
		a.updateFocusStates()

	case search.SelectResultMsg:
		if msg.Block != nil && msg.Block.FilePath != "" {
			return a, a.openEditor(msg.Block.FilePath)
		}

	case journal.OpenEditorMsg:
		if msg.FilePath != "" {
			return a, a.openEditor(msg.FilePath)
		}

	case graph.NavigateMsg:
		if msg.Block != nil {
			a.activeView = sidebar.ViewWiki
			a.updateFocusStates()
			a.refreshPanelData()
		}

	case heatmap.ViewDayMsg:
		a.previousView = a.activeView
		a.activeView = sidebar.ViewJournal
		a.updateFocusStates()
		a.refreshPanelData()

	case journal.GoBackMsg:
		if a.previousView != "" {
			a.activeView = a.previousView
			a.previousView = ""
		}
		a.updateFocusStates()
		a.refreshPanelData()

	case review.OpenReviewTemplateMsg:
		b, err := a.engine.CreateBlock(time.Now().Format("Weekly Review - 2006-01-02"))
		if err == nil && b != nil && b.FilePath != "" {
			template := "\n## What went well?\n- \n\n## Needs Improvement\n- \n\n## Next Actions\n- \n"

			// Append template to the newly created file
			f, _ := os.OpenFile(b.FilePath, os.O_APPEND|os.O_WRONLY, 0644)
			if f != nil {
				f.WriteString(template)
				f.Close()
			}

			a.refreshPanelData()
			return a, a.openEditor(b.FilePath)
		}

	case palette.ClosePaletteMsg:
		a.updateFocusStates()

	case palette.ExecuteMsg:
		switch msg.Command {
		case "new-daily":
			return a, func() tea.Msg { return daily.OpenEditorMsg{FilePath: ""} }
		case "new-weekly":
			return a, func() tea.Msg { return review.OpenReviewTemplateMsg{} }
		case "sync":
			if err := cmd.RunSync(a.cfg.DataDir, a.cfg); err != nil {
				a.flashMsg = "✗ Sync error: " + err.Error()
			} else {
				a.flashMsg = "✓ Synced successfully"
			}
			_ = a.engine.FullScan()
			a.refreshPanelData()
		case "about":
			a.flashMsg = fmt.Sprintf(" NeoCognito %s (built %s) by Lemon", info.Version, info.BuildDate)
		case "theme-tokyo":
			a.cfg.Theme = "tokyo-night"
			if err := a.cfg.Save(); err != nil {
				a.flashMsg = "Save err: " + err.Error()
			}
			styles.LoadTheme("tokyo-night")
			return a, styles.EmitsThemeChanged("tokyo-night")
		case "theme-catppuccin":
			a.cfg.Theme = "catppuccin"
			if err := a.cfg.Save(); err != nil {
				a.flashMsg = "Save err: " + err.Error()
			}
			styles.LoadTheme("catppuccin")
			return a, styles.EmitsThemeChanged("catppuccin")
		case "theme-nord":
			a.cfg.Theme = "nord"
			if err := a.cfg.Save(); err != nil {
				a.flashMsg = "Save err: " + err.Error()
			}
			styles.LoadTheme("nord")
			return a, styles.EmitsThemeChanged("nord")
		case "theme-gruvbox":
			a.cfg.Theme = "gruvbox"
			if err := a.cfg.Save(); err != nil {
				a.flashMsg = "Save err: " + err.Error()
			}
			styles.LoadTheme("gruvbox")
			return a, styles.EmitsThemeChanged("gruvbox")
		case "theme-omarchy":
			a.cfg.Theme = "omarchy"
			if err := a.cfg.Save(); err != nil {
				a.flashMsg = "Save err: " + err.Error()
			}
			styles.LoadTheme("omarchy")
			return a, styles.EmitsThemeChanged("omarchy")
		case "pomodoro-start":
			if b := a.selectedBlock(); b != nil {
				a.pomodoro.Start(b)
			}
		case "pomodoro-stop":
			a.pomodoro.Pause()
		case "pomodoro-reset":
			a.pomodoro.Reset()
		case "export-html":
			outDir := filepath.Join(a.engine.DataDir(), "export")
			if err := export.Run(a.store, outDir); err == nil {
				a.flashMsg = "✓ Exported to " + outDir
			} else {
				a.flashMsg = "✗ Export failed: " + err.Error()
			}
		case "quit":
			return a, tea.Quit
		}
		a.updateFocusStates()

	case related.CloseRelatedMsg:
		a.updateFocusStates()

	case related.SelectResultMsg:
		a.activeView = sidebar.ViewWiki
		a.updateFocusStates()
		a.refreshPanelData()
		return a, a.openEditor(msg.Block.FilePath)

	case tpl.SelectedMsg:
		// Instantiate new block from template
		b := msg.Template.Instantiate()
		b.FilePath = filepath.Join(a.engine.BlocksDir(), b.ID+".md")
		_ = block.WriteFileVersioned(b, filepath.Join(a.engine.DataDir(), "versions"))
		_ = a.engine.IndexFile(b.FilePath) // force immediate index so we can open it
		a.refreshPanelData()
		return a, a.openEditor(b.FilePath)

	case tpl.CloseMsg:
		a.refreshPanelData()

	case refreshDataMsg:
		a.refreshPanelData()

	case editorFinishedMsg:
		a.refreshPanelData()
		return a, nil
	}

	return a, tea.Batch(cmds...)
}

// View renders the entire application.
func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	title := styles.TitleStyle.Render("⚡ NeoCognito")
	// Show flash message if present
	if a.flashMsg != "" {
		title += "  " + styles.DimItemStyle.Render(a.flashMsg)
	}
	titleBar := lipgloss.PlaceHorizontal(a.width, lipgloss.Left, title)

	sidebarView := a.sidebar.View()
	mainView := a.renderActivePanel()
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainView)

	// Overlays
	if a.confirmDelete != nil {
		var prompt string
		if len(a.confirmDelete) == 1 {
			title := a.confirmDelete[0].Title
			if title == "" {
				title = "(untitled)"
			}
			prompt = fmt.Sprintf("\n  Delete block?\n\n  %s  \n\n  [y/N]  \n", styles.SelectedItemStyle.Render(title))
		} else {
			prompt = fmt.Sprintf("\n  Delete %d selected blocks?\n\n  [y/N]  \n", len(a.confirmDelete))
		}
		box := styles.ActiveBorder.
			BorderForeground(styles.Accent).
			Render(prompt)
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, box)
	} else if a.gtd.Visible {
		overlay := a.gtd.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.picker.Visible {
		overlay := a.picker.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.daily.Visible {
		overlay := a.daily.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.history.Visible {
		overlay := a.history.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)

	} else if a.graph.Visible {
		overlay := a.graph.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.related.Visible {
		overlay := a.related.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)
	} else if a.palette.Visible {
		overlay := a.palette.View()
		content = lipgloss.Place(a.width, a.height-3, lipgloss.Center, lipgloss.Center, overlay)
	}

	a.statusbar.SetWidth(a.width)
	a.statusbar.Mode = a.mode
	a.statusbar.PomodoroSegment = a.pomodoro.StatusBarSegment()
	a.updateStatusBindings()

	return lipgloss.JoinVertical(lipgloss.Left, titleBar, content, a.statusbar.View())
}

// --- layout ---

func (a *App) updateLayout() {
	sidebarWidth := 22
	mainWidth := a.width - sidebarWidth - 4
	mainHeight := a.height - 4 // Title(1) + Status(1) + Borders(2) = 4

	a.sidebar.SetSize(sidebarWidth, mainHeight)
	a.home.SetSize(mainWidth, mainHeight)
	a.inbox.SetSize(mainWidth, mainHeight)
	a.timeline.SetSize(mainWidth, mainHeight)
	a.kanban.SetSize(mainWidth, mainHeight)
	a.wiki.SetSize(mainWidth, mainHeight)
	a.projects.SetSize(mainWidth, mainHeight)
	a.search.SetSize(mainWidth, mainHeight)
	a.waiting.SetSize(mainWidth, mainHeight)
	a.journal.SetSize(mainWidth, mainHeight)

	overlayW := a.width * 2 / 3
	overlayH := a.height * 2 / 3
	a.gtd.SetSize(overlayW, overlayH)
	a.picker.SetSize(overlayW, overlayH)
	a.daily.SetSize(overlayW, overlayH)
	a.history.SetSize(overlayW, overlayH/2)
	a.graph.SetSize(overlayW, overlayH/2)
	a.palette.SetSize(overlayW, overlayH)
	a.heatmap.SetSize(mainWidth, mainHeight)
	a.review.SetSize(mainWidth, mainHeight)
	a.related.SetSize(overlayW, overlayH)
	a.graphVis.SetSize(mainWidth, mainHeight)
	a.tagCloud.SetSize(mainWidth, mainHeight)
}

func (a *App) cycleFocus() {
	if a.focus == focusSidebar {
		a.focus = focusMain
	} else {
		a.focus = focusSidebar
	}
	a.updateFocusStates()
}

func (a *App) updateFocusStates() {
	a.sidebar.Active = (a.focus == focusSidebar)
	a.home.Active = (a.focus == focusMain && a.activeView == sidebar.ViewHome)
	a.inbox.Active = (a.focus == focusMain && a.activeView == sidebar.ViewInbox)
	a.timeline.Active = (a.focus == focusMain && a.activeView == sidebar.ViewTasks)
	a.kanban.Active = (a.focus == focusMain && a.activeView == sidebar.ViewKanban)
	a.wiki.Active = (a.focus == focusMain && a.activeView == sidebar.ViewWiki)
	a.projects.Active = (a.focus == focusMain && a.activeView == sidebar.ViewProjects)
	a.waiting.Active = (a.focus == focusMain && a.activeView == sidebar.ViewWaiting)
	a.journal.Active = (a.focus == focusMain && a.activeView == sidebar.ViewJournal)
	a.heatmap.Active = (a.focus == focusMain && a.activeView == sidebar.ViewHeatmap)
	a.review.Active = (a.focus == focusMain && a.activeView == sidebar.ViewReview)
	a.graphVis.Active = (a.focus == focusMain && a.activeView == sidebar.ViewGraphVis)
	a.tagCloud.Active = (a.focus == focusMain && a.activeView == sidebar.ViewTagCloud)
	a.search.Active = (a.focus == focusMain && a.activeView == sidebar.ViewSearch)
}

func (a *App) updateActivePanel(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch a.activeView {
	case sidebar.ViewHome:
		a.home, cmd = a.home.Update(msg)
	case sidebar.ViewInbox:
		a.inbox, cmd = a.inbox.Update(msg)
	case sidebar.ViewTasks:
		a.timeline, cmd = a.timeline.Update(msg)
	case sidebar.ViewKanban:
		a.kanban, cmd = a.kanban.Update(msg)
	case sidebar.ViewWiki:
		a.wiki, cmd = a.wiki.Update(msg)
	case sidebar.ViewProjects:
		a.projects, cmd = a.projects.Update(msg)
	case sidebar.ViewWaiting:
		a.waiting, cmd = a.waiting.Update(msg)
	case sidebar.ViewJournal:
		a.journal, cmd = a.journal.Update(msg)
	case sidebar.ViewHeatmap:
		a.heatmap, cmd = a.heatmap.Update(msg)
	case sidebar.ViewReview:
		a.review, cmd = a.review.Update(msg)
	case sidebar.ViewGraphVis:
		a.graphVis, cmd = a.graphVis.Update(msg)
	case sidebar.ViewTagCloud:
		a.tagCloud, cmd = a.tagCloud.Update(msg)
	}
	return cmd
}

func (a App) renderActivePanel() string {
	switch a.activeView {
	case sidebar.ViewHome:
		return a.home.View()
	case sidebar.ViewInbox:
		return a.inbox.View()
	case sidebar.ViewTasks:
		return a.timeline.View()
	case sidebar.ViewKanban:
		return a.kanban.View()
	case sidebar.ViewWiki:
		return a.wiki.View()
	case sidebar.ViewProjects:
		return a.projects.View()
	case sidebar.ViewWaiting:
		return a.waiting.View()
	case sidebar.ViewJournal:
		return a.journal.View()
	case sidebar.ViewHeatmap:
		return a.heatmap.View()
	case sidebar.ViewReview:
		return a.review.View()
	case sidebar.ViewGraphVis:
		return a.graphVis.View()
	case sidebar.ViewTagCloud:
		return a.tagCloud.View()
	case sidebar.ViewSearch:
		return a.search.View()
	default:
		return a.inbox.View()
	}
}

func (a *App) selectedBlock() *block.Block {
	blocks := a.selectedBlocks()
	if len(blocks) > 0 {
		return blocks[0]
	}
	return nil
}

// selectedBlocks returns all currently selected blocks.
func (a *App) selectedBlocks() []*block.Block {
	switch a.activeView {
	case sidebar.ViewHome:
		return nil // No multi-select in Home
	case sidebar.ViewInbox:
		return a.inbox.SelectedBlocks()
	case sidebar.ViewTasks:
		return a.timeline.SelectedBlocks()
	case sidebar.ViewKanban:
		return a.kanban.SelectedBlocks()
	case sidebar.ViewWiki:
		if b := a.wiki.SelectedBlock(); b != nil {
			return []*block.Block{b}
		}
	case sidebar.ViewProjects:
		if b := a.projects.SelectedBlock(); b != nil {
			return []*block.Block{b}
		}
	case sidebar.ViewWaiting:
		if b := a.waiting.SelectedBlock(); b != nil {
			return []*block.Block{b}
		}
	}
	return nil
}

func (a *App) refreshPanelData() {
	blocks, err := a.store.ListBlocks(store.Filter{})
	if err != nil {
		return
	}
	// Partition in-memory — avoids 2 extra DB round-trips per refresh.
	var inboxBlocks, taskBlocks []*block.Block
	for _, b := range blocks {
		if b.Status == "" {
			inboxBlocks = append(inboxBlocks, b)
		} else {
			taskBlocks = append(taskBlocks, b)
		}
	}
	a.inbox.SetBlocks(inboxBlocks)
	a.timeline.SetBlocks(taskBlocks)
	a.kanban.SetBlocks(taskBlocks)
	a.home.SetBlocks(blocks)
	a.wiki.SetBlocks(blocks)
	a.projects.SetBlocks(blocks)
	a.search.SetItems(blocks)
	a.waiting.SetBlocks(blocks)
	a.journal.SetBlocks(blocks)
	a.heatmap.SetBlocks(blocks)
	a.review.SetBlocks(blocks)
	a.graphVis.LoadData()
	a.tagCloud.LoadData()
}

func (a *App) updateStatusBindings() {
	var bindings []statusbar.Binding
	bindings = append(bindings,
		statusbar.Binding{Key: "Tab", Desc: "switch"},
		statusbar.Binding{Key: "d", Desc: "daily"},
		statusbar.Binding{Key: "/", Desc: "search"},
	)
	if a.focus == focusMain {
		switch a.activeView {
		case sidebar.ViewHome:
			bindings = append(bindings,
				statusbar.Binding{Key: "Enter", Desc: "view"},
				statusbar.Binding{Key: "Tab", Desc: "section"},
			)
		case sidebar.ViewInbox:
			bindings = append(bindings,
				statusbar.Binding{Key: "a", Desc: "add"},
				statusbar.Binding{Key: "n", Desc: "new"},
				statusbar.Binding{Key: "Z", Desc: "zero"},
				statusbar.Binding{Key: "e", Desc: "edit"},
				statusbar.Binding{Key: "x", Desc: "delete"},
				statusbar.Binding{Key: "H", Desc: "history"},
			)
		case sidebar.ViewTasks:
			bindings = append(bindings,
				statusbar.Binding{Key: "Space", Desc: "cycle"},
				statusbar.Binding{Key: "n", Desc: "new"},
				statusbar.Binding{Key: "p", Desc: "pomodoro"},
				statusbar.Binding{Key: "e", Desc: "edit"},
				statusbar.Binding{Key: "x", Desc: "delete"},
				statusbar.Binding{Key: "H", Desc: "history"},
			)
		case sidebar.ViewKanban:
			bindings = append(bindings,
				statusbar.Binding{Key: "Space", Desc: "cycle"},
				statusbar.Binding{Key: "h/l", Desc: "column"},
				statusbar.Binding{Key: "n", Desc: "new"},
				statusbar.Binding{Key: "p", Desc: "pomodoro"},
				statusbar.Binding{Key: "e", Desc: "edit"},
				statusbar.Binding{Key: "x", Desc: "delete"},
			)
		case sidebar.ViewWiki:
			bindings = append(bindings,
				statusbar.Binding{Key: "Enter", Desc: "view"},
				statusbar.Binding{Key: "e", Desc: "edit"},
			)
		case sidebar.ViewWaiting:
			bindings = append(bindings,
				statusbar.Binding{Key: "Enter", Desc: "open"},
			)
		}
	} else {
		bindings = append(bindings, statusbar.Binding{Key: "j/k", Desc: "navigate"})
	}
	if a.undoStack.CanUndo() {
		bindings = append(bindings, statusbar.Binding{Key: "u", Desc: "undo"})
	}
	bindings = append(bindings, statusbar.Binding{Key: "q", Desc: "quit"})
	a.statusbar.SetBindings(bindings)
}

func (a *App) openDailyNote() tea.Cmd {
	b, err := a.engine.EnsureDailyNote()
	if err != nil {
		return nil
	}
	_ = a.engine.IndexFile(b.FilePath)
	a.daily.SetBlock(b)
	a.daily.Visible = true
	return nil
}

func (a *App) openEditor(filepath string) tea.Cmd {
	editor := a.cfg.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
	}
	c := exec.Command(editor, filepath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

// FormatBlockCount returns a human-readable block count.
func (a App) FormatBlockCount() string {
	count, _ := a.store.BlockCount()
	if count == 0 {
		return "No blocks"
	}
	return fmt.Sprintf("%d blocks", count)
}

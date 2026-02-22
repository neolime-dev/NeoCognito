// Package cmd provides CLI command handling.
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/config"
	"github.com/lemondesk/neocognito/internal/export"
	"github.com/lemondesk/neocognito/internal/store"
	sy "github.com/lemondesk/neocognito/internal/sync"
	"github.com/lemondesk/neocognito/internal/template"
	"github.com/lemondesk/neocognito/internal/tui/capture"
	"github.com/lemondesk/neocognito/internal/tui/styles"
	"github.com/lemondesk/neocognito/internal/tui/zen"
)

// EnsureDataDirs creates the blocks, assets, and versions directories if needed.
func EnsureDataDirs(dataDir string) error {
	dirs := []string{
		filepath.Join(dataDir, "blocks"),
		filepath.Join(dataDir, "assets"),
		filepath.Join(dataDir, "versions"),
		filepath.Join(dataDir, "templates"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}

// RunAdd handles `neocognito add "text"` — creates a new inbox block.
// Now accepts config to apply defaults and git-commit.
func RunAdd(dataDir string, title string, cfg *config.Config) error {
	if err := EnsureDataDirs(dataDir); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer st.Close()

	blocksDir := filepath.Join(dataDir, "blocks")
	engine := sy.NewEngine(blocksDir, st, cfg)

	b, err := engine.CreateBlock(title)
	if err != nil {
		return fmt.Errorf("creating block: %w", err)
	}

	fmt.Printf("✦ Created block: %s (%s)\n", b.Title, b.ID)
	if b.Due != nil {
		fmt.Printf("  Due: %s\n", b.Due.Format("2006-01-02"))
	}
	fmt.Printf("  File: %s\n", b.FilePath)
	return nil
}

// RunNew handles `neocognito new --template <name>` — creates a block from template.
func RunNew(dataDir string, templateName string, cfg *config.Config) error {
	if err := EnsureDataDirs(dataDir); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer st.Close()

	blocksDir := filepath.Join(dataDir, "blocks")
	engine := sy.NewEngine(blocksDir, st, cfg)

	// Load template manager
	tMgr, err := template.NewManager(filepath.Join(dataDir, "templates"))
	if err != nil {
		return fmt.Errorf("loading templates: %w", err)
	}

	t, err := tMgr.Get(templateName)
	if err != nil {
		return err // prints "template not found: <name>"
	}

	b := t.Instantiate()
	b.FilePath = filepath.Join(blocksDir, b.ID+".md")

	if err := block.WriteFileVersioned(b, filepath.Join(dataDir, "versions")); err != nil {
		return fmt.Errorf("writing block: %w", err)
	}
	if err := engine.IndexFile(b.FilePath); err != nil {
		return fmt.Errorf("indexing block: %w", err)
	}

	// Make sure we print so scripts can capture it if needed
	fmt.Printf("✦ Created from template '%s': %s\n", templateName, b.Title)
	fmt.Printf("  File: %s\n", b.FilePath)

	// Auto-open in editor
	editor := cfg.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
	}
	cmd := exec.Command(editor, b.FilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunSearch handles `neocognito search "query"` — prints matching blocks.
func RunSearch(dataDir string, query string) error {
	if err := EnsureDataDirs(dataDir); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer st.Close()

	blocksDir := filepath.Join(dataDir, "blocks")
	engine := sy.NewEngine(blocksDir, st, nil)
	_ = engine.FullScan()

	results, err := st.Search(query)
	if err != nil {
		return fmt.Errorf("searching: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("Found %d results for %q:\n\n", len(results), query)
	for _, b := range results {
		status := ""
		if b.Status != "" {
			status = styles.StatusBadge(b.Status) + " "
		}
		tags := ""
		if len(b.Tags) > 0 {
			tags = " [" + strings.Join(b.Tags, ", ") + "]"
		}
		fmt.Printf("  %s%s%s\n", status, b.Title, tags)
		fmt.Printf("    ID: %s  File: %s\n\n", b.ID, b.FilePath)
	}
	return nil
}

// RunClip reads from stdin and creates an inbox block, optionally with a source URL.
func RunClip(dataDir string, url string, cfg *config.Config) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	if err := EnsureDataDirs(dataDir); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	blocksDir := filepath.Join(dataDir, "blocks")
	engine := sy.NewEngine(blocksDir, st, cfg)

	title := "Clipped Content"
	if url != "" {
		title = "Clip: " + url
	}

	b, err := engine.CreateBlock(title)
	if err != nil {
		return err
	}

	// Append clipped content
	body := string(input)
	if url != "" {
		body = "Source: " + url + "\n\n" + body
	}
	b.Body = body

	if err := block.WriteFileVersioned(b, filepath.Join(dataDir, "versions")); err != nil {
		return err
	}

	fmt.Printf("✦ Clipped to: %s (%s)\n", b.Title, b.ID)
	return nil
}

// RunExport triggers the static HTML export via CLI.
func RunExport(dataDir string, targetDir string) error {
	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	if targetDir == "" {
		targetDir = filepath.Join(dataDir, "export")
	}

	fmt.Printf("📦 Exporting to %s...\n", targetDir)
	if err := export.Run(st, targetDir); err != nil {
		return err
	}
	fmt.Println("✓ Export complete!")
	return nil
}

// RunSync executes the remote sync command defined in config.
func RunSync(dataDir string, cfg *config.Config) error {
	if cfg.SyncCmd == "" {
		return fmt.Errorf("sync_cmd not configured in config.toml")
	}
	fmt.Printf("🔄 Running sync: %s\n", cfg.SyncCmd)
	cmd := exec.Command("sh", "-c", cfg.SyncCmd)
	cmd.Dir = dataDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunZen opens a specific block in Zen (distraction-free) mode.
func RunZen(dataDir string, blockID string, cfg *config.Config) error {
	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	var b *block.Block
	if strings.Contains(blockID, "/") || strings.HasSuffix(blockID, ".md") {
		// Assume it's a filepath
		b, err = block.ParseFile(blockID)
	} else {
		b, err = st.GetBlock(blockID)
	}

	if err != nil {
		return fmt.Errorf("block not found: %w", err)
	}

	styles.LoadTheme(cfg.Theme)
	return zen.Run(b)
}

// RunCapture launches the quick capture TUI and saves the result.
func RunCapture(dataDir string, cfg *config.Config) error {
	styles.LoadTheme(cfg.Theme)
	val, err := capture.Run()
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}

	return RunAdd(dataDir, val, cfg)
}

// PrintUsage shows CLI usage information.
func PrintUsage() {
	fmt.Println(`⚡ NeoCognito — Second Brain TUI

Usage:
  neocognito                       Launch the TUI dashboard
  neocognito add "text"            Capture a new block (supports natural language dates)
  neocognito capture               Launch floating quick capture window
  neocognito new --template <name> Create a block from a template and open editor
  neocognito search "q"            Search blocks from the command line
  neocognito clip [--url URL]      Capture content from stdin
  neocognito export [PATH]         Generate static HTML site
  neocognito sync                  Sync data via configured sync_cmd
  neocognito zen <id|path>         Open distraction-free editor
  neocognito config                Show/create the config file

Options:
  --data-dir PATH         Override the data directory

Environment:
  EDITOR                  Editor for deep editing (default: vi)
  XDG_CONFIG_HOME         Base directory for config
  XDG_DATA_HOME           Base directory for data

Examples:
  neocognito add "Call doctor tomorrow"
  neocognito add "Sprint planning next monday"
  neocognito add "Buy milk in 2 days"`)
}

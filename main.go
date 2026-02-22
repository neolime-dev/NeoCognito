package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lemondesk/neocognito/cmd"
	"github.com/lemondesk/neocognito/internal/config"
	"github.com/lemondesk/neocognito/internal/store"
	sy "github.com/lemondesk/neocognito/internal/sync"
	"github.com/lemondesk/neocognito/internal/tui"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

func main() {
	// Load config (falls back to defaults if no file)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	styles.LoadTheme(cfg.Theme)

	dataDir := cfg.DataDir

	// Parse --data-dir flag (overrides config)
	args := os.Args[1:]
	filteredArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--data-dir" && i+1 < len(args) {
			dataDir = args[i+1]
			i++
		} else if strings.HasPrefix(args[i], "--data-dir=") {
			dataDir = strings.TrimPrefix(args[i], "--data-dir=")
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	cfg.DataDir = dataDir

	// Route subcommands
	if len(filteredArgs) == 0 {
		runTUI(cfg)
		return
	}

	switch filteredArgs[0] {
	case "add":
		if len(filteredArgs) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: neocognito add \"block title\"")
			os.Exit(1)
		}
		title := strings.Join(filteredArgs[1:], " ")
		if err := cmd.RunAdd(dataDir, title, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "new":
		if len(filteredArgs) < 3 || filteredArgs[1] != "--template" {
			fmt.Fprintln(os.Stderr, "Usage: neocognito new --template <name>")
			os.Exit(1)
		}
		templateName := filteredArgs[2]
		if err := cmd.RunNew(dataDir, templateName, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "search":
		if len(filteredArgs) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: neocognito search \"query\"")
			os.Exit(1)
		}
		query := strings.Join(filteredArgs[1:], " ")
		if err := cmd.RunSearch(dataDir, query); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "clip":
		url := ""
		for i := 1; i < len(filteredArgs); i++ {
			if filteredArgs[i] == "--url" && i+1 < len(filteredArgs) {
				url = filteredArgs[i+1]
				break
			}
		}
		if err := cmd.RunClip(dataDir, url, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "export":
		target := ""
		if len(filteredArgs) > 1 {
			target = filteredArgs[1]
		}
		if err := cmd.RunExport(dataDir, target); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "sync":
		if err := cmd.RunSync(dataDir, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Sync Error: %v\n", err)
			os.Exit(1)
		}

	case "capture":
		if err := cmd.RunCapture(dataDir, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "zen":
		if len(filteredArgs) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: neocognito zen <id|path>")
			os.Exit(1)
		}
		if err := cmd.RunZen(dataDir, filteredArgs[1], cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "config":
		// Print config path and write default if missing
		if err := config.WriteDefault(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config: %s\n", config.ConfigPath())
		fmt.Printf("Data:   %s\n", cfg.DataDir)
		fmt.Printf("Editor: %s\n", cfg.Editor)
		fmt.Printf("Git:    %v\n", cfg.GitCommit)

	case "help", "--help", "-h":
		cmd.PrintUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", filteredArgs[0])
		cmd.PrintUsage()
		os.Exit(1)
	}
}

func runTUI(cfg *config.Config) {
	dataDir := cfg.DataDir

	if err := cmd.EnsureDataDirs(dataDir); err != nil {
		log.Fatalf("Error creating data directories: %v", err)
	}

	dbPath := filepath.Join(dataDir, "index.db")
	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer st.Close()

	blocksDir := filepath.Join(dataDir, "blocks")
	engine := sy.NewEngine(blocksDir, st, cfg)

	if err := engine.FullScan(); err != nil {
		log.Printf("Warning: initial scan failed: %v", err)
	}

	if err := engine.Watch(); err != nil {
		log.Printf("Warning: file watcher failed: %v", err)
	}
	defer engine.Stop()

	app := tui.NewApp(st, engine, cfg)
	p := tea.NewProgram(app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}

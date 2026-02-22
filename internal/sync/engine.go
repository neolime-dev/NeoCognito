// Package sync handles file-to-database synchronization.
package sync

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lemondesk/neocognito/internal/block"
	"github.com/lemondesk/neocognito/internal/config"
	gitpkg "github.com/lemondesk/neocognito/internal/git"
	"github.com/lemondesk/neocognito/internal/nldate"
	"github.com/lemondesk/neocognito/internal/store"
)

// Engine synchronizes Markdown block files with the SQLite index.
type Engine struct {
	blocksDir string
	dataDir   string
	store     *store.Store
	cfg       *config.Config
	watcher   *fsnotify.Watcher
	done      chan struct{}
}

// NewEngine creates a new sync engine for the given blocks directory and store.
func NewEngine(blocksDir string, st *store.Store, cfg *config.Config) *Engine {
	dataDir := filepath.Dir(blocksDir)
	return &Engine{
		blocksDir: blocksDir,
		dataDir:   dataDir,
		store:     st,
		cfg:       cfg,
		done:      make(chan struct{}),
	}
}

// FullScan walks the blocks directory and upserts every .md file into the index.
// This is used at startup and for index rebuilds.
func (e *Engine) FullScan() error {
	count := 0
	err := filepath.Walk(e.blocksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		if err := e.indexFile(path); err != nil {
			log.Printf("warn: failed to index %s: %v", path, err)
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking blocks dir: %w", err)
	}
	log.Printf("sync: indexed %d blocks from %s", count, e.blocksDir)
	return nil
}

// IndexFile parses a single file and upserts it into the store.
func (e *Engine) IndexFile(path string) error {
	return e.indexFile(path)
}

func (e *Engine) indexFile(path string) error {
	b, err := block.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}
	return e.store.UpsertBlock(b)
}

// Watch starts watching the blocks directory for changes and re-indexes files.
func (e *Engine) Watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	e.watcher = watcher

	err = filepath.Walk(e.blocksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		watcher.Close()
		return fmt.Errorf("adding watch paths: %w", err)
	}

	go e.watchLoop()
	return nil
}

func (e *Engine) watchLoop() {
	for {
		select {
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}
			if !strings.HasSuffix(event.Name, ".md") {
				continue
			}
			switch {
			case event.Has(fsnotify.Write), event.Has(fsnotify.Create):
				if err := e.indexFile(event.Name); err != nil {
					log.Printf("sync: error indexing %s: %v", event.Name, err)
				}
			case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
				b, err := e.store.GetBlockByPath(event.Name)
				if err == nil && b != nil {
					if err := e.store.DeleteBlock(b.ID); err != nil {
						log.Printf("sync: error removing %s: %v", event.Name, err)
					}
				}
			}
		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("sync: watcher error: %v", err)
		case <-e.done:
			return
		}
	}
}

// Stop terminates the file watcher.
func (e *Engine) Stop() {
	close(e.done)
	if e.watcher != nil {
		e.watcher.Close()
	}
}

// CreateBlock creates a new block, parses natural language dates from the title,
// saves it to disk (versioned), indexes it, and git-commits if configured.
func (e *Engine) CreateBlock(title string) (*block.Block, error) {
	b := block.NewBlock(title)

	// Apply default tags from config
	if e.cfg != nil && len(e.cfg.DefaultTags) > 0 {
		b.Tags = append(e.cfg.DefaultTags, b.Tags...)
	}

	// Extract natural language date from title
	cleanTitle, due, found := nldate.ExtractDate(title, time.Now())
	if found {
		b.Title = cleanTitle
		b.Due = due
		if b.Status == "" {
			b.Status = block.StatusTodo
		}
	}

	// Build filename from (possibly cleaned) title
	safeName := sanitizeFilename(b.Title)
	if safeName == "" {
		safeName = b.ID
	}
	b.FilePath = filepath.Join(e.blocksDir, safeName+".md")
	for i := 1; fileExists(b.FilePath); i++ {
		b.FilePath = filepath.Join(e.blocksDir, fmt.Sprintf("%s-%d.md", safeName, i))
	}

	versionsDir := filepath.Join(e.dataDir, "versions")
	if err := block.WriteFileVersioned(b, versionsDir); err != nil {
		return nil, fmt.Errorf("writing block file: %w", err)
	}

	if err := e.store.UpsertBlock(b); err != nil {
		return nil, fmt.Errorf("indexing block: %w", err)
	}

	e.maybeGitCommit(fmt.Sprintf("add: %s", b.Title))
	return b, nil
}

// UpdateBlock writes a block to disk (versioned), re-indexes it, and git-commits.
func (e *Engine) UpdateBlock(b *block.Block) error {
	versionsDir := filepath.Join(e.dataDir, "versions")
	if err := block.WriteFileVersioned(b, versionsDir); err != nil {
		return fmt.Errorf("writing block: %w", err)
	}
	if err := e.store.UpsertBlock(b); err != nil {
		return err
	}
	e.maybeGitCommit(fmt.Sprintf("update: %s", b.Title))
	return nil
}

// DeleteBlock removes a block file from disk and triggers a git commit.
func (e *Engine) DeleteBlock(b *block.Block) error {
	if b.FilePath != "" {
		if err := os.Remove(b.FilePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing block file: %w", err)
		}
	}
	// Delete synchronously from DB so UI can refresh instantly
	if err := e.store.DeleteBlock(b.ID); err != nil {
		return err
	}
	e.maybeGitCommit(fmt.Sprintf("delete: %s", b.Title))
	return nil
}

// EnsureDailyNote creates today's daily note if it doesn't exist and returns it.
func (e *Engine) EnsureDailyNote() (*block.Block, error) {
	today := time.Now().Format("2006-01-02")
	filename := today + "-daily.md"
	fp := filepath.Join(e.blocksDir, filename)

	if _, err := os.Stat(fp); err == nil {
		return block.ParseFile(fp)
	}

	b := block.NewBlock(fmt.Sprintf("Daily — %s", today))
	b.Tags = []string{"@daily"}
	b.Body = dailyTemplate
	b.FilePath = fp

	versionsDir := filepath.Join(e.dataDir, "versions")
	if err := block.WriteFileVersioned(b, versionsDir); err != nil {
		return nil, err
	}
	if err := e.store.UpsertBlock(b); err != nil {
		return nil, err
	}
	e.maybeGitCommit(fmt.Sprintf("daily: %s", today))
	return b, nil
}

const dailyTemplate = `## 🎯 Top 3 Priorities

1. 
2. 
3. 

## 📝 Notes


## 💭 Reflections

`

// DataDir returns the root data directory.
func (e *Engine) DataDir() string { return e.dataDir }

// BlocksDir returns the blocks directory.
func (e *Engine) BlocksDir() string { return e.blocksDir }

func (e *Engine) maybeGitCommit(message string) {
	if e.cfg == nil || !e.cfg.GitCommit {
		return
	}
	if err := gitpkg.Init(e.dataDir); err != nil {
		log.Printf("git: init error: %v", err)
		return
	}
	if err := gitpkg.Commit(e.dataDir, message); err != nil {
		log.Printf("git: commit error: %v", err)
	}
}

func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	result := b.String()
	result = strings.Trim(result, "-")
	if len(result) > 64 {
		result = result[:64]
	}
	return result
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

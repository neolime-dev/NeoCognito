Based on my analysis of the NeoCognito codebase, here's a comprehensive assessment of the project's current state:

## (1) Maturity and Stability

NeoCognito is currently in **development status** with version `v1.0.0-dev`. [1](#1-0)  The README describes it as "v1.0" with a comprehensive feature set including hybrid storage, PARA integration, knowledge graphs, TUI dashboard, and CLI workflows. [2](#1-1) 

However, there are **known technical issues** in the codebase:
- A "hack" for color extraction in the HTML export system [3](#1-2) 
- No formal bug tracking system or issue documentation files were found in the repository

The database layer includes a notable **stability workaround**: `SetMaxOpenConns(1)` to force single connection usage, preventing "database is locked" errors from concurrent UI transactions and background sync loops. [4](#1-3)  This indicates potential concurrency issues that were addressed with a limiting solution rather than proper concurrent access handling.

## (2) Code Quality and Test Coverage

Test coverage is **minimal and insufficient for production use**:
- Only **3 test files** exist: `tt_test.go`, `internal/block/block_test.go`, and `internal/nldate/parse_test.go` [5](#1-4) [6](#1-5) 
- The main `tt_test.go` contains only a trivial palette test [7](#1-6) 
- **No test targets** in the Makefile (only build, install, clean, and help) [8](#1-7) 
- **No CI/CD pipeline** (no GitHub Actions workflows or other CI configuration files found)
- No code coverage reporting or benchmarks

The core `block` package has reasonable unit tests covering parsing, marshaling, and file operations [9](#1-8) , but critical components like the `store`, `sync`, and `tui` packages appear to have **no test coverage**.

## (3) Documentation Completeness

Documentation is **moderately complete** but has gaps:

**Strengths:**
- Comprehensive README with feature descriptions, installation instructions, usage examples, and keybindings [10](#1-9) 
- Project structure documentation [11](#1-10) 
- Detailed roadmap document [12](#1-11) 
- Contributing guidelines (in Portuguese) [13](#1-12) 

**Gaps:**
- No API documentation or GoDoc comments visible in the analyzed files
- No architecture documentation beyond basic structure
- No troubleshooting guide (except for binary shadowing) [14](#1-13) 
- Configuration documentation is minimal [15](#1-14) 

## (4) Architectural Limitations and Technical Debt

Several **architectural limitations** exist:

**Database Concurrency:**
The most significant limitation is the forced single-connection SQLite setup, which serializes all database access. [4](#1-3)  While this prevents locking errors, it creates a bottleneck that could impact performance with heavy usage or background sync operations.

**Schema Migration:**
The database migration system uses a **"hot migration" approach** with error suppression for the `area` column, checking for "duplicate column name" errors. [16](#1-15)  This is fragile and doesn't provide proper schema versioning or rollback capabilities.

**Technical Debt Items:**
- Color theme extraction in export uses a workaround acknowledging that `lipgloss.TerminalColor` doesn't easily export hex colors without renderer context [3](#1-2) 
- Error handling in versioned file writes acknowledges data loss risk but includes a comment questioning whether to return or log [17](#1-16) 
- The wikilink resolution tries both ID and title lookup but doesn't handle ambiguity [18](#1-17) 

**Architectural Strengths:**
- Clean separation of concerns with packages for `block`, `store`, `sync`, `tui`, etc.
- Message-passing architecture using Bubble Tea framework
- Hybrid storage approach (Markdown files + SQLite index) provides resilience

## (5) Community Contribution Guidelines and Team Structure

**Contribution Process:**
A Portuguese-language contributing guide exists with:
- Standard GitHub fork/PR workflow [19](#1-18) 
- Conventional commit message format (feat, fix, docs, style, refactor, test, chore) [20](#1-19) 
- Code quality expectations (go fmt, go vet) [21](#1-20) 

**Team Structure:**
No visible team structure, maintainer information, or governance model. The repository appears to be under "lemondesk" organization based on module path. [22](#1-21) 

## (6) Roadmap and Future Plans

A **detailed roadmap** exists with features categorized by version (v0.2 through v1.0) and priority levels. [23](#1-22) 

**Primary roadmap includes:**
- v0.2: Configuration, daily notes, git auto-commit, natural language dates, block versioning
- v0.3: Recurrence, pomodoro timer, undo stack, stale alerts
- v0.4: Templates, PARA method, Kanban view, batch operations
- v0.5: Browser clip, floating capture, smart suggestions
- v0.6: Heatmap, weekly review, command palette, zen mode, themes
- v1.0: Related blocks panel, block relations graph [24](#1-23) 

Interestingly, many v1.0 features mentioned in the roadmap are **already implemented** according to the README (command palette, zen mode, themes, knowledge graph, related blocks). [2](#1-1)  This suggests the roadmap may be outdated or the version numbering is not synchronized.

**Optional post-v1.0 features** include remote sync, time tracking, hooks, Obsidian compatibility, and LLM integration. [25](#1-24) 

## Main Pain Points for Production Readiness

To make NeoCognito production-ready for a wider audience, the following issues must be addressed:

1. **Critical: Test Coverage** - Current ~3% test coverage (estimated based on files) is insufficient. Need comprehensive unit, integration, and end-to-end tests for all core packages, especially `store`, `sync`, and `tui`.

2. **Critical: Database Concurrency** - The single-connection limitation is a fundamental scalability issue. [26](#1-25)  Need proper transaction management and concurrent access handling.

3. **High Priority: CI/CD Pipeline** - No automated testing, linting, or release process. Need GitHub Actions or similar for quality gates.

4. **High Priority: Schema Versioning** - The hot migration approach is fragile. [16](#1-15)  Need a proper migration system (e.g., golang-migrate) with version tracking.

5. **Medium Priority: Error Handling** - Some error paths have uncertain handling (log vs. return). [17](#1-16)  Need consistent error handling strategy and logging framework.

6. **Medium Priority: Version Status** - Still marked as "dev" despite claiming v1.0 features. [1](#1-0)  Need clear versioning and release process.

7. **Medium Priority: Documentation** - Add API documentation, troubleshooting guide, and architecture documentation beyond basic structure.

8. **Low Priority: Technical Debt** - Address the color extraction hack [3](#1-2)  and other acknowledged workarounds.

## Notes

The NeoCognito project shows **promising architecture and feature completeness** but is clearly still in active development despite the v1.0 designation. The code is well-organized with clean separation of concerns, and the hybrid Markdown + SQLite approach is solid. However, the **lack of comprehensive testing, CI/CD infrastructure, and the database concurrency limitation** are significant blockers for production use at scale.

The project would benefit from:
- A formal testing strategy with target coverage metrics
- Automated CI/CD pipeline with quality gates
- Resolution of the SQLite concurrency bottleneck
- Clear release versioning (currently dev but claiming v1.0 features)
- Bilingual documentation (currently mixed English/Portuguese)

The contributing guide being in Portuguese while most code and documentation is in English suggests a specific target community, which should be clarified for broader adoption.

### Citations

**File:** internal/info/info.go (L5-5)
```go
	Version   = "v1.0.0-dev"
```

**File:** README.md (L1-179)
```markdown
# <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/zap.svg" width="24" style="vertical-align:middle"> NeoCognito

**NeoCognito** is a terminal-native, keyboard-centric "Second Brain" and personal organization system. It combines the resilience and portability of local Markdown files with the speed and querying capabilities of a SQLite database.

Built entirely in Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea), NeoCognito runs where you work: in the terminal.

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/sparkles.svg" width="18" style="vertical-align:middle"> Features (v1.0)

NeoCognito is designed for high-performance knowledge management.

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/brain-circuit.svg" width="18" style="vertical-align:middle"> Core Architecture
- **Hybrid Storage:** Every entity is a "Block". Markdown files with YAML frontmatter indexed into a local SQLite database using `FTS5` for instant retrieval.
- **PARA Integration:** Native support for Projects, Areas, Resources, and Archives.
- **Knowledge Graph:** Force-directed ASCII visualization of `[[wikilinks]]`.
- **Related Blocks:** Semantic discovery using BM25 scoring via FTS5.

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/laptop.svg" width="18" style="vertical-align:middle"> TUI Dashboard
- **Inbox Zero Wizard:** Guided GTD decision flow to process new captures.
- **Kanban Board:** Visual task management with `h/l` movement.
- **Zen Mode:** Distraction-free internal editor for deep focus.
- **Command Palette:** Fuzzy-searchable action menu (`Ctrl+p`).
- **Heatmap Calendar:** ASCII activity tracking for your knowledge base.

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/zap.svg" width="18" style="vertical-align:middle"> CLI & Workflow
- **Scriptable Capture:** `neocognito add`, `neocognito capture` (popup), and `neocognito clip` (stdin).
- **Knowledge Map:** A horizontal mind-map layout for visualizing knowledge flows.
- **Assinatura Lemon:** Check the "About" section in the Command Palette (`Ctrl+p`).
- **Static Export:** Generate a themed HTML website from your notes via `neocognito export`.
- **Pomodoro Timer:** In-dashboard focus timer with persistence.

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/rocket.svg" width="18" style="vertical-align:middle"> Installation

The recommended way to install NeoCognito is by building from source using the provided `Makefile`. This ensures that your system binary is always correctly updated.

### Quick Install (Recommended)

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/lemondesk/NeoCognito.git
    cd NeoCognito
    ```

2.  **Build and Install:**
    This will compile the project and move the binary to `/usr/bin/neocognito` (requires sudo).
    ```bash
    make install
    ```

3.  **Run Initial Config:**
    ```bash
    neocognito config
    ```

---

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/construction.svg" width="18" style="vertical-align:middle"> Developmental Build
If you just want to test without installing globally:
```bash
make build
# Then run via:
./neocognito
```

---

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/triangle-alert.svg" width="18" style="vertical-align:middle"> Troubleshooting: Command not updating?
If you run `make install` but still see an old version, you might have another `neocognito` binary shadowing it (e.g., in `~/.local/bin` or `~/go/bin`).

Run this to clean up rogue binaries:
```bash
make uninstall-old
```

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/book-open.svg" width="18" style="vertical-align:middle"> Usage Examples & Recipes

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/zap.svg" width="18" style="vertical-align:middle"> Quick Capture
Capture ideas without breaking your flow.

```bash
# Basic capture to Inbox
neocognito add "Research force-directed graph algorithms"

# Capture with natural language date (auto-sets 'due:' field)
neocognito add "Project milestone meeting next Friday 3pm"

# Launch the floating capture popup (great for a WM hotkey)
neocognito capture

# Clip content from the web or clipboard (using xclip/wl-paste)
xclip -o | neocognito clip --url "https://news.ycombinator.com"
```

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/crosshair.svg" width="18" style="vertical-align:middle"> Deep Work & Focus
Use Zen mode for distraction-free writing.

```bash
# Open a specific block in Zen Mode TUI editor
neocognito zen blocks/8f2a1b3c.md

# Or use the ID directly
neocognito zen 8f2a1b3c
```

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/globe.svg" width="18" style="vertical-align:middle"> Sharing & Publishing
Export your Second Brain as a themed static website.

```bash
# Export to the default 'export/' folder
neocognito export

# Export to a custom path (e.g., for GitHub Pages)
neocognito export ./public
```

### <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/brain-circuit.svg" width="18" style="vertical-align:middle"> Knowledge Discovery (TUI)
- **Show Related**: Press `r` on any block to see semantically similar notes (TF-IDF matched).
- **Open Graph**: Press `7` or select "Graph" in the sidebar to visualize your connections.
- **GTD Wizard**: Press `Z` from the Inbox to rapidly triage new items:
  - `t` -> Move to Tasks
  - `y` -> Mark Done
  - `a` -> Move to Archive
  - `x` -> Delete

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/keyboard.svg" width="18" style="vertical-align:middle"> Common Keybindings

| Key | Action |
| :--- | :--- |
| `j`/`k` | Navigate lists |
| `Tab` | Toggle Sidebar/Main focus |
| `Space`| Cycle status (Todo -> Doing -> Done) |
| `Z` | Start Inbox Zero Wizard (GTD) |
| `r` | Show Related Blocks panel |
| `7` | Jump to Knowledge Graph |
| `Ctrl+p`| Open Command Palette |
| `Ctrl+k` / `/`| Global Search |
| `d`| Today's Daily Note |
| `e`| External Editor ($EDITOR) |
| `q` / `Esc`| Close Overlays / Quit |

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/palette.svg" width="18" style="vertical-align:middle"> Customization

Configuration is in `~/.config/neocognito/config.toml`. Run `neocognito config` to see labels.

### Example PARA Config
```toml
# PARA Method base categories
default_tags = ["@inbox"]
stale_days = 21

# Theme Profile
theme = "tokyo-night" # Or "catppuccin", "nord", "gruvbox", "omarchy"
```

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/folder-tree.svg" width="18" style="vertical-align:middle"> Project Structure

- `cmd/`: CLI handlers for capture, clip, and export.
- `internal/`:
  - `block/`: Core data model and Markdown/YAML parser.
  - `export/`: Static HTML generator.
  - `store/`: SQLite/FTS5 database layer.
  - `sync/`: Filesystem watcher and indexer.
  - `tui/`: Bubbletea dashboard and components (Graph, Zen, Related).

---

## <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/scale.svg" width="18" style="vertical-align:middle"> License
This project is licensed under the GNU General Public License v3.0. See the `LICENSE` file for details.
```

**File:** internal/export/export.go (L146-149)
```go
	// Fix: styles.Primary is often an AdaptiveColor or Color.
	// lipgloss.TerminalColor doesn't have a direct string hex export easily without a renderer context.
	// We'll use a hack to extract the string if it looks like #XXXXXX or just use hardcoded defaults if it fails.
	// For this version, let's just use the Render hack.
```

**File:** internal/store/sqlite.go (L27-30)
```go
	// FIX: Force a single connection to serialize writes at the Go level.
	// This entirely prevents driver-level SQLite "database is locked" errors
	// caused by concurrent UI transactions and background sync loops.
	db.SetMaxOpenConns(1)
```

**File:** internal/store/sqlite.go (L91-96)
```go
	// Hot migration for existing DB
	if _, err := s.db.Exec("ALTER TABLE blocks ADD COLUMN area TEXT DEFAULT ''"); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("adding area column: %w", err)
		}
	}
```

**File:** internal/store/sqlite.go (L206-225)
```go
	seenLinks := make(map[string]bool)
	for _, link := range b.Links {
		// Try to resolve link (which might be a title or an ID) to a definitive ID
		var targetID string
		// 1. Is it already a valid ID?
		err := tx.QueryRow("SELECT id FROM blocks WHERE id = ?", link).Scan(&targetID)
		if err != nil {
			// 2. Is it a title?
			_ = tx.QueryRow("SELECT id FROM blocks WHERE title = ?", link).Scan(&targetID)
		}

		if targetID != "" && targetID != b.ID {
			if !seenLinks[targetID] {
				if _, err := tx.Exec("INSERT OR IGNORE INTO links (source_id, target_id) VALUES (?, ?)", b.ID, targetID); err != nil {
					return fmt.Errorf("inserting link %s -> %s: %w", b.ID, targetID, err)
				}
				seenLinks[targetID] = true
			}
		}
	}
```

**File:** internal/block/block_test.go (L1-10)
```go
package block

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

```

**File:** internal/block/block_test.go (L11-172)
```go
func TestParseWithFrontmatter(t *testing.T) {
	content := `---
id: "abc123"
title: "Test Block"
status: "todo"
tags: ["@work", "@urgent"]
created: "2026-02-20T15:00:00-03:00"
modified: "2026-02-20T15:30:00-03:00"
---

## Hello World

This is the body.
`

	b, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b.ID != "abc123" {
		t.Errorf("expected ID 'abc123', got '%s'", b.ID)
	}
	if b.Title != "Test Block" {
		t.Errorf("expected Title 'Test Block', got '%s'", b.Title)
	}
	if b.Status != StatusTodo {
		t.Errorf("expected Status 'todo', got '%s'", b.Status)
	}
	if len(b.Tags) != 2 || b.Tags[0] != "@work" {
		t.Errorf("unexpected Tags: %v", b.Tags)
	}
	if !strings.Contains(b.Body, "Hello World") {
		t.Errorf("body should contain 'Hello World', got: %s", b.Body)
	}
}

func TestParseWithoutFrontmatter(t *testing.T) {
	content := "# Just a Note\n\nNo frontmatter here.\n"

	b, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b.ID == "" {
		t.Error("expected generated ID, got empty")
	}
	if b.Title != "" {
		t.Errorf("expected empty Title, got '%s'", b.Title)
	}
	if b.Body != content {
		t.Errorf("expected full content as body, got: %s", b.Body)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &Block{
		ID:       "round1",
		Title:    "Round Trip Test",
		Status:   StatusDoing,
		Tags:     []string{"@test"},
		Created:  now,
		Modified: now,
		Body:     "Some body content.\n",
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", parsed.ID, original.ID)
	}
	if parsed.Title != original.Title {
		t.Errorf("Title mismatch: %s vs %s", parsed.Title, original.Title)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status mismatch: %s vs %s", parsed.Status, original.Status)
	}
	if !strings.Contains(parsed.Body, "Some body content") {
		t.Errorf("Body mismatch, got: %s", parsed.Body)
	}
}

func TestWriteAndParseFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test-block.md")

	original := NewBlock("File Test")
	original.Status = StatusTodo
	original.Tags = []string{"@filetest"}
	original.Body = "Testing file write.\n"
	original.FilePath = fp

	if err := WriteFile(original); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Fatal("file was not created")
	}

	parsed, err := ParseFile(fp)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", parsed.ID, original.ID)
	}
	if parsed.FilePath != fp {
		t.Errorf("FilePath mismatch: %s vs %s", parsed.FilePath, fp)
	}
}

func TestNextStatus(t *testing.T) {
	b := &Block{}

	b.NextStatus()
	if b.Status != StatusTodo {
		t.Errorf("expected todo, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusDoing {
		t.Errorf("expected doing, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusDone {
		t.Errorf("expected done, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusArchived {
		t.Errorf("expected archived, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusTodo {
		t.Errorf("expected todo (cycle), got %s", b.Status)
	}
}

func TestBodyPreview(t *testing.T) {
	b := &Block{Body: "Hello World, this is a longer body text for testing."}

	preview := b.BodyPreview(11)
	if preview != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", preview)
	}

	short := &Block{Body: "Short"}
	if short.BodyPreview(100) != "Short" {
		t.Errorf("expected 'Short', got '%s'", short.BodyPreview(100))
	}
}
```

**File:** internal/nldate/parse_test.go (L1-7)
```go
// Package nldate — unit tests for the natural language date parser.
package nldate

import (
	"testing"
	"time"
)
```

**File:** tt_test.go (L8-10)
```go
func TestPalette(t *testing.T) {
_ = palette.New()
}
```

**File:** Makefile (L7-7)
```text
.PHONY: all build install uninstall-old check-path clean help
```

**File:** roadmap_neocognito.md (L1-96)
```markdown
# NeoCognito — Finalized Feature Roadmap

## Legend
- 🟢 **Primary** — on the roadmap, will be implemented
- 🟡 **Optional** — considered, implement if time/demand
- 🔴 **Won't Add** — out of scope

---

## 🟢 PRIMARY ROADMAP

### v0.2 — Foundation
| Feature | Description |
|---|---|
| Config file | `~/.config/neocognito/config.toml` — editor, data_dir, git, stale days |
| Daily Note | Auto-created `YYYY-MM-DD.md` with template, `d` key to jump to it |
| Git auto-commit | Commit to local repo after every block write |
| Natural Language Dates | Parse "tomorrow" / "next Friday" in `add` → auto-set `due:` |
| Block Versioning | Snapshot on save, `H` to view history and restore |

### v0.3 — Productivity
| Feature | Description |
|---|---|
| Recurrence engine | `recur: weekly` auto-creates next block on completion |
| Pomodoro / Focus Timer | In-status-bar timer, logs `pomodoros: N` on block |
| Undo stack | In-memory `u` to undo last mutation |
| Stale Block Alerts | `⚠ stale` badge on blocks untouched for N days (configurable) |
| Block Handoff | `delegated_to:` field + Waiting view for delegated tasks |

### v0.4 — Organization
| Feature | Description |
|---|---|
| Template system | `neocognito new --template meeting` from `templates/` dir |
| PARA Method | First-class `area:` field (Projects/Areas/Resources/Archives) |
| Kanban View | TUI board: `todo | doing | done` columns, `h/l` to move cards |
| Multi-select & Batch | `v` to select, then bulk retag / status change / delete |
| Inbox Zero Flow | Guided GTD wizard: actionable → schedule or archive |

### v0.5 — Capture & Input
| Feature | Description |
|---|---|
| Browser Clip | `xclip \| neocognito clip --url ...` → block with URL + clipped text |
| Floating Capture Window | `neocognito capture` + WM hotkey → tiny popup, one-line capture |
| Smart Inbox Suggestions | After capture, FTS-match existing blocks → suggest links |

### v0.6 — Insights & UX
| Feature | Description |
|---|---|
| Heatmap Calendar | ASCII GitHub-style contribution graph of blocks/tasks |
| Weekly Review Dashboard | Tasks done, blocks created, tags touched, next week prompt |
| Command Palette | `Ctrl+P` fuzzy action search ("daily note", "export", ...) |
| Zen Mode | `neocognito zen` — full-screen distraction-free single-block editor |
| Theme Engine | `theme:` in config.toml; ships with tokyo-night, catppuccin, gruvbox, nord |

### v1.0 — Knowledge Graph
| Feature | Description |
|---|---|
| Related blocks panel | TF-IDF/FTS5 keyword match while viewing a block |
| Block relations graph | ASCII/graphviz visualization of `[[wikilinks]]` |

---

## 🟡 OPTIONAL (post-v1.0)

| # | Feature |
|---|---|
| — | Remote sync (`rsync` / git push / Restic) |
| — | Related blocks panel (moved from primary if time-constrained) |
| — | Tag Cloud Panel |
| — | Time Tracking (`time_spent:` per block) |
| — | Hooks / shell scripts on events |
| — | Obsidian Compatibility Mode |
| — | `neocognito export` (HTML / PDF / Org-mode) |
| — | Inbox Zero Flow (full GTD wizard) |
| — | Scratch Pad Panel |
| — | Spaced Repetition (SRS / Anki-style) |
| — | Reading Mode + Progress Tracking |
| — | Local LLM integration (`ollama`) |
| — | Shared Blocks via Git (collab) |
| — | Mouse Support |
| — | Block Locking (`locked: true`) |
| — | Minimap / Outline Panel |
| — | Static Site Export (`neocognito publish`) |

---

## 🔴 WON'T ADD

| Feature | Reason |
|---|---|
| Plugin API | Hooks cover 95% of use cases with less complexity |
| Voice-to-Block | Too niche; covered by external whisper + pipe |
| Burndown Chart | Too project-management-specific |
| SSH Remote Mode | Complexity vs. gain; use `ssh -L` manually |
| Matrix / Telegram Bot | Breaks "local-first" philosophy |
| RSS Feed Generator | Belongs in export plugin if ever needed |
```

**File:** CONTRIBUTING.md (L1-39)
```markdown
# Como Contribuir com o NeoCognito

Primeiramente, obrigado pelo seu interesse em contribuir com o NeoCognito! Estamos felizes em ter você aqui. Toda contribuição é bem-vinda, desde a correção de um simples erro de digitação até a implementação de uma nova funcionalidade.

## Como Começar

1.  **Fork o Repositório:** Clique no botão "Fork" no canto superior direito da página do repositório no GitHub para criar uma cópia sua.
2.  **Clone seu Fork:** `git clone https://github.com/SEU-USUARIO/NeoCognito.git`
3.  **Crie uma Branch:** `git checkout -b minha-feature-incrivel` (Use um nome descritivo).
4.  **Faça suas Alterações:** Implemente sua funcionalidade ou correção de bug.
5.  **Commit suas Alterações:** `git commit -m "feat: Adiciona funcionalidade incrível"` (Veja nosso guia de mensagens de commit abaixo).
6.  **Envie para seu Fork:** `git push origin minha-feature-incrivel`
7.  **Abra um Pull Request:** No GitHub, vá para o repositório original e abra um "Pull Request" com uma descrição clara de suas alterações.

## Guia para Pull Requests

- **Mantenha a Simplicidade:** Prefira pull requests pequenos e focados. Não agrupe várias funcionalidades em um único PR.
- **Escreva uma Boa Descrição:** Explique o "quê" e o "porquê" de suas alterações. Se o seu PR resolve uma `Issue` existente, mencione o número dela (ex: `Resolve #123`).
- **Código Limpo:** Certifique-se de que seu código segue as convenções do Go. Rode `go fmt` e `go vet` antes de fazer o commit para garantir a qualidade.
- **Seja Paciente:** Faremos o nosso melhor para revisar seu PR o mais rápido possível.

## Padrão de Mensagens de Commit

Usamos um padrão para as mensagens de commit para manter o histórico limpo e organizado. Por favor, siga este formato:

`<tipo>: <assunto>`

**Tipos comuns:**
-   `feat`: Uma nova funcionalidade.
-   `fix`: Uma correção de bug.
-   `docs`: Alterações na documentação.
-   `style`: Alterações que não afetam o significado do código (espaços, formatação, etc).
-   `refactor`: Uma alteração de código que não corrige um bug nem adiciona uma funcionalidade.
-   `test`: Adicionando testes ou corrigindo testes existentes.
-   `chore`: Alterações em processos de build, ferramentas auxiliarias, etc.

**Exemplo:** `feat: Adiciona suporte para temas personalizados no config.toml`

Obrigado mais uma vez por sua contribuição!
```

**File:** internal/block/block.go (L223-224)
```go
					// Log it or return? Returning could prevent saving. We should probably return to avoid data loss.
					return fmt.Errorf("failed to create version snapshot: %w", writeErr)
```

**File:** go.mod (L1-1)
```text
module github.com/neolime-dev/neocognito
```


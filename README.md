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
    git clone https://github.com/neolime-dev/NeoCognito.git
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

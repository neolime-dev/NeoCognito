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

# Changelog

All notable changes to NeoCognito will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nerd Font icon migration — all emoji replaced with `nf-md-*` glyphs across every TUI panel for consistent rendering in any Nerd Font terminal.
- Home dashboard enhancements: mini-calendar, daily focus list, upcoming tasks, recent notes, session stats, and pomodoro bar all visible at a glance.
- Session-scoped pomodoro counter wired from the timer to the home dashboard.

### Changed
- `sync.Engine.maybeGitCommit` now runs in a background goroutine, eliminating the lag spike when closing the daily note editor.
- Removed redundant `IndexFile` call from `openDailyNote`.

### Fixed
- `NormalModeStyle` (vim mode pill) was incorrectly used for the date header and inbox stat count on the home dashboard; replaced with `PrimaryStyle` / `SelectedItemStyle`.
- `%-10s` format verb applied to ANSI-rendered strings in home Quick Actions and GTD options caused visual misalignment; fixed by padding the raw string before styling.
- GTD "Is this actionable?" prompt now uses `AccentStyle` instead of the vim-mode pill style.
- Kanban empty-column state now reads "No tasks yet" instead of the bare "Empty".
- Removed dead `borderColor` helper and unused `lipgloss` import from `sidebar`.
- Removed duplicate `TextStyle` assignment in `inbox`.

## [1.1.0] - 2026-02-26

### Added
- Structured logging with `log/slog` throughout the sync engine (replaces `log.Printf`).
- `SetLogger` method on `sync.Engine` for test-time log suppression.
- `styles.ColorHex` / `ThemeColors.CSSTheme()` / `styles.CurrentTheme()` — clean API for converting Lipgloss theme colors to CSS hex strings.
- `store.New` GoDoc comment.
- GitHub Actions CI workflow (build + vet + race-detector tests on every push/PR).
- GoReleaser release pipeline — publishes binaries for linux/darwin × amd64/arm64 on `v*` tags.
- `Makefile` `test`, `lint`, `tag`, and `release` targets.
- Store integration test suite (14 tests).
- Config package test suite (12 tests).
- Recur package test suite (13 tests).
- Sync engine test suite (9 tests, in-memory stub store).
- Export package test suite (6 tests).
- `CONTRIBUTING.pt-BR.md` — Portuguese translation of the contribution guide.
- `CHANGELOG.md` — this file.

### Changed
- Module path corrected to `github.com/neolime-dev/neocognito`.
- `refreshPanelData` in the TUI now issues a single `ListBlocks` query and partitions results in-memory (was 3 separate DB queries).
- `export.Run` now propagates template-parse and template-execute errors instead of silencing them.
- `store.DeleteBlock` wraps the underlying error for better diagnostics.
- SQLite connection pool tuned: `SetMaxOpenConns(4)`, `SetMaxIdleConns(4)`, `SetConnMaxLifetime(0)` (was `SetMaxOpenConns(1)`).
- `nldate.ExtractDate` probe order fixed — three-word phrases are tried before two-word phrases, preventing greedy mis-match (e.g. "Buy milk in 3 days" no longer consumed as "3 days").
- `CONTRIBUTING.md` translated to English; Portuguese original preserved as `CONTRIBUTING.pt-BR.md`.

### Fixed
- Markdown visualization rendering.

## [1.0.0] - 2025-01-01

### Added
- Initial release of NeoCognito — a terminal-based second brain / PKM tool.
- Bubble Tea TUI with panels: Inbox, Timeline, Kanban, Home, Search, Daily note, Wiki, Projects, Waiting, Journal, Graph, Tag cloud, Pomodoro, GTD, Heatmap, Zen mode, Related blocks, History, Review.
- SQLite-backed block store with FTS5 full-text search.
- Markdown file sync with `fsnotify` debounced watcher.
- Natural-language date parsing (`nldate`).
- Recurrence rules (`recur`).
- Block versioning.
- Git auto-commit on block create/update/delete.
- Static HTML site export.
- Themes: Tokyo Night, Catppuccin, Nord, Gruvbox, Omarchy/system.
- `config.toml` with XDG Base Directory support.

[Unreleased]: https://github.com/neolime-dev/neocognito/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/neolime-dev/neocognito/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/neolime-dev/neocognito/releases/tag/v1.0.0

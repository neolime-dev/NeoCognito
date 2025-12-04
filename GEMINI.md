# User Environment Context

## Primary Project: NeoCognito (PKM System)

**Location:** `~/Dev_Pro/NeoCognito`
**Data Vault:** `~/Vault`

### Overview
NeoCognito is a custom Personal Knowledge Management system tailored for data scientists and ADHD/ASD workflows, focusing on **low-latency capture** and **object permanence**. It adheres to the Unix Philosophy, utilizing small, modular scripts connected by plain text (Markdown).

### Key Components

1.  **Capture Layer (Write-Only)**
    *   **Script:** `capture.sh`
    *   **Function:** Instant capture of "Fleeting Notes" into `00_Inbox.md`. Uses `walker` for a floating input bar.
    *   **Format:** Adds timestamped bullets (no checkboxes) to reduce anxiety.

2.  **Persistence Layer (Read-Only)**
    *   **Script:** `launch_wall.sh`
    *   **Function:** Renders `00_Inbox.md` and `TODO_Today.md` directly onto the desktop wallpaper using `Conky`.
    *   **Hardware Aware:** Automatically detects monitor count (via `hyprctl` or `xrandr`) to load either `neocognito_dual.conf` or `neocognito_single.conf`.
    *   **Rendering:** Supports bold text rendering via `render_bold.sh`.

3.  **Action Layer (Interactive)**
    *   **Script:** `mark_task.sh`
    *   **Function:** "Task Killer". Opens a menu (Walker) listing only pending tasks. Selecting one marks it as `[x]` in the file without opening a text editor.

4.  **Processing Layer (Batch)**
    *   **Script:** `daily_review.sh`
    *   **Function:** Morning protocol. Archives completed tasks to `99_Archive/Journal/YYYY-MM-DD.md`, keeps pending tasks in `TODO_Today.md`, and opens Obsidian for organization.

### Directory Map

*   **`~/Dev_Pro/NeoCognito/`**: Source code repository.
    *   `scripts/`: Operational scripts (symlinked to `~/.local/bin`).
    *   `config/`: Conky configuration files.
    *   `setup.sh`: Installer for new machines.
*   **`~/Vault/`**: Private data repository (Markdown).
    *   `00_Inbox.md`: Raw capture stream.
    *   `10_Planning/TODO_Today.md`: Active tasks.
    *   `20_Zettels/`: Permanent knowledge notes.
    *   `99_Archive/`: Journal and attachments.

### Usage Commands

These scripts are symlinked to `~/.local/bin` and are available in the path:

*   `capture.sh`: Trigger via hotkey (e.g., `Super+Shift+Z`).
*   `mark_task.sh`: Trigger via hotkey (e.g., `Super+Alt+Z`).
*   `daily_review.sh`: Run at start of day (e.g., `Super+Shift+Alt+Z`).
*   `launch_wall.sh`: Runs automatically on login (`exec-once`) or manually to refresh.
*   `autosave.sh`: Background Git backup.

## System Context

*   **OS:** Arch-based Linux (Omarchy).
*   **Window Manager:** Likely Hyprland (scripts check for `hyprctl`).
*   **Shell:** Bash.
*   **Dependencies:** `walker`, `conky`, `git`, `obsidian`, `perl`.

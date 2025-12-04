# NeoCognito 🧠

**Personal Knowledge Management (PKM) Ecosystem** designed for me in Linux (Arch-based) environment, focusing on **Zettelkasten Philosophy** and **GTD method**.

---

## 🏗 Architecture & Philosophy

The system follows a hybrid **Zettelkasten + GTD** approach, prioritizing capture speed and visibility of pending tasks.

1.  **Capture Layer (Write-Only):** `capture.sh`
    *   *Goal:* Low latency. Capture *Fleeting Notes* (quick thoughts) without friction.
    *   *Tool:* Walker (Floating input).
2.  **Action Layer (Interactive):** `mark_task.sh`
    *   *Goal:* Quick "Task Killing" without opening Obsidian.
    *   *Tool:* Walker (Selection menu).
3.  **Persistence Layer (Read-Only):** `launch_wall.sh` (Conky)
    *   *Goal:* Object permanence. Displays Inbox and TODO on the desktop.
    *   *Rendering:* Supports **Bold** (`**text**`) for visual emphasis.
4.  **Processing Layer (Batch):** Obsidian + `daily_review.sh`
    *   *Goal:* Refinement, Archiving, and Planning.

---

## 📂 Directory Structure (Vault)

```text
~/Vault/
├── 00_Inbox.md           # Raw data input (Bullets with Timestamp)
├── 00_Inbox/             # Folder for new notes created via Obsidian
├── 10_Planning/
│   └── TODO_Today.md     # Day's focus (Rendered on Desktop)
├── 20_Zettels/           # Atomic notes and permanent knowledge
├── 30_Projects/          # Ongoing projects
└── 99_Archive/           # Archive and Journal
    ├── Attachments/      # Images/PDFs
    └── Journal/          # Daily history (Completed tasks are moved here)
```

---

## 🚀 Installation

### 1. Prerequisites (Arch Linux)
```bash
sudo pacman -S conky git obsidian perl
yay -S walker-bin
```

### 2. Setup (Codebase)
The scripts are located in `~/NeoCognito/scripts`. Create the symbolic links:

```bash
ln -sf ~/NeoCognito/scripts/capture.sh ~/.local/bin/capture.sh
ln -sf ~/NeoCognito/scripts/autosave.sh ~/.local/bin/autosave.sh
ln -sf ~/NeoCognito/scripts/launch_wall.sh ~/.local/bin/launch_wall.sh
ln -sf ~/NeoCognito/scripts/daily_review.sh ~/.local/bin/daily_review.sh
ln -sf ~/NeoCognito/scripts/mark_task.sh ~/.local/bin/mark_task.sh
```

*Note: `render_bold.sh` is used internally by Conky and does not require a symlink in `~/.local/bin`.*

---

## ⌨️ Usage Guide

### ⚡ 1. Instant Capture (`capture.sh`)
**Suggested Hotkey:** `Super + Shift + Z`
*   Captures an idea as a *bullet point* in the Inbox.
*   Example: `- **14:00** - Raw Idea`
*   *Does not create a checkbox* (reduces "yet another task" anxiety).

### ✅ 2. Complete Tasks (`mark_task.sh`)
**Suggested Hotkey:** `Super + Alt + Z`
*   Opens a floating menu listing only your pending tasks (`- [ ]`).
*   Upon selection, marks it as done (`- [x]`) in the file.
*   The *The Wall* automatically updates shortly after.

### 🖥️ 3. The Wall (`launch_wall.sh`)
**Suggested Hotkey:** `exec-once` in your Hyprland/i3 config.
*   Displays `00_Inbox.md` (Recent ideas) and `TODO_Today.md` (Day's Focus).
*   Supports **Bold** text (`**text**`) rendering.

### 🔄 4. Daily Review (`daily_review.sh`)
**Suggested Hotkey:** `Super + Shift + Alt + Z` (Run at the start of the day).
1.  **Checks:** If already run today, it just opens Obsidian.
2.  **Migrates:**
    *   Completed tasks `[x]` -> Moved to `99_Archive/Journal/YYYY-MM-DD.md`.
    *   Pending tasks `[ ]` -> **Retained** in `TODO_Today.md`.
3.  **Opens:** Launches Obsidian focused on the Vault for you to process the Inbox (turn *Ideas* into *Tasks* or *Zettels*).

---

## ⚙️ Maintenance & Advanced

### Automatic Backup (`autosave.sh`)
The script performs `git add/commit` locally. Can be scheduled via cron or Systemd Timer.

### Visual Customization
Edit `~/NeoCognito/config/conky/neocognito.conf` to change colors, fonts, or monitor (`xinerama_head`).

---

## 🌍 Distributed Setup (Desktop & Laptop)

To use NeoCognito across multiple machines, we will implement an "Infrastructure as Code" approach:

1.  **Codebase (`~/NeoCognito`):**
    *   Upload `~/NeoCognito` to a **Git Remote** (e.g., GitHub, GitLab). This repository contains all scripts and configurations.
    *   On your other machines, `git clone` this repository.
    *   A `setup.sh` script will be created to automate the installation of dependencies and the creation of symbolic links.

2.  **Data (`~/Vault`):**
    *   Your `~/Vault` will be managed in a **private Git Repository**.
    *   This ensures your personal notes remain secure and private.
    *   The `autosave.sh` script will be enhanced to `git pull --rebase` before committing and `git push` after, ensuring your notes are synchronized across devices.
    *   The `daily_review.sh` will also include a `git pull --rebase` at startup to fetch the latest changes.

3.  **Hardware Adaptability:**
    *   The `launch_wall.sh` script will be refactored to detect the environment (e.g., `hostname`, number of monitors).
    *   It will then dynamically load the appropriate Conky configuration (`neocognito_desktop.conf`, `neocognito_laptop.conf`) tailored for each machine's display setup.

---
*Developed for high cognitive performance and thinking on how my brain works, use for inspiration to create your method.*

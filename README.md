# NeoCognito 🧠

**Personal Knowledge Management (PKM) Ecosystem** designed for Data Scientists and developers in Linux (Arch-based) environments, focusing on **reducing cognitive latency** and **object permanence** (ADHD/ASD Friendly).

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
4.  **Mobile Link (Remote):** Telegram Bot
    *   *Goal:* Capture ideas from anywhere (Phone/Watch) directly to Inbox.
    *   *Tool:* Python Telegram Bot (Self-hosted).

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
sudo pacman -S conky git obsidian perl python
yay -S walker-bin
```

### 2. Setup (Codebase)
Clone the repository to `~/Dev_Pro/NeoCognito` and run the installer:

```bash
git clone https://github.com/YOUR_USER/NeoCognito.git ~/Dev_Pro/NeoCognito
cd ~/Dev_Pro/NeoCognito
./setup.sh
```

This script will:
*   Create the Vault structure.
*   Link all scripts to `~/.local/bin`.
*   Install Python dependencies for the Bot.
*   Link the systemd service.

---

## 🤖 Mobile Bot Setup

1.  **Get Credentials:**
    *   Talk to `@BotFather` on Telegram -> Create Bot -> Get **Token**.
    *   Talk to `@userinfobot` -> Get your **Numeric ID**.

2.  **Configure Secrets:**
    Edit `~/Dev_Pro/NeoCognito/.env`:
    ```env
    TELEGRAM_BOT_TOKEN=your_token_here
    ALLOWED_USER_ID=your_id_here
    ```

3.  **Activate Service:**
    ```bash
    systemctl --user enable --now neocognito-bot.service
    ```

---

## ⌨️ Usage Guide

### ⚡ 1. Instant Capture (`capture.sh`)
**Suggested Hotkey:** `Super + Shift + Z`
*   Captures an idea as a *bullet point* in the Inbox.
*   Example: `- **14:00** - Raw Idea`

### ✅ 2. Complete Tasks (`mark_task.sh`)
**Suggested Hotkey:** `Super + Alt + Z`
*   Opens a floating menu listing only your pending tasks.
*   Marks as done (`[x]`) and updates the Wall.

### 🖥️ 3. The Wall (`launch_wall.sh`)
**Suggested Hotkey:** `exec-once` in WM config.
*   Displays `00_Inbox.md` and `TODO_Today.md`.
*   **Hardware Aware:** Detects dual/single monitor setup automatically.

### 🔄 4. Daily Review (`daily_review.sh`)
**Suggested Hotkey:** `Super + Shift + Alt + Z` (Start of day).
1.  **Syncs:** Pulls latest data from Remote Vault.
2.  **Migrates:** Archives completed tasks, keeps pending ones.
3.  **Opens:** Launches Obsidian.

---

## 🌍 Distributed Setup

To use NeoCognito across multiple machines:

1.  **Codebase:** Sync this repo (`~/Dev_Pro/NeoCognito`) via GitHub public/private repo.
2.  **Data:** Sync `~/Vault` via a **PRIVATE** Git repository.
    *   Scripts (`autosave.sh`, `daily_review.sh`) automatically handle `git pull --rebase` and `git push`.

---
*Developed for high cognitive performance and thinking on how my brain works, use for inspiration to create your own method.*

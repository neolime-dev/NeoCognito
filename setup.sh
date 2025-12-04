#!/bin/bash

echo "🧠 NeoCognito Installer"
echo "-----------------------"

# 1. Create Directories
echo "[*] Creating Vault structure..."
mkdir -p ~/Vault/{00_Inbox,10_Planning,20_Zettels,30_Projects,99_Archive/Journal,99_Archive/Attachments}
touch ~/Vault/00_Inbox.md
touch ~/Vault/10_Planning/TODO_Today.md

# 2. Create Config Directories
mkdir -p ~/.config/conky
mkdir -p ~/.local/bin

# 3. Dependencies Check (Arch Linux)
echo "[*] Checking dependencies..."
MISSING=""

if ! command -v conky &> /dev/null; then MISSING="$MISSING conky"; fi
if ! command -v walker &> /dev/null; then MISSING="$MISSING walker"; fi
if ! command -v git &> /dev/null; then MISSING="$MISSING git"; fi
if ! command -v obsidian &> /dev/null; then MISSING="$MISSING obsidian"; fi

if [ -n "$MISSING" ]; then
    echo "[!] Missing packages:$MISSING"
    echo "    Please install them via pacman/yay."
else
    echo "[OK] All dependencies found."
fi

# 4. Symlink Scripts
echo "[*] Linking scripts..."
SCRIPT_DIR="$HOME/Dev_Pro/NeoCognito/scripts"
BIN_DIR="$HOME/.local/bin"

# Function to safely link
link_script() {
    local name=$1
    ln -sf "$SCRIPT_DIR/$name" "$BIN_DIR/$name"
    echo "    -> Linked $name"
}

link_script "capture.sh"
link_script "autosave.sh"
link_script "launch_wall.sh"
link_script "daily_review.sh"
link_script "mark_task.sh"

echo "-----------------------"
echo "✅ Installation Complete!"
echo "Next steps:"
echo "1. Configure your Window Manager hotkeys (see README)."
echo "2. Run 'launch_wall.sh' to start the widget."

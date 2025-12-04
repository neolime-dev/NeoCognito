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
mkdir -p ~/.config/systemd/user

# 3. Dependencies Check (System)
echo "[*] Checking system dependencies..."
MISSING=""

if ! command -v conky &> /dev/null; then MISSING="$MISSING conky"; fi
if ! command -v walker &> /dev/null; then MISSING="$MISSING walker"; fi
if ! command -v git &> /dev/null; then MISSING="$MISSING git"; fi
if ! command -v obsidian &> /dev/null; then MISSING="$MISSING obsidian"; fi

if [ -n "$MISSING" ]; then
    echo "[!] Missing packages:$MISSING"
    echo "    Please install them via pacman/yay."
else
    echo "[OK] System dependencies found."
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

# 5. Telegram Bot Setup (Optional)
echo "[*] Setting up Telegram Bot..."

if command -v python3 &> /dev/null; then
    # Install Python deps
    echo "    -> Installing Python requirements..."
    # Check if uv is installed for speed, else pip
    if command -v uv &> /dev/null; then
        uv pip install -r "$HOME/Dev_Pro/NeoCognito/requirements.txt" --system
    else
        pip install -r "$HOME/Dev_Pro/NeoCognito/requirements.txt" --break-system-packages
    fi

    # Link Service
    SERVICE_SRC="$HOME/Dev_Pro/NeoCognito/config/systemd/neocognito-bot.service"
    SERVICE_DEST="$HOME/.config/systemd/user/neocognito-bot.service"
    
    if [ -f "$SERVICE_SRC" ]; then
        ln -sf "$SERVICE_SRC" "$SERVICE_DEST"
        echo "    -> Linked systemd service"
        
        # Reload daemon
        systemctl --user daemon-reload
        echo "    -> Systemd reloaded. (Enable service manually after configuring .env)"
    fi
else
    echo "[!] Python3 not found. Skipping Bot setup."
fi

# 6. Env File Warning
ENV_FILE="$HOME/Dev_Pro/NeoCognito/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "IMPORTANT: Create .env file from template to use the Bot!"
    # Optional: Copy template if we had one
fi

echo "-----------------------"
echo "✅ Installation Complete!"
echo "Next steps:"
echo "1. Configure hotkeys (see README)."
echo "2. Run 'launch_wall.sh'."
echo "3. Configure '.env' and run 'systemctl --user enable --now neocognito-bot' for the bot."
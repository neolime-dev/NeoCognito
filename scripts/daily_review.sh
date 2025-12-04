#!/bin/bash

# Configuration
VAULT_DIR="$HOME/Vault"
TODO_FILE="$VAULT_DIR/10_Planning/TODO_Today.md"
JOURNAL_DIR="$VAULT_DIR/99_Archive/Journal"
DATE_STR=$(date +%Y-%m-%d)
JOURNAL_FILE="$JOURNAL_DIR/${DATE_STR}.md"

# Create directory if it doesn't exist
mkdir -p "$JOURNAL_DIR"

cd "$VAULT_DIR" || exit # Change to Vault directory for Git operations

# 0. Sync Vault before starting the review process
notify-send "NeoCognito" "Syncing Vault with remote..." -i folder-sync
git pull --rebase origin master 2>/dev/null # Pull latest changes

# 1. Idempotency Check
if [ -f "$JOURNAL_FILE" ]; then
    notify-send "NeoCognito" "Daily review already done. Opening Vault..." -i dialog-information
    if command -v obsidian &> /dev/null; then
        obsidian --vault "$VAULT_DIR" &
    fi
    exit 0
fi

# 2. Smart Migration
if [ -f "$TODO_FILE" ]; then
    echo "# Daily Log: $DATE_STR" > "$JOURNAL_FILE"
    echo "## Completed Tasks" >> "$JOURNAL_FILE"
    
    grep "\- \[x\]" "$TODO_FILE" >> "$JOURNAL_FILE"
    grep "\- \[ \]" "$TODO_FILE" > "${TODO_FILE}.tmp"
    
    echo "# Focus Today ($DATE_STR)" > "$TODO_FILE"
    cat "${TODO_FILE}.tmp" >> "$TODO_FILE"
    rm "${TODO_FILE}.tmp"
    
    notify-send "NeoCognito" "Completed tasks archived. Pending tasks retained." -i mail-send
else
    echo "# Focus Today ($DATE_STR)" > "$TODO_FILE"
fi

# 3. Open Obsidian for planning
if command -v obsidian &> /dev/null; then
    obsidian --vault "$VAULT_DIR" &
else
    notify-send "NeoCognito" "Open Obsidian to plan your day."
fi

# Optional: Push changes after review is done (if any were made during review)
# git push origin master 2>/dev/null
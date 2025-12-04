#!/bin/bash

# Configuration
VAULT_DIR="$HOME/Vault"
TODO_FILE="$VAULT_DIR/10_Planning/TODO_Today.md" # Renamed
JOURNAL_DIR="$VAULT_DIR/99_Archive/Journal"
DATE_STR=$(date +%Y-%m-%d)
JOURNAL_FILE="$JOURNAL_DIR/${DATE_STR}.md"

# Create directory if it doesn't exist
mkdir -p "$JOURNAL_DIR"

# 1. Idempotency Check
# If today's Journal file already exists, assume review has been done.
if [ -f "$JOURNAL_FILE" ]; then
    notify-send "NeoCognito" "Daily review already done. Opening Vault..." -i dialog-information
    if command -v obsidian &> /dev/null; then
        obsidian --vault "$VAULT_DIR" &
    fi
    exit 0
fi

# 2. Smart Migration
# If TODO file exists...
if [ -f "$TODO_FILE" ]; then
    echo "# Daily Log: $DATE_STR" > "$JOURNAL_FILE" # Translated
    echo "## Completed Tasks" >> "$JOURNAL_FILE"    # Translated
    
    # A. Move COMPLETED tasks [x] to Journal
    grep "\- \[x\]" "$TODO_FILE" >> "$JOURNAL_FILE"
    
    # B. Keep PENDING tasks [ ] in a temporary file
    grep "\- \[ \]" "$TODO_FILE" > "${TODO_FILE}.tmp"
    
    # C. Recreate TODO_Today keeping pending tasks
    echo "# Focus Today ($DATE_STR)" > "$TODO_FILE" # Translated
    cat "${TODO_FILE}.tmp" >> "$TODO_FILE"
    rm "${TODO_FILE}.tmp"
    
    notify-send "NeoCognito" "Completed tasks archived. Pending tasks retained." -i mail-send # Translated
else
    # If TODO file doesn't exist (first use)
    echo "# Focus Today ($DATE_STR)" > "$TODO_FILE" # Translated
fi

# 3. Open Obsidian for planning
if command -v obsidian &> /dev/null; then
    obsidian --vault "$VAULT_DIR" &
else
    notify-send "NeoCognito" "Open Obsidian to plan your day." # Translated
fi

#!/bin/bash

# Configuration
VAULT_DIR="$HOME/Vault"
TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")

cd "$VAULT_DIR" || exit

# 1. Pull latest changes (rebase to keep history clean)
# This prevents conflicts if you made changes on another machine
git pull --rebase origin master 2>/dev/null

# Check for changes (local or pulled)
if [[ -n $(git status -s) ]]; then
    git add .
    git commit -m "Auto-save: $TIMESTAMP"
    
    # 2. Push changes to remote
    git push origin master 2>/dev/null
    
    notify-send "NeoCognito" "Vault synced: changes saved and pushed." -i document-save
else
    notify-send "NeoCognito" "Vault synced: no changes detected." -i document-save
fi

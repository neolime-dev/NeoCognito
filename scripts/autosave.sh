#!/bin/bash

# Configuration
VAULT_DIR="$HOME/Vault"
TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")

cd "$VAULT_DIR" || exit

# Check for changes
if [[ -n $(git status -s) ]]; then
    git add .
    git commit -m "Auto-save: $TIMESTAMP"
    
    # Optional: Push if remote exists
    # git push origin master
    
    notify-send "NeoCognito" "Vault saved and versioned." -i document-save # Translated
fi
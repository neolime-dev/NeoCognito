#!/bin/bash

# Configuration
VAULT_DIR="$HOME/Vault"
INBOX_FILE="$VAULT_DIR/00_Inbox.md"

# Check if Vault exists, create if not (redundancy)
mkdir -p "$VAULT_DIR"
touch "$INBOX_FILE"

# Launch Input Dialog (Walker)
# We pipe nothing into it so it's just an input box. 
# If the user types and hits enter, that text is returned.
CONTENT=$(echo "" | walker --dmenu --placeholder "Capture your idea..." --inputonly) # Translated

# Logic: Only process if content is not empty
if [ -n "$CONTENT" ]; then
    # Timestamp HH:MM
    TIMESTAMP=$(date +%H:%M)
    
    # Format: - HH:MM - Content (Fleeting Note style)
    ENTRY="- **$TIMESTAMP** - $CONTENT"
    
    # Append to Inbox
    echo "$ENTRY" >> "$INBOX_FILE"
    
    # Visual Confirmation
    notify-send "NeoCognito" "Captured: $CONTENT" -i accessories-text-editor # Translated
else
    # Optional: Notify if cancelled or empty? 
    # Keeping it silent is usually better for "no friction".
    exit 0
fi
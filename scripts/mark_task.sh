#!/bin/bash

TODO_FILE="$HOME/Vault/10_Planning/TODO_Today.md" # Renamed

SELECTION_WITH_NUM=$(grep -n "\- \[ \]" "$TODO_FILE") # Get selection with number
SELECTION=$(echo "$SELECTION_WITH_NUM" | sed 's/^[0-9]\+://' | walker --dmenu --placeholder "Complete task...") # Remove number before sending to walker

# 2. Validate selection
if [ -z "$SELECTION" ]; then
    exit 0
fi

# 3. Extract line number from the original selection with number
LINE_NUM=$(echo "$SELECTION_WITH_NUM" | grep -F "$SELECTION" | cut -d':' -f1)

# Validate if LINE_NUM is a number
if ! [[ "$LINE_NUM" =~ ^[0-9]+$ ]]; then
    notify-send "NeoCognito" "Error: Could not identify task line." -i dialog-error # Translated
    exit 1
fi

# 4. Mark as done on that specific line
sed -i "${LINE_NUM}s/\[ \]/\[x\]/" "$TODO_FILE"

# 5. Feedback
notify-send "NeoCognito" "Task completed!" -i task-complete # Translated
#!/bin/bash

TODO_FILE="$HOME/Vault/10_Planning/TODO_Today.md" # Renamed

# 1. Select the task
SELECTION=$(grep -n "\- \[ \]" "$TODO_FILE" | walker --dmenu --placeholder "Complete task..." --inputonly) # Translated

# 2. Validate selection
if [ -z "$SELECTION" ]; then
    exit 0
fi

# 3. Extract line number
LINE_NUM=$(echo "$SELECTION" | cut -d':' -f1)

# Validate if LINE_NUM is a number
if ! [[ "$LINE_NUM" =~ ^[0-9]+$ ]]; then
    notify-send "NeoCognito" "Error: Could not identify task line." -i dialog-error # Translated
    exit 1
fi

# 4. Mark as done on that specific line
sed -i "${LINE_NUM}s/\[ \]/\[x\]/" "$TODO_FILE"

# 5. Feedback
notify-send "NeoCognito" "Task completed!" -i task-complete # Translated
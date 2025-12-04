#!/bin/bash

TODO_FILE="$HOME/Vault/10_Planning/TODO_Today.md"

# 1. Get raw lines with line numbers (e.g., "3:- [ ] Task")
RAW_LIST=$(grep -n "\- \[ \]" "$TODO_FILE")

# If list is empty, exit
if [ -z "$RAW_LIST" ]; then
    notify-send "NeoCognito" "No pending tasks!" -i dialog-information
    exit 0
fi

# 2. Prepare visual list (remove numbers) and get SELECTED INDEX
# -i: Returns the index (0, 1, 2...) instead of the text string
SELECTED_INDEX=$(echo "$RAW_LIST" | sed 's/^[0-9]\+://' | walker --dmenu --placeholder "Complete task..." --index)

# 3. Validate selection
# Check if SELECTED_INDEX is an integer (handles Cancel/Esc)
if ! [[ "$SELECTED_INDEX" =~ ^[0-9]+$ ]]; then
    exit 0
fi

# 4. Map Index to Line Number
# Bash arrays or sed can handle this. Since index is 0-based, we want line (Index + 1) from RAW_LIST
SED_LINE=$((SELECTED_INDEX + 1))
LINE_INFO=$(echo "$RAW_LIST" | sed -n "${SED_LINE}p")

# Extract the actual file line number (before the colon)
REAL_LINE_NUM=$(echo "$LINE_INFO" | cut -d':' -f1)

# 5. Execute Mark as Done
if [[ "$REAL_LINE_NUM" =~ ^[0-9]+$ ]]; then
    sed -i "${REAL_LINE_NUM}s/\[ \]/\[x\]/" "$TODO_FILE"
    notify-send "NeoCognito" "Task completed!" -i task-complete
else
    notify-send "NeoCognito" "Error processing task." -i dialog-error
fi

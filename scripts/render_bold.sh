#!/bin/bash

FILE="$1"
MODE="$2"

if [ "$MODE" == "tail" ]; then
    CMD="tail -n 15"
else
    CMD="cat"
fi

# Apply bold formatting using perl
$CMD "$FILE" | \
fold -s -w 50 | \
perl -pe 's/\*\*(.*?)\*\*/\${font Monospace:style=Bold}$1\${font}/g'
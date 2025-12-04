#!/bin/bash

# Configuration Paths
CONFIG_DIR="$HOME/Dev_Pro/NeoCognito/config/conky"
DUAL_CONF="$CONFIG_DIR/neocognito_dual.conf"
SINGLE_CONF="$CONFIG_DIR/neocognito_single.conf"

# Kill old instances
killall conky 2>/dev/null
sleep 1

# Detect Monitor Count
# Method 1: Try hyprctl (Best for Hyprland)
if command -v hyprctl &> /dev/null; then
    MONITORS=$(hyprctl monitors | grep "Monitor" | wc -l)
# Method 2: Try xrandr (Fallback)
elif command -v xrandr &> /dev/null; then
    MONITORS=$(xrandr --listmonitors | grep -v "Monitors:" | wc -l)
else
    # Fallback safe: Assume 1
    MONITORS=1
fi

# Launch Logic
if [ "$MONITORS" -gt 1 ]; then
    notify-send "NeoCognito" "Dual Monitor detected. Loading Wall on Secondary."
    conky -c "$DUAL_CONF" &
else
    notify-send "NeoCognito" "Single Monitor detected. Loading Wall on Primary."
    conky -c "$SINGLE_CONF" &
fi
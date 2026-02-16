#!/bin/bash
# ==========================================
#  Figma Discord Rich Presence - macOS Uninstaller
# ==========================================

BINARY_NAME="figma-rpc"
PLIST_LABEL="com.figma.discord-rpc"
PLIST_FILE="$HOME/Library/LaunchAgents/${PLIST_LABEL}.plist"
DEFAULT_INSTALL_DIR="/usr/local/bin"

echo "=========================================="
echo "  Figma Discord RPC - Uninstaller"
echo "=========================================="
echo ""

read -p "Where was it installed? [$DEFAULT_INSTALL_DIR]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# Stop the LaunchAgent if running
if [ -f "$PLIST_FILE" ]; then
    echo "Stopping LaunchAgent..."
    launchctl unload "$PLIST_FILE" 2>/dev/null || true
    rm -f "$PLIST_FILE"
    echo "  LaunchAgent removed."
fi

# Remove the binary
INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"
if [ -f "$INSTALL_PATH" ]; then
    rm -f "$INSTALL_PATH"
    echo "  Binary removed: $INSTALL_PATH"
fi

# Remove .env
ENV_PATH="$INSTALL_DIR/.env"
if [ -f "$ENV_PATH" ]; then
    rm -f "$ENV_PATH"
    echo "  Config removed: $ENV_PATH"
fi

# Remove log file
if [ -f "/tmp/figma-rpc.log" ]; then
    rm -f "/tmp/figma-rpc.log"
    echo "  Log file removed."
fi

echo ""
echo "Figma Discord RPC has been uninstalled."
echo "=========================================="

#!/bin/bash
# ==========================================
#  Figma Discord Rich Presence - macOS Installer
# ==========================================

set -e

APP_NAME="Figma Discord RPC"
BINARY_NAME="figma-rpc"
PLIST_LABEL="com.figma.discord-rpc"
PLIST_FILE="$HOME/Library/LaunchAgents/${PLIST_LABEL}.plist"
DEFAULT_INSTALL_DIR="/usr/local/bin"

echo "=========================================="
echo "  $APP_NAME - Installer"
echo "=========================================="
echo ""

# ---- Step 1: Choose install location ----
read -p "Install location [$DEFAULT_INSTALL_DIR]: " INSTALL_DIR
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# Create directory if needed
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating directory: $INSTALL_DIR"
    sudo mkdir -p "$INSTALL_DIR"
fi

INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

# ---- Step 2: Build ----
echo ""
echo "[1/3] Building $BINARY_NAME..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_DIR="$SCRIPT_DIR/../src"

if [ ! -f "$SRC_DIR/main.go" ]; then
    echo "Error: Cannot find source code at $SRC_DIR"
    exit 1
fi

cd "$SRC_DIR"
go build -o "$INSTALL_PATH" .

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "       Built successfully!"

# ---- Step 3: Copy .env ----
echo "[2/3] Copying configuration..."
ENV_SRC="$SRC_DIR/.env"
ENV_DEST="$INSTALL_DIR/.env"

if [ -f "$ENV_SRC" ]; then
    cp "$ENV_SRC" "$ENV_DEST"
    echo "       .env copied to $INSTALL_DIR"
else
    echo "Warning: No .env file found. Create one at $ENV_DEST with your DISCORD_CLIENT_ID."
fi

# ---- Step 4: Ask about startup ----
echo ""
echo "[3/3] Startup configuration"
read -p "Would you like $APP_NAME to start automatically on login? (y/n): " AUTO_START

if [[ "$AUTO_START" =~ ^[Yy]$ ]]; then
    echo "       Setting up LaunchAgent..."

    # Create the LaunchAgent plist
    cat > "$PLIST_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${PLIST_LABEL}</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_PATH}</string>
    </array>
    <key>WorkingDirectory</key>
    <string>${INSTALL_DIR}</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/figma-rpc.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/figma-rpc.log</string>
</dict>
</plist>
EOF

    # Load the LaunchAgent
    launchctl load "$PLIST_FILE" 2>/dev/null || true

    echo "       Auto-start enabled! It will start on your next login."
    echo "       (Logs available at /tmp/figma-rpc.log)"
fi

# ---- Done! ----
echo ""
echo "=========================================="
echo "  Installation complete!"
echo ""
echo "  Binary:  $INSTALL_PATH"
echo "  Config:  $ENV_DEST"
if [[ "$AUTO_START" =~ ^[Yy]$ ]]; then
echo "  Startup: Enabled (LaunchAgent)"
fi
echo ""
echo "  To run manually:  $INSTALL_PATH"
echo "  To uninstall:     ./uninstall_macos.sh"
echo "=========================================="

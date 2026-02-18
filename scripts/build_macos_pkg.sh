#!/bin/bash
set -e

# Resolve paths relative to this script's location (works from any CWD)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

APP_NAME="Figma Discord RPC"
BINARY_NAME="figma-rpc"
IDENTIFIER="com.figma.discord-rpc"
VERSION="2.0.0"
INSTALL_DIR="/usr/local/bin"

echo "Building macOS package for $APP_NAME..."

# 1. Build binary
echo "[1/4] Compiling binary..."
mkdir -p "$ROOT_DIR/dist/root$INSTALL_DIR"
mkdir -p "$ROOT_DIR/dist/build"
cd "$ROOT_DIR/src"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o "$ROOT_DIR/dist/build/${BINARY_NAME}-amd64" .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$ROOT_DIR/dist/build/${BINARY_NAME}-arm64" .
lipo -create \
  -output "$ROOT_DIR/dist/root$INSTALL_DIR/$BINARY_NAME" \
  "$ROOT_DIR/dist/build/${BINARY_NAME}-amd64" \
  "$ROOT_DIR/dist/build/${BINARY_NAME}-arm64"

# 2. Prepare payload
echo "[2/5] Preparing payload..."

# Note: The LaunchAgent plist is installed by the postinstall script
# into ~/Library/LaunchAgents (per-user) so it runs in the correct
# user context and can request Accessibility permissions.

# 3. Create postinstall script
echo "[3/5] Creating postinstall script..."
mkdir -p "$ROOT_DIR/dist/scripts"
cat > "$ROOT_DIR/dist/scripts/postinstall" <<'POSTINSTALL'
#!/bin/bash

IDENTIFIER="com.figma.discord-rpc"
BINARY_PATH="/usr/local/bin/figma-rpc"
LOGGED_IN_USER=$(stat -f "%Su" /dev/console)
USER_HOME=$(dscl . -read /Users/$LOGGED_IN_USER NFSHomeDirectory | awk '{print $2}')
PLIST_DIR="$USER_HOME/Library/LaunchAgents"
PLIST_PATH="$PLIST_DIR/$IDENTIFIER.plist"

# Unload existing agent if present
if [ -f "$PLIST_PATH" ]; then
    sudo -u "$LOGGED_IN_USER" launchctl unload "$PLIST_PATH" 2>/dev/null || true
fi

# Create LaunchAgents directory if needed
sudo -u "$LOGGED_IN_USER" mkdir -p "$PLIST_DIR"

# Write the LaunchAgent plist
cat > "$PLIST_PATH" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$IDENTIFIER</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/usr/local/bin</string>
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

# Fix ownership so launchd loads it properly
chown "$LOGGED_IN_USER" "$PLIST_PATH"

# Load the LaunchAgent immediately so it starts without reboot
sudo -u "$LOGGED_IN_USER" launchctl load "$PLIST_PATH"

exit 0
POSTINSTALL
chmod +x "$ROOT_DIR/dist/scripts/postinstall"

# 4. Build component package
echo "[4/5] Building component package..."
pkgbuild --root "$ROOT_DIR/dist/root" \
         --scripts "$ROOT_DIR/dist/scripts" \
         --identifier "$IDENTIFIER" \
         --version "$VERSION" \
         --install-location / \
         "$ROOT_DIR/dist/component.pkg"

# 5. Build product archive (distribution pkg)
echo "[5/5] Building distribution package..."
productbuild --distribution "$SCRIPT_DIR/distribution.xml" \
             --package-path "$ROOT_DIR/dist" \
             --resources "$SCRIPT_DIR/resources" \
             "$ROOT_DIR/dist/FigmaRPC_Installer.pkg"

echo "Done! Package at dist/FigmaRPC_Installer.pkg"
echo ""
echo "IMPORTANT: Users must grant Accessibility permissions to figma-rpc"
echo "System Settings -> Privacy & Security -> Accessibility -> add /usr/local/bin/figma-rpc"

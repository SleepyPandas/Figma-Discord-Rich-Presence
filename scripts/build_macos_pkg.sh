#!/bin/bash
set -e

APP_NAME="Figma Discord RPC"
BINARY_NAME="figma-rpc"
IDENTIFIER="com.figma.discord-rpc"
VERSION="1.0.0"

echo "Building macOS package for $APP_NAME..."

# 1. Build binary
echo "[1/4] Compiling binary..."
cd ../src
GOOS=darwin GOARCH=amd64 go build -o ../dist/root/usr/local/bin/$BINARY_NAME .
cd ../scripts

# 2. Prepare payload
echo "[2/4] Preparing payload..."
mkdir -p ../dist/root/usr/local/bin
mkdir -p ../dist/root/Library/LaunchAgents

# Create LaunchAgent plist
cat > ../dist/root/Library/LaunchAgents/$IDENTIFIER.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$IDENTIFIER</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/$BINARY_NAME</string>
    </array>
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

# 3. Build component package
echo "[3/4] Building component package..."
pkgbuild --root ../dist/root \
         --identifier $IDENTIFIER \
         --version $VERSION \
         --install-location / \
         ../dist/component.pkg

# 4. Build product archive (distribution pkg)
echo "[4/4] Building distribution package..."
productbuild --distribution distribution.xml \
             --package-path ../dist \
             --resources resources \
             ../dist/FigmaRPC_Installer.pkg

echo "Done! Package at dist/FigmaRPC_Installer.pkg"

#!/bin/bash
set -e

REPO="TheSpiritMan/Cloudflared-TCP"
TMP_DEB="/tmp/cloudflared-tcp_latest.deb"

echo "🔍 Checking dependencies..."
if ! command -v jq >/dev/null 2>&1; then
    sudo apt update && sudo apt install jq -y
fi

echo "🌐 Fetching latest release URL..."
LATEST_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" \
  | jq -r '.assets[] | select(.name | endswith(".deb")) | .browser_download_url')

if [ -z "$LATEST_URL" ]; then
    echo "❌ Error: Could not find a .deb asset in the latest release of $REPO"
    exit 1
fi

echo "⬇️  Downloading and installing package..."
curl -L "$LATEST_URL" -o "$TMP_DEB"
sudo apt install "$TMP_DEB" -f -y

# Cleanup the temporary .deb file
rm -f "$TMP_DEB"

echo "✅ cloudflared-tcp installation complete!"
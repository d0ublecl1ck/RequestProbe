#!/bin/zsh

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="RequestProbe"
APP_PATH="$PROJECT_ROOT/build/bin/${APP_NAME}.app"
DMG_PATH="$PROJECT_ROOT/build/bin/${APP_NAME}.dmg"
STAGING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/${APP_NAME}-dmg.XXXXXX")"
VOLUME_NAME="${APP_NAME} Installer"

cleanup() {
  rm -rf "$STAGING_DIR"
}

trap cleanup EXIT

cd "$PROJECT_ROOT"

echo "==> Building ${APP_NAME}.app"
wails build -clean -platform darwin/universal

if [[ ! -d "$APP_PATH" ]]; then
  echo "Expected app bundle not found: $APP_PATH" >&2
  exit 1
fi

echo "==> Preparing DMG staging directory"
cp -R "$APP_PATH" "$STAGING_DIR/"
ln -s /Applications "$STAGING_DIR/Applications"
rm -f "$DMG_PATH"

echo "==> Creating ${DMG_PATH}"
hdiutil create \
  -volname "$VOLUME_NAME" \
  -srcfolder "$STAGING_DIR" \
  -ov \
  -format UDZO \
  "$DMG_PATH"

echo "==> Done"
echo "App: $APP_PATH"
echo "DMG: $DMG_PATH"

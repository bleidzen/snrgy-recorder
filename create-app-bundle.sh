#!/bin/bash
set -e

APP_NAME="SNRGY Recorder"
BUNDLE_ID="com.snrgy.recorder"
VERSION="1.0.0"

echo "Creating macOS .app bundle..."

# Create bundle structure
mkdir -p "$APP_NAME.app/Contents/MacOS"
mkdir -p "$APP_NAME.app/Contents/Resources"

# Copy binary
cp snrgy-recorder "$APP_NAME.app/Contents/MacOS/"

# Create Info.plist
cat > "$APP_NAME.app/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>snrgy-recorder</string>
    <key>CFBundleIdentifier</key>
    <string>$BUNDLE_ID</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSMicrophoneUsageDescription</key>
    <string>SNRGY Recorder needs microphone access to record audio.</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

echo "Done! Created: $APP_NAME.app"
echo ""
echo "To distribute, you can:"
echo "  1. Zip the .app: zip -r 'SNRGY Recorder.zip' 'SNRGY Recorder.app'"
echo "  2. Create a DMG: hdiutil create -volname 'SNRGY Recorder' -srcfolder '$APP_NAME.app' -ov 'SNRGY-Recorder.dmg'"

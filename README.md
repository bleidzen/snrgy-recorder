# SNRGY Recorder

Native desktop audio recorder for Windows and macOS.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- **Windows**: No additional requirements
- **macOS**: Xcode Command Line Tools (`xcode-select --install`)

## Building

### Windows

```batch
build.bat
```

Output: `snrgy-recorder.exe` (~12MB)

### macOS

```bash
chmod +x build.sh create-app-bundle.sh
./build.sh
./create-app-bundle.sh
```

Output: `SNRGY Recorder.app`

### Cross-compile (requires Make)

```bash
make all        # Build for Windows + macOS
make windows    # Windows only
make darwin     # macOS only (Intel + Apple Silicon)
```

## Configuration

Set your token via environment variable:

```bash
# Windows (PowerShell)
$env:SNRGY_TOKEN = "your-token-here"
.\snrgy-recorder.exe

# macOS/Linux
SNRGY_TOKEN="your-token-here" ./snrgy-recorder
```

Or edit `main.go` line 26 to hardcode it:

```go
token = "your-actual-token"
```

## Usage

- **Start Recording**: `Ctrl+Shift+R` or click "Start Recording" in system tray
- **Stop Recording**: `Ctrl+Shift+S` or click "Stop Recording" in system tray
- **Quit**: Click "Quit" in system tray

## Distribution

### Windows
Just distribute the `.exe` file. No installer needed.

### macOS
Distribute the `.app` bundle or create a DMG:

```bash
hdiutil create -volname "SNRGY Recorder" -srcfolder "SNRGY Recorder.app" -ov "SNRGY-Recorder.dmg"
```

**Note**: For macOS distribution outside the App Store, users may need to right-click → Open the first time, or you'll need to notarize the app with an Apple Developer account.

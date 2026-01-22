//go:build darwin

package main

// On macOS, we skip systray to avoid duplicate symbol conflicts
// between getlantern/systray and fyne.io/systray.
// The app works via hotkeys and the main window.

func initSystray() {
	// No systray on macOS - use hotkeys instead
	go registerHotkeys()
}

func quitSystray() {
	// No-op on macOS
}

func setSystrayTooltip(text string) {
	// No-op on macOS
}

func updateSystrayRecordingStarted() {
	// No-op on macOS
}

func updateSystrayRecordingStopped() {
	// No-op on macOS
}

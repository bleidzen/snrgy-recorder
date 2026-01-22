//go:build windows

package main

import (
	"github.com/getlantern/systray"
)

var (
	menuStart *systray.MenuItem
	menuStop  *systray.MenuItem
)

func initSystray() {
	go systray.Run(onSystrayReady, onSystrayExit)
}

func quitSystray() {
	systray.Quit()
}

func setSystrayTooltip(text string) {
	systray.SetTooltip(text)
}

func onSystrayReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("SNRGY")
	systray.SetTooltip("SNRGY Recorder")

	menuStart = systray.AddMenuItem("Start Recording", "Start recording audio")
	menuStop = systray.AddMenuItem("Stop Recording", "Stop recording audio")
	menuStop.Disable()

	systray.AddSeparator()
	menuShow := systray.AddMenuItem("Show Window", "Show the main window")
	menuQuit := systray.AddMenuItem("Quit", "Quit the application")

	go registerHotkeys()

	go func() {
		for {
			select {
			case <-menuStart.ClickedCh:
				startRecording()
			case <-menuStop.ClickedCh:
				stopRecording()
			case <-menuShow.ClickedCh:
				mainWindow.Show()
			case <-menuQuit.ClickedCh:
				stopRecording()
				fyneApp.Quit()
			}
		}
	}()
}

func onSystrayExit() {
	stopRecording()
}

func updateSystrayRecordingStarted() {
	if menuStart != nil {
		menuStart.Disable()
		menuStop.Enable()
	}
	systray.SetTooltip("SNRGY Recorder - Recording...")
}

func updateSystrayRecordingStopped() {
	if menuStart != nil {
		menuStart.Enable()
		menuStop.Disable()
	}
	systray.SetTooltip("SNRGY Recorder")
}

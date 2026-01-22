package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/malgo"
	"github.com/getlantern/systray"
	"golang.design/x/hotkey"
)

const (
	apiURL       = "https://mfuklmpmokuexofevloc.supabase.co/functions/v1/desktop-recorder"
	sampleRate   = 16000
	channels     = 1
	chunkSeconds = 5
)

// Config holds user settings
type Config struct {
	Token       string `json:"token"`
	StartHotkey string `json:"start_hotkey"`
	StopHotkey  string `json:"stop_hotkey"`
}

// SNRGYTheme is our custom dark theme
type SNRGYTheme struct{}

var _ fyne.Theme = (*SNRGYTheme)(nil)

func (t *SNRGYTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 45, G: 45, B: 45, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 35, G: 35, B: 35, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 255, G: 82, B: 82, A: 255}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0, G: 200, B: 150, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0, G: 180, B: 130, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 220, B: 160, A: 255} // SNRGY accent green
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0, G: 150, B: 110, A: 128}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0, G: 220, B: 160, A: 255}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 193, B: 7, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *SNRGYTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *SNRGYTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *SNRGYTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 12
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	default:
		return theme.DefaultTheme().Size(name)
	}
}

var (
	config        Config
	recording     = false
	recordingMu   sync.Mutex
	stopChan      chan struct{}
	menuStart     *systray.MenuItem
	menuStop      *systray.MenuItem
	logList       *widget.List
	logMessages   []logEntry
	logMu         sync.Mutex
	statusLabel   *widget.Label
	statusDot     *canvas.Circle
	fyneApp       fyne.App
	mainWindow    fyne.Window
	recordBtn     *widget.Button
	stopBtn       *widget.Button
)

type logEntry struct {
	Time    string
	Message string
	IsError bool
}

var appIcon fyne.Resource

func main() {
	loadConfig()

	fyneApp = app.NewWithID("com.snrgy.recorder")
	fyneApp.Settings().SetTheme(&SNRGYTheme{})

	// Generate and set app icon
	appIcon = generateAppIcon()
	fyneApp.SetIcon(appIcon)

	mainWindow = fyneApp.NewWindow("SNRGY Recorder")
	mainWindow.SetContent(createUI())
	mainWindow.Resize(fyne.NewSize(420, 580))
	mainWindow.CenterOnScreen()
	mainWindow.SetIcon(appIcon)

	// Start systray in background
	go func() {
		time.Sleep(500 * time.Millisecond)
		systray.Run(onSystrayReady, onSystrayExit)
	}()

	mainWindow.SetCloseIntercept(func() {
		mainWindow.Hide()
	})

	addLog("Welcome to SNRGY Recorder", false)
	if config.Token == "" || config.Token == "YOUR_TOKEN_HERE" {
		addLog("Please enter your API token to get started", false)
	} else {
		addLog("Ready to record", false)
	}

	mainWindow.ShowAndRun()
}

func createUI() fyne.CanvasObject {
	// Header with logo
	logoText := canvas.NewText("SNRGY", color.White)
	logoText.TextSize = 28
	logoText.TextStyle = fyne.TextStyle{Bold: true}

	studioText := canvas.NewText(" Studios™", color.NRGBA{R: 120, G: 120, B: 120, A: 255})
	studioText.TextSize = 16

	subtitleText := canvas.NewText("Desktop Recorder", color.NRGBA{R: 0, G: 220, B: 160, A: 255})
	subtitleText.TextSize = 12
	subtitleText.Alignment = fyne.TextAlignCenter

	logoRow := container.NewHBox(layout.NewSpacer(), logoText, studioText, layout.NewSpacer())
	subtitleRow := container.NewHBox(layout.NewSpacer(), subtitleText, layout.NewSpacer())

	// Status indicator - use a colored rectangle instead of circle for reliability
	statusDot = canvas.NewCircle(color.NRGBA{R: 0, G: 200, B: 140, A: 255})
	dotContainer := container.NewWithoutLayout(statusDot)
	dotContainer.Resize(fyne.NewSize(14, 14))
	statusDot.Resize(fyne.NewSize(10, 10))
	statusDot.Move(fyne.NewPos(2, 2))

	statusLabel = widget.NewLabel("Ready")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	statusBox := container.NewHBox(
		layout.NewSpacer(),
		container.NewPadded(dotContainer),
		statusLabel,
		layout.NewSpacer(),
	)

	// Separator
	sep1 := canvas.NewRectangle(color.NRGBA{R: 50, G: 50, B: 50, A: 255})
	sep1.SetMinSize(fyne.NewSize(0, 1))

	// Settings form - more compact
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Enter your API token")
	tokenEntry.SetText(config.Token)
	tokenEntry.OnChanged = func(s string) {
		config.Token = s
	}

	startKeyEntry := widget.NewEntry()
	startKeyEntry.SetText(config.StartHotkey)
	startKeyEntry.OnChanged = func(s string) {
		config.StartHotkey = s
	}

	stopKeyEntry := widget.NewEntry()
	stopKeyEntry.SetText(config.StopHotkey)
	stopKeyEntry.OnChanged = func(s string) {
		config.StopHotkey = s
	}

	// Use form widget for cleaner layout
	settingsForm := widget.NewForm(
		widget.NewFormItem("API Token", tokenEntry),
		widget.NewFormItem("Start Hotkey", startKeyEntry),
		widget.NewFormItem("Stop Hotkey", stopKeyEntry),
	)

	saveBtn := widget.NewButton("Save Settings", func() {
		saveConfig()
		addLog("Settings saved", false)
		go registerHotkeys()
	})
	saveBtn.Importance = widget.HighImportance

	// Control buttons
	recordBtn = widget.NewButton("⏺ Start Recording", func() {
		startRecording()
	})
	recordBtn.Importance = widget.SuccessImportance

	stopBtn = widget.NewButton("⏹ Stop Recording", func() {
		stopRecording()
	})
	stopBtn.Importance = widget.DangerImportance
	stopBtn.Disable()

	controlBtns := container.NewGridWithColumns(2, recordBtn, stopBtn)

	// Separator
	sep2 := canvas.NewRectangle(color.NRGBA{R: 50, G: 50, B: 50, A: 255})
	sep2.SetMinSize(fyne.NewSize(0, 1))

	// Log - simple text area instead of list for reliability
	logMessages = make([]logEntry, 0)
	logList = widget.NewList(
		func() int { return len(logMessages) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Log entry")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			logMu.Lock()
			defer logMu.Unlock()
			if id < len(logMessages) {
				entry := logMessages[id]
				label := obj.(*widget.Label)
				if entry.IsError {
					label.SetText("⚠ " + entry.Time + " " + entry.Message)
				} else {
					label.SetText("• " + entry.Time + " " + entry.Message)
				}
			}
		},
	)

	logLabel := widget.NewLabel("Activity Log")
	logLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Main layout - everything in one VBox with proper structure
	return container.NewPadded(
		container.NewVBox(
			// Header section
			logoRow,
			subtitleRow,
			widget.NewSeparator(),

			// Status
			statusBox,
			widget.NewSeparator(),

			// Settings
			settingsForm,
			saveBtn,
			widget.NewSeparator(),

			// Controls
			controlBtns,
			widget.NewSeparator(),

			// Log section - give it remaining space
			logLabel,
			logList,
		),
	)
}

func addLog(msg string, isError bool) {
	logMu.Lock()
	entry := logEntry{
		Time:    time.Now().Format("15:04:05"),
		Message: msg,
		IsError: isError,
	}
	logMessages = append(logMessages, entry)
	logMu.Unlock()

	if logList != nil {
		logList.Refresh()
		logList.ScrollToBottom()
	}
	fmt.Printf("[%s] %s\n", entry.Time, msg)
}

func setStatus(status string, isRecording bool) {
	if statusLabel != nil {
		statusLabel.SetText(status)
	}
	if statusDot != nil {
		if isRecording {
			statusDot.FillColor = color.NRGBA{R: 255, G: 80, B: 80, A: 255} // Red for recording
		} else {
			statusDot.FillColor = color.NRGBA{R: 0, G: 200, B: 140, A: 255} // Green for ready
		}
		canvas.Refresh(statusDot)
	}
}

func getConfigPath() string {
	configDir, _ := os.UserConfigDir()
	dir := filepath.Join(configDir, "SNRGYRecorder")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

func loadConfig() {
	config = Config{
		Token:       "",
		StartHotkey: "Ctrl+Shift+R",
		StopHotkey:  "Ctrl+Shift+S",
	}
	data, err := os.ReadFile(getConfigPath())
	if err == nil {
		json.Unmarshal(data, &config)
	}
}

func saveConfig() {
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(getConfigPath(), data, 0644)
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

func registerHotkeys() {
	startMods, startKey := parseHotkey(config.StartHotkey)
	if startKey != 0 {
		hkStart := hotkey.New(startMods, startKey)
		if err := hkStart.Register(); err != nil {
			addLog(fmt.Sprintf("Hotkey error: %v", err), true)
		} else {
			addLog(fmt.Sprintf("Registered: %s", config.StartHotkey), false)
			go func() {
				for range hkStart.Keydown() {
					startRecording()
				}
			}()
		}
	}

	stopMods, stopKey := parseHotkey(config.StopHotkey)
	if stopKey != 0 {
		hkStop := hotkey.New(stopMods, stopKey)
		if err := hkStop.Register(); err != nil {
			addLog(fmt.Sprintf("Hotkey error: %v", err), true)
		} else {
			addLog(fmt.Sprintf("Registered: %s", config.StopHotkey), false)
			go func() {
				for range hkStop.Keydown() {
					stopRecording()
				}
			}()
		}
	}
}

func parseHotkey(s string) ([]hotkey.Modifier, hotkey.Key) {
	var mods []hotkey.Modifier
	var key hotkey.Key
	parts := splitHotkey(s)
	for _, p := range parts {
		switch p {
		case "Ctrl":
			mods = append(mods, hotkey.ModCtrl)
		case "Shift":
			mods = append(mods, hotkey.ModShift)
		case "Alt":
			mods = append(mods, hotkey.ModAlt)
		default:
			if len(p) == 1 {
				key = hotkey.Key(p[0])
			}
		}
	}
	return mods, key
}

func splitHotkey(s string) []string {
	var parts []string
	var current string
	for _, c := range s {
		if c == '+' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func startRecording() {
	recordingMu.Lock()
	if recording {
		recordingMu.Unlock()
		return
	}
	recording = true
	stopChan = make(chan struct{})
	recordingMu.Unlock()

	if menuStart != nil {
		menuStart.Disable()
		menuStop.Enable()
	}
	if recordBtn != nil {
		recordBtn.Disable()
		stopBtn.Enable()
	}
	setStatus("Recording", true)
	systray.SetTooltip("SNRGY Recorder - Recording...")

	addLog("Starting recording...", false)

	resp, err := apiRequest(map[string]interface{}{
		"action": "start",
		"title":  "Desktop Recording",
	})
	if err != nil {
		addLog(fmt.Sprintf("API error: %v", err), true)
	} else {
		addLog("Connected to server", false)
	}

	if meetingURL, ok := resp["meeting_url"].(string); ok && meetingURL != "" {
		addLog("Opening live view...", false)
		openBrowser(meetingURL)
	}

	go captureAudio()
}

func stopRecording() {
	recordingMu.Lock()
	if !recording {
		recordingMu.Unlock()
		return
	}
	recording = false
	close(stopChan)
	recordingMu.Unlock()

	if menuStart != nil {
		menuStart.Enable()
		menuStop.Disable()
	}
	if recordBtn != nil {
		recordBtn.Enable()
		stopBtn.Disable()
	}
	setStatus("Ready", false)
	systray.SetTooltip("SNRGY Recorder")

	_, err := apiRequest(map[string]interface{}{
		"action": "stop",
	})
	if err != nil {
		addLog(fmt.Sprintf("API error: %v", err), true)
	}

	addLog("Recording stopped", false)
}

func captureAudio() {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		addLog(fmt.Sprintf("Audio error: %v", err), true)
		return
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = channels
	deviceConfig.SampleRate = sampleRate

	var audioBuffer []byte
	var bufferMu sync.Mutex
	samplesPerChunk := sampleRate * chunkSeconds * channels * 2

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		bufferMu.Lock()
		audioBuffer = append(audioBuffer, pSample...)
		bufferMu.Unlock()
	}

	captureCallbacks := malgo.DeviceCallbacks{Data: onRecvFrames}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if err != nil {
		addLog(fmt.Sprintf("Device error: %v", err), true)
		return
	}
	defer device.Uninit()

	if err = device.Start(); err != nil {
		addLog(fmt.Sprintf("Start error: %v", err), true)
		return
	}
	defer device.Stop()

	addLog("Audio capture started", false)

	chunkIndex := 0
	ticker := time.NewTicker(time.Second * chunkSeconds)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			bufferMu.Lock()
			if len(audioBuffer) >= samplesPerChunk {
				chunk := audioBuffer[:samplesPerChunk]
				audioBuffer = audioBuffer[samplesPerChunk:]
				bufferMu.Unlock()

				wavData := createWav(chunk)
				audioB64 := base64.StdEncoding.EncodeToString(wavData)

				go func(idx int, data string) {
					_, err := apiRequest(map[string]interface{}{
						"action":       "chunk",
						"audio_base64": data,
						"chunk_index":  idx,
					})
					if err != nil {
						addLog(fmt.Sprintf("Chunk %d failed", idx), true)
					} else {
						addLog(fmt.Sprintf("Chunk %d sent", idx), false)
					}
				}(chunkIndex, audioB64)
				chunkIndex++
			} else {
				bufferMu.Unlock()
			}
		}
	}
}

func createWav(samples []byte) []byte {
	buf := new(bytes.Buffer)
	dataSize := uint32(len(samples))
	fileSize := dataSize + 36
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, fileSize)
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(channels))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*channels*2))
	binary.Write(buf, binary.LittleEndian, uint16(channels*2))
	binary.Write(buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, dataSize)
	buf.Write(samples)
	return buf.Bytes()
}

func apiRequest(payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-recorder-token", config.Token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		addLog(fmt.Sprintf("Browser error: %v", err), true)
	}
}

func getIcon() []byte {
	// Use smaller 32x32 icon for system tray
	return generateTrayIcon()
}

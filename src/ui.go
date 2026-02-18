package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	colorConnected    = color.NRGBA{R: 76, G: 175, B: 80, A: 255}  // Green
	colorDisconnected = color.NRGBA{R: 244, G: 67, B: 54, A: 255}  // Red
)

// UIEvents carries signals from the UI to the RPC loop.
type UIEvents struct {
	Disconnect    chan struct{}
	Reconnect     chan struct{}
	ConfigChanged chan *Config
}

// NewUIEvents creates a new UIEvents with buffered channels.
func NewUIEvents() *UIEvents {
	return &UIEvents{
		Disconnect:    make(chan struct{}, 1),
		Reconnect:     make(chan struct{}, 1),
		ConfigChanged: make(chan *Config, 1),
	}
}

// statusIndicator holds the UI elements for the connection status display.
type statusIndicator struct {
	circle *canvas.Circle
	label  *widget.Label
}

func newStatusIndicator() *statusIndicator {
	circle := canvas.NewCircle(colorConnected)
	circle.Resize(fyne.NewSize(12, 12))

	label := widget.NewLabel("Connected to Discord")
	label.TextStyle = fyne.TextStyle{Bold: true}

	return &statusIndicator{circle: circle, label: label}
}

func (s *statusIndicator) setConnected() {
	s.circle.FillColor = colorConnected
	s.circle.Refresh()
	s.label.SetText("Connected to Discord")
}

func (s *statusIndicator) setDisconnected() {
	s.circle.FillColor = colorDisconnected
	s.circle.Refresh()
	s.label.SetText("Disconnected from Discord")
}

// AppUI holds all the Fyne app components.
type AppUI struct {
	App    fyne.App
	Window fyne.Window
	Events *UIEvents
	Config *Config
	Status *statusIndicator
}

// SetupUI creates the Fyne application, window, system tray, and all widgets.
// It returns an AppUI that the caller can use to run the app.
func SetupUI(cfg *Config, events *UIEvents) *AppUI {
	fyneApp := app.NewWithID("com.figma.discord-rpc")
	fyneApp.Settings().SetTheme(theme.DarkTheme())

	win := fyneApp.NewWindow(fmt.Sprintf("Figma Discord Rich Presence  v%s", appVersion))
	win.Resize(fyne.NewSize(420, 380))
	win.SetFixedSize(true)
	win.CenterOnScreen()

	ui := &AppUI{
		App:    fyneApp,
		Window: win,
		Events: events,
		Config: cfg,
		Status: newStatusIndicator(),
	}

	// Build the window content
	win.SetContent(ui.buildContent())

	// Minimize to tray on close instead of quitting
	win.SetCloseIntercept(func() {
		win.Hide()
	})

	// Setup system tray
	ui.setupSystemTray()

	return ui
}

// buildContent creates the main settings panel layout.
func (ui *AppUI) buildContent() fyne.CanvasObject {
	// ── Status Section ──
	statusCircle := container.New(layout.NewCenterLayout(), ui.Status.circle)
	statusCircle.Resize(fyne.NewSize(12, 12))

	statusRow := container.NewHBox(
		statusCircle,
		ui.Status.label,
	)

	// ── Privacy Section ──
	privacyHeader := widget.NewLabel("Privacy")
	privacyHeader.TextStyle = fyne.TextStyle{Bold: true}

	privacyCheck := widget.NewCheck("Privacy Mode", func(checked bool) {
		ui.Config.PrivacyMode = checked
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving config:", err)
		}
		ui.notifyConfigChanged()
	})
	privacyCheck.Checked = ui.Config.PrivacyMode

	customLabelEntry := widget.NewEntry()
	customLabelEntry.SetPlaceHolder("Working on a project")
	customLabelEntry.SetText(ui.Config.CustomLabel)
	customLabelEntry.OnChanged = func(text string) {
		ui.Config.CustomLabel = text
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving config:", err)
		}
		ui.notifyConfigChanged()
	}

	customLabelForm := widget.NewFormItem("Custom Label", customLabelEntry)
	privacyForm := widget.NewForm(customLabelForm)

	privacySection := container.NewVBox(
		privacyHeader,
		privacyCheck,
		privacyForm,
	)

	// ── Connection Section ──
	connectionHeader := widget.NewLabel("Connection")
	connectionHeader.TextStyle = fyne.TextStyle{Bold: true}

	disconnectBtn := widget.NewButton("Disconnect", func() {
		ui.Config.RPCEnabled = false
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving config:", err)
		}
		ui.Status.setDisconnected()
		select {
		case ui.Events.Disconnect <- struct{}{}:
		default:
		}
	})

	reconnectBtn := widget.NewButton("Reconnect", func() {
		ui.Config.RPCEnabled = true
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving config:", err)
		}
		ui.Status.setConnected()
		select {
		case ui.Events.Reconnect <- struct{}{}:
		default:
		}
	})

	buttonRow := container.NewGridWithColumns(2, disconnectBtn, reconnectBtn)

	connectionSection := container.NewVBox(
		connectionHeader,
		buttonRow,
	)

	// ── Version Footer ──
	versionLabel := widget.NewLabel(fmt.Sprintf("v%s", appVersion))
	versionLabel.Alignment = fyne.TextAlignCenter
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}

	// ── Assemble ──
	content := container.NewVBox(
		statusRow,
		widget.NewSeparator(),
		privacySection,
		widget.NewSeparator(),
		connectionSection,
		widget.NewSeparator(),
		versionLabel,
	)

	return container.NewPadded(content)
}

// setupSystemTray configures the system tray icon and menu.
func (ui *AppUI) setupSystemTray() {
	if deskApp, ok := ui.App.(desktop.App); ok {
		menu := fyne.NewMenu("FigmaRPC",
			fyne.NewMenuItem("Show Settings", func() {
				ui.Window.Show()
				ui.Window.RequestFocus()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Disconnect from RPC", func() {
				ui.Config.RPCEnabled = false
				if err := ui.Config.Save(); err != nil {
					fmt.Println("Error saving config:", err)
				}
				ui.Status.setDisconnected()
				select {
				case ui.Events.Disconnect <- struct{}{}:
				default:
				}
			}),
			fyne.NewMenuItem("Reconnect to RPC", func() {
				ui.Config.RPCEnabled = true
				if err := ui.Config.Save(); err != nil {
					fmt.Println("Error saving config:", err)
				}
				ui.Status.setConnected()
				select {
				case ui.Events.Reconnect <- struct{}{}:
				default:
				}
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				ui.App.Quit()
			}),
		)
		deskApp.SetSystemTrayMenu(menu)
	}
}

// notifyConfigChanged sends the current config to the RPC loop (non-blocking).
func (ui *AppUI) notifyConfigChanged() {
	select {
	case ui.Events.ConfigChanged <- ui.Config:
	default:
	}
}

// Run starts the Fyne event loop. This blocks until the app exits.
// If FirstRun is true, the window is shown; otherwise it starts hidden in the tray.
func (ui *AppUI) Run() {
	if ui.Config.FirstRun {
		ui.Config.FirstRun = false
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving first-run flag:", err)
		}
		ui.Window.Show()
	}

	// Start the Fyne event loop (blocks main thread)
	ui.App.Run()
}

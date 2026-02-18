package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var (
	colorConnected    = color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	colorDisconnected = color.NRGBA{R: 244, G: 67, B: 54, A: 255}
	colorCardFill     = color.NRGBA{R: 10, G: 10, B: 10, A: 209}
	colorCardStroke   = color.NRGBA{R: 255, G: 255, B: 255, A: 36}
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
	fyneApp.Settings().SetTheme(newWebsiteDarkTheme())

	win := fyneApp.NewWindow(fmt.Sprintf("Figma Discord Rich Presence  v%s", appVersion))
	win.Resize(fyne.NewSize(460, 500))
	win.SetFixedSize(true)
	win.CenterOnScreen()

	ui := &AppUI{
		App:    fyneApp,
		Window: win,
		Events: events,
		Config: cfg,
		Status: newStatusIndicator(),
	}
	if !cfg.RPCEnabled {
		ui.Status.setDisconnected()
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

func (ui *AppUI) handleDisconnectAction() {
	ui.Config.RPCEnabled = false
	if err := ui.Config.Save(); err != nil {
		fmt.Println("Error saving config:", err)
	}
	ui.Status.setDisconnected()
	select {
	case ui.Events.Disconnect <- struct{}{}:
	default:
	}
}

func (ui *AppUI) handleReconnectAction() {
	ui.Config.RPCEnabled = true
	if err := ui.Config.Save(); err != nil {
		fmt.Println("Error saving config:", err)
	}
	ui.Status.setConnected()
	select {
	case ui.Events.Reconnect <- struct{}{}:
	default:
	}
}

func spacer(height float32) fyne.CanvasObject {
	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(0, height))
	return gap
}

func horizontalSpacer(width float32) fyne.CanvasObject {
	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(width, 0))
	return gap
}

func sectionHeader(title, subtitle string) fyne.CanvasObject {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	subtitleLabel := widget.NewLabel(subtitle)
	subtitleLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(titleLabel, subtitleLabel)
}

func sectionCard(objects ...fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(colorCardFill)
	background.StrokeColor = colorCardStroke
	background.StrokeWidth = 1
	background.CornerRadius = 14

	content := container.NewVBox(objects...)
	return container.NewStack(background, container.NewPadded(content))
}

// buildContent creates the main settings panel layout.
func (ui *AppUI) buildContent() fyne.CanvasObject {
	// Status section
	statusRow := container.NewHBox(
		ui.Status.circle,
		horizontalSpacer(8),
		ui.Status.label,
	)
	statusCard := sectionCard(
		sectionHeader("Status", "Current Discord Rich Presence connection."),
		spacer(8),
		statusRow,
	)

	// Privacy section
	privacyCheck := widget.NewCheck("Privacy Mode", func(checked bool) {
		ui.Config.PrivacyMode = checked
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving config:", err)
		}
		ui.notifyConfigChanged()
	})
	privacyCheck.Checked = ui.Config.PrivacyMode

	customLabelEntry := widget.NewEntry()
	customLabelEntry.SetPlaceHolder("xyz")
	customLabelEntry.SetText(ui.Config.CustomLabel)
	customLabelEntry.OnChanged = func(text string) {
		ui.Config.CustomLabel = text
		if err := ui.Config.Save(); err != nil {
			fmt.Println("Error saving config:", err)
		}
		ui.notifyConfigChanged()
	}

	customLabelLabel := widget.NewLabel("Replacement Label")
	customLabelLabel.TextStyle = fyne.TextStyle{Bold: true}

	privacyCard := sectionCard(
		sectionHeader("Privacy", "Hide project names and replace them with your custom text."),
		spacer(8),
		privacyCheck,
		spacer(4),
		customLabelLabel,
		customLabelEntry,
	)

	// Connection section
	disconnectBtn := widget.NewButton("Disconnect", ui.handleDisconnectAction)
	disconnectBtn.Importance = widget.DangerImportance

	reconnectBtn := widget.NewButton("Reconnect", ui.handleReconnectAction)
	reconnectBtn.Importance = widget.HighImportance

	connectionCard := sectionCard(
		sectionHeader("Connection", "Control Discord RPC without exiting the app."),
		spacer(8),
		container.NewGridWithColumns(2, disconnectBtn, reconnectBtn),
	)

	// Version footer
	versionLabel := widget.NewLabel(fmt.Sprintf("v%s", appVersion))
	versionLabel.Alignment = fyne.TextAlignCenter
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		statusCard,
		spacer(12),
		privacyCard,
		spacer(12),
		connectionCard,
		spacer(8),
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
			fyne.NewMenuItem("Disconnect from RPC", ui.handleDisconnectAction),
			fyne.NewMenuItem("Reconnect to RPC", ui.handleReconnectAction),
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

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Jared-Boschmann/skwad-linux/internal/models"
	"github.com/Jared-Boschmann/skwad-linux/internal/persistence"
)

// SettingsWindow wraps the settings UI in its own Fyne window.
type SettingsWindow struct {
	fyneApp fyne.App
	store   *persistence.Store
	window  fyne.Window
}

// NewSettingsWindow creates but does not show the settings window.
func NewSettingsWindow(app fyne.App, store *persistence.Store) *SettingsWindow {
	return &SettingsWindow{fyneApp: app, store: store}
}

// Show opens the settings window (or brings it to front if already open).
func (s *SettingsWindow) Show() {
	if s.window != nil {
		s.window.Show()
		return
	}
	s.window = s.fyneApp.NewWindow("Settings")
	s.window.SetContent(s.buildContent())
	s.window.Resize(fyne.NewSize(600, 500))
	s.window.SetOnClosed(func() { s.window = nil })
	s.window.Show()
}

func (s *SettingsWindow) buildContent() fyne.CanvasObject {
	settings := s.store.Settings()

	tabs := container.NewAppTabs(
		container.NewTabItem("General", s.generalTab(&settings)),
		container.NewTabItem("Terminal", s.terminalTab(&settings)),
		container.NewTabItem("MCP Server", s.mcpTab(&settings)),
		container.NewTabItem("Autopilot", s.autopilotTab(&settings)),
		container.NewTabItem("Appearance", s.appearanceTab(&settings)),
	)

	saveBtn := widget.NewButton("Save", func() {
		_ = s.store.SaveSettings(settings)
		if s.window != nil {
			s.window.Hide()
		}
	})

	return container.NewBorder(nil, saveBtn, nil, nil, tabs)
}

func (s *SettingsWindow) generalTab(settings *models.AppSettings) fyne.CanvasObject {
	restoreCheck := widget.NewCheck("Restore layout on launch", func(v bool) {
		settings.RestoreLayoutOnLaunch = v
	})
	restoreCheck.Checked = settings.RestoreLayoutOnLaunch

	trayCheck := widget.NewCheck("Keep in system tray", func(v bool) {
		settings.KeepInTray = v
	})
	trayCheck.Checked = settings.KeepInTray

	sourceEntry := widget.NewEntry()
	sourceEntry.SetText(settings.SourceBaseFolder)
	sourceEntry.OnChanged = func(v string) { settings.SourceBaseFolder = v }

	return container.NewVBox(
		restoreCheck,
		trayCheck,
		widget.NewLabel("Source base folder"),
		sourceEntry,
	)
}

func (s *SettingsWindow) terminalTab(settings *models.AppSettings) fyne.CanvasObject {
	fontEntry := widget.NewEntry()
	fontEntry.SetText(settings.TerminalFontName)
	fontEntry.OnChanged = func(v string) { settings.TerminalFontName = v }

	bgEntry := widget.NewEntry()
	bgEntry.SetText(settings.TerminalBgColor)
	bgEntry.OnChanged = func(v string) { settings.TerminalBgColor = v }

	fgEntry := widget.NewEntry()
	fgEntry.SetText(settings.TerminalFgColor)
	fgEntry.OnChanged = func(v string) { settings.TerminalFgColor = v }

	return container.NewVBox(
		widget.NewLabel("Font"), fontEntry,
		widget.NewLabel("Background color (hex)"), bgEntry,
		widget.NewLabel("Foreground color (hex)"), fgEntry,
	)
}

func (s *SettingsWindow) mcpTab(settings *models.AppSettings) fyne.CanvasObject {
	enableCheck := widget.NewCheck("Enable MCP server", func(v bool) {
		settings.MCPServerEnabled = v
	})
	enableCheck.Checked = settings.MCPServerEnabled

	portEntry := widget.NewEntry()
	portEntry.SetText(itoa(settings.MCPServerPort))
	portEntry.OnChanged = func(v string) {
		// TODO: parse int
	}

	return container.NewVBox(
		enableCheck,
		widget.NewLabel("Port"), portEntry,
	)
}

func (s *SettingsWindow) autopilotTab(settings *models.AppSettings) fyne.CanvasObject {
	enableCheck := widget.NewCheck("Enable Autopilot", func(v bool) {
		settings.Autopilot.Enabled = v
	})
	enableCheck.Checked = settings.Autopilot.Enabled

	providerSelect := widget.NewSelect([]string{"openai", "anthropic", "google"}, func(v string) {
		settings.Autopilot.Provider = models.AutopilotProvider(v)
	})
	providerSelect.SetSelected(string(settings.Autopilot.Provider))

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(settings.Autopilot.APIKey)
	apiKeyEntry.OnChanged = func(v string) { settings.Autopilot.APIKey = v }

	actionSelect := widget.NewSelect([]string{"mark", "ask", "continue", "custom"}, func(v string) {
		settings.Autopilot.Action = models.AutopilotAction(v)
	})
	actionSelect.SetSelected(string(settings.Autopilot.Action))

	return container.NewVBox(
		enableCheck,
		widget.NewLabel("Provider"), providerSelect,
		widget.NewLabel("API key"), apiKeyEntry,
		widget.NewLabel("Action"), actionSelect,
	)
}

func (s *SettingsWindow) appearanceTab(settings *models.AppSettings) fyne.CanvasObject {
	modeSelect := widget.NewSelect([]string{"auto", "system", "light", "dark"}, func(v string) {
		settings.AppearanceMode = models.AppearanceMode(v)
	})
	modeSelect.SetSelected(string(settings.AppearanceMode))

	mermaidSelect := widget.NewSelect([]string{"auto", "light", "dark"}, func(v string) {
		settings.MermaidTheme = v
	})
	mermaidSelect.SetSelected(settings.MermaidTheme)

	return container.NewVBox(
		widget.NewLabel("Appearance mode"), modeSelect,
		widget.NewLabel("Mermaid theme"), mermaidSelect,
	)
}

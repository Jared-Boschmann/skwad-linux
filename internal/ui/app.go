// Package ui contains all Fyne-based UI components for Skwad.
// It assembles the main window, sidebar, terminal panes, git panel,
// markdown preview, file finder, agent sheet, and settings window.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/google/uuid"
	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
	"github.com/Jared-Boschmann/skwad-linux/internal/models"
	"github.com/Jared-Boschmann/skwad-linux/internal/persistence"
	"github.com/Jared-Boschmann/skwad-linux/internal/terminal"
)

const (
	appID    = "com.kochava.skwad"
	appTitle = "Skwad"

	minWidth  = 800
	minHeight = 600

	// sidebarSplitOffset is the initial fraction of the window width
	// occupied by [workspaceBar + sidebar]. ~20% ≈ 160-200px on a 1000px window.
	sidebarSplitOffset = 0.20
)

// App is the top-level Fyne application wrapper.
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	manager *agent.Manager
	coord   *agent.Coordinator
	store   *persistence.Store
	pool    *terminal.Pool

	workspaceBar *WorkspaceBar
	sidebar      *Sidebar
	terminalArea *TerminalArea
	mainSplit    *container.Split
}

// NewApp creates and configures the Fyne app.
func NewApp(mgr *agent.Manager, coord *agent.Coordinator, store *persistence.Store, pool *terminal.Pool) *App {
	a := &App{
		fyneApp: app.NewWithID(appID),
		manager: mgr,
		coord:   coord,
		store:   store,
		pool:    pool,
	}
	a.buildWindow()
	// Spawn sessions for all agents that were persisted from the last run.
	// This runs after buildWindow so that OnAgentChanged is registered.
	for _, ag := range mgr.AllAgents() {
		pool.Spawn(ag)
	}
	return a
}

func (a *App) buildWindow() {
	a.window = a.fyneApp.NewWindow(appTitle)
	a.window.Resize(fyne.NewSize(minWidth, minHeight))
	a.window.SetMaster()

	a.workspaceBar = NewWorkspaceBar(a.manager)
	a.sidebar = NewSidebar(a.manager)
	a.terminalArea = NewTerminalArea(a.manager)

	// Wire new-agent creation: manager.AddAgent + pool.Spawn.
	// Called by the sidebar "New Agent" button AFTER the manager lock is released.
	a.sidebar.OnAddAgent = func(ag *models.Agent) {
		a.manager.AddAgent(ag, nil)
		a.pool.Spawn(ag)
	}
	a.sidebar.OnRemoveAgent = func(id uuid.UUID) {
		a.pool.Kill(id)
		a.manager.RemoveAgent(id)
	}
	a.sidebar.OnRestartAgent = func(id uuid.UUID) {
		a.pool.Restart(id)
	}
	a.sidebar.window = a.window
	a.sidebar.store = a.store

	// Wire change callbacks so the UI stays in sync with manager state.
	// NOTE: these are called while the manager mutex is held, so we must not
	// call any manager method (or pool.Spawn) from inside them.
	a.manager.OnAgentChanged = func(_ uuid.UUID) {
		a.sidebar.Refresh()
		a.terminalArea.Refresh()
		a.workspaceBar.Refresh()
	}
	a.manager.OnWorkspaceChanged = func() {
		a.workspaceBar.Refresh()
		a.sidebar.Refresh()
		a.terminalArea.Refresh()
	}

	// Main layout:
	//   [ workspaceBar (fixed 48px) ][ sidebar ]  |  [ terminal area ]
	// The HSplit drag handle lets the user resize the sidebar.
	leftPanel := container.NewBorder(nil, nil, a.workspaceBar.Widget(), nil, a.sidebar.Widget())
	a.mainSplit = container.NewHSplit(leftPanel, a.terminalArea.Widget())
	a.mainSplit.Offset = sidebarSplitOffset

	a.window.SetContent(a.mainSplit)
	a.setupKeyboardShortcuts()
}

func (a *App) setupKeyboardShortcuts() {
	// fyne.KeyModifierShortcutDefault = Ctrl on Linux/Windows, Cmd on macOS.
	mod := fyne.KeyModifierShortcutDefault

	// Ctrl/Cmd+N: new agent
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: mod},
		func(_ fyne.Shortcut) { a.sidebar.OpenNewAgentSheet() },
	)

	// Ctrl/Cmd+G: toggle git panel
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyG, Modifier: mod},
		func(_ fyne.Shortcut) { a.terminalArea.ToggleGitPanel() },
	)

	// Ctrl/Cmd+\: toggle sidebar
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyBackslash, Modifier: mod},
		func(_ fyne.Shortcut) {
			a.sidebar.Toggle()
			// Rebuild the left panel without the sidebar and re-set content.
			leftPanel := container.NewBorder(nil, nil, a.workspaceBar.Widget(), nil, a.sidebar.Widget())
			a.mainSplit.Leading = leftPanel
			a.mainSplit.Refresh()
		},
	)

	// Ctrl/Cmd+P: file finder (TODO)
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: mod},
		func(_ fyne.Shortcut) { /* TODO: open FileFinder overlay */ },
	)

	// Ctrl/Cmd+]: next agent
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyRightBracket, Modifier: mod},
		func(_ fyne.Shortcut) { a.selectAdjacentAgent(1) },
	)

	// Ctrl/Cmd+[: previous agent
	a.window.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket, Modifier: mod},
		func(_ fyne.Shortcut) { a.selectAdjacentAgent(-1) },
	)
}

// selectAdjacentAgent moves the focused pane to the next or previous agent.
func (a *App) selectAdjacentAgent(delta int) {
	ws := a.manager.ActiveWorkspace()
	if ws == nil {
		return
	}
	agents := a.manager.Agents()
	if len(agents) == 0 {
		return
	}

	var curID uuid.UUID
	if len(ws.ActiveAgentIDs) > ws.FocusedPaneIndex {
		curID = ws.ActiveAgentIDs[ws.FocusedPaneIndex]
	}

	idx := 0
	for i, ag := range agents {
		if ag.ID == curID {
			idx = i
			break
		}
	}
	next := (idx + delta + len(agents)) % len(agents)
	nextID := agents[next].ID

	a.manager.UpdateWorkspace(ws.ID, func(w *models.Workspace) {
		if len(w.ActiveAgentIDs) == 0 {
			w.ActiveAgentIDs = []uuid.UUID{nextID}
		} else {
			w.ActiveAgentIDs[w.FocusedPaneIndex] = nextID
		}
	})
}

// Run starts the Fyne event loop (blocks until window is closed).
func (a *App) Run() {
	a.window.ShowAndRun()
}

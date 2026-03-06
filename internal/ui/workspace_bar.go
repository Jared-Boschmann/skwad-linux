package ui

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
	"github.com/Jared-Boschmann/skwad-linux/internal/models"
)

const workspaceBarWidth float32 = 48

// WorkspaceBar is the vertical strip on the far left showing workspace badges.
type WorkspaceBar struct {
	manager   *agent.Manager
	container *fyne.Container
}

// NewWorkspaceBar creates the workspace bar.
func NewWorkspaceBar(mgr *agent.Manager) *WorkspaceBar {
	wb := &WorkspaceBar{manager: mgr}
	wb.build()
	return wb
}

func (wb *WorkspaceBar) build() {
	wb.container = container.NewVBox(wb.items()...)
}

func (wb *WorkspaceBar) items() []fyne.CanvasObject {
	workspaces := wb.manager.Workspaces()
	activeWS := wb.manager.ActiveWorkspace()

	var items []fyne.CanvasObject
	for _, ws := range workspaces {
		ws := ws // capture
		active := activeWS != nil && ws.ID == activeWS.ID
		badge := wb.makeBadge(ws, active)
		items = append(items, badge)
	}

	addBtn := widget.NewButton("+", func() {
		// TODO: show new workspace dialog
	})
	items = append(items, addBtn)

	return items
}

func (wb *WorkspaceBar) makeBadge(ws *models.Workspace, active bool) fyne.CanvasObject {
	bgColor := parseHexColor(ws.ColorHex)
	if !active {
		// Dim inactive workspaces.
		bgColor = color.NRGBA{R: 70, G: 70, B: 70, A: 255}
	}

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(workspaceBarWidth-8, workspaceBarWidth-8))
	bg.CornerRadius = 6

	label := canvas.NewText(initials(ws.Name), color.White)
	label.TextSize = 13
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	// Status dot: shows worst agent status for this workspace.
	dotColor := workspaceStatusColor(wb.wsAgents(ws))
	dot := canvas.NewCircle(dotColor)
	dot.Resize(fyne.NewSize(6, 6))
	dotContainer := container.NewCenter(dot)

	btn := widget.NewButton("", func() {
		wb.manager.SetActiveWorkspace(ws.ID)
		wb.Refresh()
	})

	return container.NewStack(bg, container.NewCenter(label), dotContainer, btn)
}

// wsAgents returns all agents belonging to the given workspace.
func (wb *WorkspaceBar) wsAgents(ws *models.Workspace) []*models.Agent {
	agents := make([]*models.Agent, 0, len(ws.AgentIDs))
	for _, id := range ws.AgentIDs {
		if a, ok := wb.manager.Agent(id); ok {
			agents = append(agents, a)
		}
	}
	return agents
}

// Refresh rebuilds the workspace bar to reflect current state.
func (wb *WorkspaceBar) Refresh() {
	wb.container.Objects = wb.items()
	wb.container.Refresh()
}

// Widget returns the Fyne widget for embedding in the layout.
func (wb *WorkspaceBar) Widget() fyne.CanvasObject {
	return wb.container
}

func initials(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "?"
	}
	return strings.ToUpper(string(runes[0:1]))
}

// parseHexColor converts an "#RRGGBB" string to color.NRGBA.
// Returns a fallback blue if the string is malformed.
func parseHexColor(hex string) color.NRGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return color.NRGBA{R: 74, G: 144, B: 217, A: 255}
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return color.NRGBA{R: 74, G: 144, B: 217, A: 255}
	}
	return color.NRGBA{
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
		A: 255,
	}
}

// workspaceStatusColor returns the dot color for the worst agent status.
func workspaceStatusColor(agents []*models.Agent) color.NRGBA {
	switch models.WorstStatus(agents) {
	case models.AgentStatusRunning:
		return color.NRGBA{R: 255, G: 165, B: 0, A: 255} // orange
	case models.AgentStatusInput:
		return color.NRGBA{R: 255, G: 59, B: 48, A: 255} // red
	case models.AgentStatusError:
		return color.NRGBA{R: 255, G: 59, B: 48, A: 255} // red
	default:
		return color.NRGBA{R: 100, G: 210, B: 80, A: 180} // green (slightly transparent)
	}
}

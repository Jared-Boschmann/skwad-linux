package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"

	"github.com/google/uuid"
	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
)

// TerminalPane is a single pane slot that hosts one agent's terminal.
//
// The pane renders a placeholder rectangle; the VTE overlay window is
// positioned to match this rectangle's screen coordinates. When the agent
// changes, the overlay is swapped to the new agent's VTE window.
type TerminalPane struct {
	paneIndex int
	manager   *agent.Manager
	agentID   uuid.UUID

	placeholder *canvas.Rectangle
	label       *widget.Label
	outer       *fyne.Container
}

// NewTerminalPane creates a pane for the given pane index.
func NewTerminalPane(paneIndex int, mgr *agent.Manager) *TerminalPane {
	tp := &TerminalPane{
		paneIndex: paneIndex,
		manager:   mgr,
	}
	tp.build()
	return tp
}

func (tp *TerminalPane) build() {
	tp.placeholder = canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	tp.placeholder.SetMinSize(fyne.NewSize(200, 100))

	tp.label = widget.NewLabel("No agent selected")

	tp.outer = container.NewStack(tp.placeholder, container.NewCenter(tp.label))
}

// SetAgentID assigns an agent to this pane and updates the display.
func (tp *TerminalPane) SetAgentID(id uuid.UUID) {
	tp.agentID = id
	a, ok := tp.manager.Agent(id)
	if ok {
		tp.label.SetText(a.Name)
	} else {
		tp.label.SetText("No agent selected")
	}
	tp.placeholder.Refresh()
	// TODO: show/hide VTE overlay window for this agent
}

// Widget returns the Fyne canvas object for this pane.
func (tp *TerminalPane) Widget() fyne.CanvasObject {
	return tp.outer
}

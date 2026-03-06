package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
	"github.com/Jared-Boschmann/skwad-linux/internal/models"
)

// TerminalArea manages the main content area with split-pane layout.
//
// NOTE on VTE embedding: because VTE widgets are native GTK widgets, they
// cannot be placed directly into a Fyne container. Instead, each TerminalPane
// holds a placeholder Fyne widget that tracks its position/size; the actual
// VTE window is a sibling X11 window kept in sync with those bounds.
// See internal/terminal/vte.go for the embedding strategy details.
type TerminalArea struct {
	manager   *agent.Manager
	container *fyne.Container

	gitPanel      *GitPanel
	markdownPanel *MarkdownPanel

	showGit      bool
	showMarkdown bool
}

// NewTerminalArea creates the terminal area.
func NewTerminalArea(mgr *agent.Manager) *TerminalArea {
	ta := &TerminalArea{
		manager:       mgr,
		gitPanel:      NewGitPanel(mgr),
		markdownPanel: NewMarkdownPanel(),
	}
	ta.build()
	return ta
}

func (ta *TerminalArea) build() {
	ta.container = container.NewStack(ta.panes())
}

func (ta *TerminalArea) panes() fyne.CanvasObject {
	ws := ta.manager.ActiveWorkspace()
	if ws == nil {
		return container.NewStack()
	}

	switch ws.LayoutMode {
	case models.LayoutModeSplitVertical:
		return ta.splitVertical(ws)
	case models.LayoutModeSplitHorizontal:
		return ta.splitHorizontal(ws)
	case models.LayoutModeThreePane:
		return ta.threePane(ws)
	case models.LayoutModeGridFourPane:
		return ta.gridFourPane(ws)
	default:
		return ta.singlePane(ws)
	}
}

func (ta *TerminalArea) singlePane(ws *models.Workspace) fyne.CanvasObject {
	pane := NewTerminalPane(0, ta.manager)
	if len(ws.ActiveAgentIDs) > 0 {
		pane.SetAgentID(ws.ActiveAgentIDs[0])
	}
	return pane.Widget()
}

func (ta *TerminalArea) splitVertical(ws *models.Workspace) fyne.CanvasObject {
	left := NewTerminalPane(0, ta.manager)
	right := NewTerminalPane(1, ta.manager)
	if len(ws.ActiveAgentIDs) > 0 {
		left.SetAgentID(ws.ActiveAgentIDs[0])
	}
	if len(ws.ActiveAgentIDs) > 1 {
		right.SetAgentID(ws.ActiveAgentIDs[1])
	}
	split := container.NewHSplit(left.Widget(), right.Widget())
	split.Offset = ws.SplitRatio
	return split
}

func (ta *TerminalArea) splitHorizontal(ws *models.Workspace) fyne.CanvasObject {
	top := NewTerminalPane(0, ta.manager)
	bottom := NewTerminalPane(1, ta.manager)
	if len(ws.ActiveAgentIDs) > 0 {
		top.SetAgentID(ws.ActiveAgentIDs[0])
	}
	if len(ws.ActiveAgentIDs) > 1 {
		bottom.SetAgentID(ws.ActiveAgentIDs[1])
	}
	split := container.NewVSplit(top.Widget(), bottom.Widget())
	split.Offset = ws.SplitRatio
	return split
}

func (ta *TerminalArea) threePane(ws *models.Workspace) fyne.CanvasObject {
	left := NewTerminalPane(0, ta.manager)
	rightTop := NewTerminalPane(1, ta.manager)
	rightBottom := NewTerminalPane(2, ta.manager)
	if len(ws.ActiveAgentIDs) > 0 {
		left.SetAgentID(ws.ActiveAgentIDs[0])
	}
	if len(ws.ActiveAgentIDs) > 1 {
		rightTop.SetAgentID(ws.ActiveAgentIDs[1])
	}
	if len(ws.ActiveAgentIDs) > 2 {
		rightBottom.SetAgentID(ws.ActiveAgentIDs[2])
	}
	rightSplit := container.NewVSplit(rightTop.Widget(), rightBottom.Widget())
	rightSplit.Offset = ws.SplitRatioSecondary
	mainSplit := container.NewHSplit(left.Widget(), rightSplit)
	mainSplit.Offset = ws.SplitRatio
	return mainSplit
}

func (ta *TerminalArea) gridFourPane(ws *models.Workspace) fyne.CanvasObject {
	panes := make([]*TerminalPane, 4)
	for i := range panes {
		panes[i] = NewTerminalPane(i, ta.manager)
		if i < len(ws.ActiveAgentIDs) {
			panes[i].SetAgentID(ws.ActiveAgentIDs[i])
		}
	}
	topSplit := container.NewHSplit(panes[0].Widget(), panes[1].Widget())
	topSplit.Offset = ws.SplitRatio
	botSplit := container.NewHSplit(panes[2].Widget(), panes[3].Widget())
	botSplit.Offset = ws.SplitRatio
	mainSplit := container.NewVSplit(topSplit, botSplit)
	mainSplit.Offset = ws.SplitRatioSecondary
	return mainSplit
}

// Refresh rebuilds the layout.
func (ta *TerminalArea) Refresh() {
	ta.container.Objects = []fyne.CanvasObject{ta.panes()}
	ta.container.Refresh()
}

// Widget returns the terminal area widget.
func (ta *TerminalArea) Widget() fyne.CanvasObject {
	return ta.container
}

// ToggleGitPanel shows or hides the git panel.
func (ta *TerminalArea) ToggleGitPanel() {
	ta.showGit = !ta.showGit
	ta.Refresh()
}

// ToggleMarkdownPanel shows or hides the markdown panel.
func (ta *TerminalArea) ToggleMarkdownPanel() {
	ta.showMarkdown = !ta.showMarkdown
	ta.Refresh()
}

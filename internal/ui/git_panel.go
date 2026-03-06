package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/google/uuid"
	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
	"github.com/Jared-Boschmann/skwad-linux/internal/git"
)

// GitPanel is the sliding panel showing git status for the active agent's folder.
type GitPanel struct {
	manager *agent.Manager
	repo    *git.Repository
	agentID uuid.UUID

	branchLabel  *widget.Label
	aheadBehind  *widget.Label
	fileList     *widget.List
	commitMsg    *widget.Entry
	commitBtn    *widget.Button
	stageAllBtn  *widget.Button

	files   []git.FileStatus
	outer   *fyne.Container
}

// NewGitPanel creates the git panel.
func NewGitPanel(mgr *agent.Manager) *GitPanel {
	gp := &GitPanel{manager: mgr}
	gp.build()
	return gp
}

func (gp *GitPanel) build() {
	gp.branchLabel = widget.NewLabel("No branch")
	gp.aheadBehind = widget.NewLabel("")

	gp.fileList = widget.NewList(
		func() int { return len(gp.files) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""), // status badges
				widget.NewLabel(""), // path
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(gp.files) {
				return
			}
			f := gp.files[id]
			row := obj.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(string(f.Staged) + string(f.Unstaged))
			row.Objects[1].(*widget.Label).SetText(f.Path)
		},
	)
	gp.fileList.OnSelected = func(id widget.ListItemID) {
		if id >= len(gp.files) {
			return
		}
		// TODO: open diff view for gp.files[id]
	}

	gp.stageAllBtn = widget.NewButton("Stage All", func() {
		if gp.repo != nil {
			_ = gp.repo.StageAll()
			gp.Reload()
		}
	})

	gp.commitMsg = widget.NewMultiLineEntry()
	gp.commitMsg.SetPlaceHolder("Commit message…")
	gp.commitMsg.SetMinRowsVisible(2)

	gp.commitBtn = widget.NewButton("Commit", func() {
		if gp.repo != nil && gp.commitMsg.Text != "" {
			_ = gp.repo.Commit(gp.commitMsg.Text)
			gp.commitMsg.SetText("")
			gp.Reload()
		}
	})

	header := container.NewHBox(gp.branchLabel, gp.aheadBehind, gp.stageAllBtn)
	footer := container.NewBorder(nil, nil, nil, gp.commitBtn, gp.commitMsg)

	gp.outer = container.NewBorder(header, footer, nil, nil, gp.fileList)
}

// SetAgent updates the panel to show git state for the given agent's folder.
func (gp *GitPanel) SetAgent(id uuid.UUID) {
	gp.agentID = id
	a, ok := gp.manager.Agent(id)
	if !ok {
		return
	}
	gp.repo = git.NewRepository(a.Folder)
	gp.Reload()
}

// Reload refreshes branch info and file status from git.
func (gp *GitPanel) Reload() {
	if gp.repo == nil {
		return
	}

	branch, err := gp.repo.Branch()
	if err == nil {
		gp.branchLabel.SetText(branch.Name)
		if branch.Upstream != "" {
			gp.aheadBehind.SetText(
				"↑" + itoa(branch.Ahead) + " ↓" + itoa(branch.Behind),
			)
		}
	}

	files, err := gp.repo.Status()
	if err == nil {
		gp.files = files
		gp.fileList.Refresh()
	}
}

// Widget returns the git panel widget.
func (gp *GitPanel) Widget() fyne.CanvasObject { return gp.outer }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

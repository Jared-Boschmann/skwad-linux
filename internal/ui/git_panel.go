package ui

import (
	"strings"

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

	branchLabel *widget.Label
	aheadBehind *widget.Label
	fileList    *widget.List
	diffView    *widget.RichText
	commitMsg   *widget.Entry
	commitBtn   *widget.Button
	stageAllBtn *widget.Button
	stageBtn    *widget.Button
	unstageBtn  *widget.Button
	discardBtn  *widget.Button

	files       []git.FileStatus
	selectedIdx int
	outer       *fyne.Container
}

// NewGitPanel creates the git panel.
func NewGitPanel(mgr *agent.Manager) *GitPanel {
	gp := &GitPanel{manager: mgr, selectedIdx: -1}
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
				widget.NewLabel(""),  // XY status
				widget.NewLabel(""), // path
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(gp.files) {
				return
			}
			f := gp.files[id]
			row := obj.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(fileStatusIcon(f))
			row.Objects[1].(*widget.Label).SetText(f.Path)
		},
	)
	gp.fileList.OnSelected = func(id widget.ListItemID) {
		if id >= len(gp.files) || gp.repo == nil {
			return
		}
		gp.selectedIdx = id
		gp.updateActionButtons()
		gp.loadDiff(gp.files[id])
	}

	gp.diffView = widget.NewRichText()
	diffScroll := container.NewScroll(gp.diffView)

	// Per-file action buttons (enabled only when a file is selected).
	gp.stageBtn = widget.NewButton("Stage", func() {
		if gp.selectedIdx >= 0 && gp.selectedIdx < len(gp.files) && gp.repo != nil {
			_ = gp.repo.Stage(gp.files[gp.selectedIdx].Path)
			gp.Reload()
		}
	})
	gp.unstageBtn = widget.NewButton("Unstage", func() {
		if gp.selectedIdx >= 0 && gp.selectedIdx < len(gp.files) && gp.repo != nil {
			_ = gp.repo.Unstage(gp.files[gp.selectedIdx].Path)
			gp.Reload()
		}
	})
	gp.discardBtn = widget.NewButton("Discard", func() {
		if gp.selectedIdx >= 0 && gp.selectedIdx < len(gp.files) && gp.repo != nil {
			_ = gp.repo.Discard(gp.files[gp.selectedIdx].Path)
			gp.Reload()
		}
	})
	gp.stageBtn.Disable()
	gp.unstageBtn.Disable()
	gp.discardBtn.Disable()

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
			gp.selectedIdx = -1
			gp.Reload()
		}
	})

	header := container.NewHBox(gp.branchLabel, gp.aheadBehind, gp.stageAllBtn)
	actions := container.NewHBox(gp.stageBtn, gp.unstageBtn, gp.discardBtn)

	// Split: file list (top) | diff view (bottom)
	listAndDiff := container.NewVSplit(gp.fileList, diffScroll)
	listAndDiff.Offset = 0.5

	footer := container.NewBorder(nil, nil, nil, gp.commitBtn, gp.commitMsg)

	gp.outer = container.NewBorder(header, container.NewVBox(actions, footer), nil, nil, listAndDiff)
}

// updateActionButtons enables/disables stage/unstage/discard based on file status.
func (gp *GitPanel) updateActionButtons() {
	if gp.selectedIdx < 0 || gp.selectedIdx >= len(gp.files) {
		gp.stageBtn.Disable()
		gp.unstageBtn.Disable()
		gp.discardBtn.Disable()
		return
	}
	f := gp.files[gp.selectedIdx]

	// Can stage if file has unstaged changes or is untracked.
	if f.Unstaged != git.FileStatusCode(" ") {
		gp.stageBtn.Enable()
	} else {
		gp.stageBtn.Disable()
	}
	// Can unstage if file has staged changes.
	if f.Staged != git.FileStatusCode(" ") && f.Staged != git.FileStatusCode("?") {
		gp.unstageBtn.Enable()
	} else {
		gp.unstageBtn.Disable()
	}
	// Can discard tracked files with unstaged changes.
	if f.Unstaged != git.FileStatusCode(" ") && f.Unstaged != git.FileStatusCode("?") {
		gp.discardBtn.Enable()
	} else {
		gp.discardBtn.Disable()
	}
}

// loadDiff loads and displays the diff for the selected file.
func (gp *GitPanel) loadDiff(f git.FileStatus) {
	if gp.repo == nil {
		return
	}
	// Show staged diff if the file is staged, otherwise unstaged.
	staged := f.Staged != git.FileStatusCode(" ") && f.Staged != git.FileStatusCode("?")
	lines, err := gp.repo.Diff(f.Path, staged)
	if err != nil || len(lines) == 0 {
		gp.diffView.ParseMarkdown("*No diff available*")
		return
	}

	var sb strings.Builder
	sb.WriteString("```diff\n")
	for _, l := range lines {
		sb.WriteString(l.Content)
		sb.WriteByte('\n')
	}
	sb.WriteString("```\n")
	gp.diffView.ParseMarkdown(sb.String())
	gp.diffView.Refresh()
}

// SetAgent updates the panel to show git state for the given agent's folder.
func (gp *GitPanel) SetAgent(id uuid.UUID) {
	gp.agentID = id
	a, ok := gp.manager.Agent(id)
	if !ok {
		return
	}
	gp.repo = git.NewRepository(a.Folder)
	gp.selectedIdx = -1
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
			gp.aheadBehind.SetText("↑" + itoa(branch.Ahead) + " ↓" + itoa(branch.Behind))
		} else {
			gp.aheadBehind.SetText("")
		}
	}

	files, err := gp.repo.Status()
	if err == nil {
		gp.files = files
		gp.fileList.Refresh()
	}
	gp.updateActionButtons()
}

// Widget returns the git panel widget.
func (gp *GitPanel) Widget() fyne.CanvasObject { return gp.outer }

// fileStatusIcon returns a short text indicator for the file's XY status codes.
func fileStatusIcon(f git.FileStatus) string {
	staged := string(f.Staged)
	unstaged := string(f.Unstaged)
	if unstaged == "?" {
		return "??" // untracked
	}
	s := staged
	if s == " " {
		s = "·"
	}
	u := unstaged
	if u == " " {
		u = "·"
	}
	return s + u
}

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

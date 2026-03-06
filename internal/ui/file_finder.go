package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Jared-Boschmann/skwad-linux/internal/git"
	"github.com/Jared-Boschmann/skwad-linux/internal/search"
)

const maxFinderResults = 50

// FileFinder is a keyboard-invocable overlay for fuzzy file search.
type FileFinder struct {
	folder  string
	files   []string // relative paths, indexed from folder
	results []search.Result

	input *widget.Entry
	list  *widget.List
	outer *fyne.Container

	// OnSelect is called with the full absolute path of the selected file.
	OnSelect func(path string)
	// OnClose is called when the user dismisses the finder.
	OnClose func()
}

// NewFileFinder creates a file finder for the given folder.
func NewFileFinder(folder string) *FileFinder {
	ff := &FileFinder{folder: folder}
	ff.build()
	return ff
}

func (ff *FileFinder) build() {
	ff.input = widget.NewEntry()
	ff.input.SetPlaceHolder("Search files…")
	ff.input.OnChanged = ff.onQueryChanged

	ff.list = widget.NewList(
		func() int { return len(ff.results) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(ff.results) {
				return
			}
			obj.(*widget.Label).SetText(ff.results[id].Path)
		},
	)
	ff.list.OnSelected = func(id widget.ListItemID) {
		if id >= len(ff.results) {
			return
		}
		absPath := filepath.Join(ff.folder, ff.results[id].Path)
		if ff.OnSelect != nil {
			ff.OnSelect(absPath)
		}
	}

	closeBtn := widget.NewButton("×", func() {
		if ff.OnClose != nil {
			ff.OnClose()
		}
	})

	ff.outer = container.NewBorder(
		container.NewBorder(nil, nil, nil, closeBtn, ff.input),
		nil, nil, nil,
		ff.list,
	)
}

func (ff *FileFinder) onQueryChanged(query string) {
	if query == "" {
		ff.results = nil
		ff.list.Refresh()
		return
	}
	ff.results = search.FuzzySearch(ff.files, query, maxFinderResults)
	ff.list.Refresh()
}

// IndexFolder walks the folder and builds the file list, respecting ExcludedDirs.
func (ff *FileFinder) IndexFolder(folder string) {
	ff.folder = folder
	var files []string
	_ = filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if git.IsExcluded(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(folder, path)
		if err != nil {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	ff.files = files
}

// Widget returns the overlay widget.
func (ff *FileFinder) Widget() fyne.CanvasObject { return ff.outer }

// FocusInput focuses the search input on the given canvas.
func (ff *FileFinder) FocusInput(c fyne.Canvas) { c.Focus(ff.input) }

// openFileExternal opens a file using the platform's default application.
func openFileExternal(path string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	_ = exec.Command(cmd, path).Start()
}

// runDetached launches an external application with the given argument,
// detached from the current process (fire-and-forget).
func runDetached(app, arg string) error {
	return exec.Command(app, arg).Start()
}

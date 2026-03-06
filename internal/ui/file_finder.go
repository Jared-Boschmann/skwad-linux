package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Jared-Boschmann/skwad-linux/internal/search"
)

const maxFinderResults = 50

// FileFinder is a keyboard-invocable overlay for fuzzy file search.
type FileFinder struct {
	folder  string
	files   []string // indexed file list
	results []search.Result

	input    *widget.Entry
	list     *widget.List
	outer    *fyne.Container

	OnSelect func(path string)
	OnClose  func()
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
		if ff.OnSelect != nil {
			ff.OnSelect(ff.results[id].Path)
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

// Index re-scans the folder and builds the file list.
func (ff *FileFinder) Index(files []string) {
	ff.files = files
}

// Widget returns the overlay widget.
func (ff *FileFinder) Widget() fyne.CanvasObject { return ff.outer }

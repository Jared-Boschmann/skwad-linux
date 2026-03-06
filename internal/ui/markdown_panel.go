package ui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/yuin/goldmark"
)

// MarkdownPanel renders a markdown file in a scrollable panel.
type MarkdownPanel struct {
	filePath  string
	history   []string
	historyPos int

	content  *widget.RichText
	title    *widget.Label
	backBtn  *widget.Button
	fwdBtn   *widget.Button
	closeBtn *widget.Button
	outer    *fyne.Container

	OnClose func()
}

// NewMarkdownPanel creates a new markdown preview panel.
func NewMarkdownPanel() *MarkdownPanel {
	mp := &MarkdownPanel{}
	mp.build()
	return mp
}

func (mp *MarkdownPanel) build() {
	mp.content = widget.NewRichText()
	mp.title = widget.NewLabel("")

	mp.backBtn = widget.NewButton("←", func() {
		if mp.historyPos > 0 {
			mp.historyPos--
			mp.loadFile(mp.history[mp.historyPos])
		}
	})
	mp.fwdBtn = widget.NewButton("→", func() {
		if mp.historyPos < len(mp.history)-1 {
			mp.historyPos++
			mp.loadFile(mp.history[mp.historyPos])
		}
	})
	mp.closeBtn = widget.NewButton("×", func() {
		if mp.OnClose != nil {
			mp.OnClose()
		}
	})

	toolbar := container.NewHBox(mp.backBtn, mp.fwdBtn, mp.title, mp.closeBtn)
	scroll := container.NewScroll(mp.content)
	mp.outer = container.NewBorder(toolbar, nil, nil, nil, scroll)
}

// ShowFile opens a markdown file and adds it to the history.
func (mp *MarkdownPanel) ShowFile(path string) {
	// Truncate forward history on new navigation.
	if mp.historyPos < len(mp.history)-1 {
		mp.history = mp.history[:mp.historyPos+1]
	}
	mp.history = append(mp.history, path)
	mp.historyPos = len(mp.history) - 1
	mp.loadFile(path)
}

func (mp *MarkdownPanel) loadFile(path string) {
	mp.filePath = path
	mp.title.SetText(path)

	data, err := os.ReadFile(path)
	if err != nil {
		mp.content.ParseMarkdown("*Error loading file: " + err.Error() + "*")
		return
	}

	// Use goldmark to convert markdown to a plain text representation.
	// For richer rendering, convert to HTML and embed in a WebView.
	// TODO: swap to HTML+WebView for full styling support.
	var buf []byte
	md := goldmark.New()
	_ = md

	// Fallback: display raw markdown text via Fyne RichText parser.
	_ = buf
	mp.content.ParseMarkdown(string(data))
	mp.content.Refresh()
}

// Widget returns the panel widget.
func (mp *MarkdownPanel) Widget() fyne.CanvasObject { return mp.outer }

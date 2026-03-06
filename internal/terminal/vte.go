//go:build linux

package terminal

// VTE terminal backend using libvte via CGo.
//
// IMPORTANT — Embedding strategy:
//   Fyne uses its own OpenGL rendering pipeline and does not natively support
//   embedding GTK widgets. To integrate VTE with Fyne on Linux we use one of
//   two approaches:
//
//   1. XEmbed (X11):  Create a GtkSocket in Fyne's native window (obtained via
//      fyne.Window.Canvas().(desktop.Canvas).Content().NativeWindow()), then
//      embed a VTE GtkPlug into it. Requires cgo + libvte-2.91 + libgtk-3.
//
//   2. Overlay window:  Spawn a borderless child X11 window containing the VTE
//      terminal, position it over the Fyne terminal pane area, and keep it in
//      sync with layout changes. Simpler to implement but requires manual
//      coordinate mapping.
//
//   The stub below defines the Go interface; the full CGo implementation lives
//   in vte_impl.c / vte_impl.h (to be created during the implementation phase).
//
// Build tags: this file is only compiled on Linux.
// CGo flags (add to build/Makefile):
//   pkg-config --cflags --libs vte-2.91 gtk+-3.0

// #cgo pkg-config: vte-2.91 gtk+-3.0
// #include "vte_impl.h"
import "C"

// VTETerminal wraps a native VTE terminal widget.
// This is a stub — the CGo implementation will be added during implementation.
type VTETerminal struct {
	// handle is the opaque pointer to the native VTE widget.
	// handle C.SkwadVTE

	onOutput     func()
	onTitleChange func(string)
	onUserInput  func(uint16)
}

// NewVTETerminal creates a new VTE terminal widget and returns it.
// The returned terminal must be embedded into the window before use.
func NewVTETerminal(command string, env []string) (*VTETerminal, error) {
	// TODO: call C.skwad_vte_new(command, env) and store handle
	return &VTETerminal{}, nil
}

func (v *VTETerminal) SendText(text string) {
	// TODO: C.skwad_vte_feed_child(v.handle, text)
}

func (v *VTETerminal) SendReturn() {
	v.SendText("\r")
}

func (v *VTETerminal) InjectText(text string) {
	v.SendText(text)
	v.SendReturn()
}

func (v *VTETerminal) Resize(cols, rows int) {
	// TODO: C.skwad_vte_set_size(v.handle, cols, rows)
}

func (v *VTETerminal) Focus() {
	// TODO: C.skwad_vte_grab_focus(v.handle)
}

func (v *VTETerminal) Kill() {
	// TODO: C.skwad_vte_kill(v.handle)
}

func (v *VTETerminal) SetOnOutput(f func()) {
	v.onOutput = f
}

func (v *VTETerminal) SetOnTitleChange(f func(string)) {
	v.onTitleChange = f
}

func (v *VTETerminal) SetOnUserInput(f func(uint16)) {
	v.onUserInput = f
}

//export goOnVTEOutput
func goOnVTEOutput(ptr C.uintptr_t) {
	// TODO: retrieve VTETerminal by ptr, call v.onOutput()
}

//export goOnVTETitleChange
func goOnVTETitleChange(ptr C.uintptr_t, title *C.char) {
	// TODO: retrieve VTETerminal by ptr, call v.onTitleChange(C.GoString(title))
}

//export goOnVTEKeyPress
func goOnVTEKeyPress(ptr C.uintptr_t, keyval C.uint) {
	// TODO: map GDK keyval to a uint16 keyCode, call v.onUserInput(keyCode)
}

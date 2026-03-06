//go:build !linux

package terminal

// VTETerminal is a no-op stub on non-Linux platforms.
// The real implementation (using libvte via CGo) lives in vte.go.
type VTETerminal struct {
	onOutput      func()
	onTitleChange func(string)
	onUserInput   func(uint16)
}

func NewVTETerminal(command string, env []string) (*VTETerminal, error) {
	return &VTETerminal{}, nil
}

func (v *VTETerminal) SendText(text string)            {}
func (v *VTETerminal) SendReturn()                     {}
func (v *VTETerminal) InjectText(text string)          {}
func (v *VTETerminal) Resize(cols, rows int)           {}
func (v *VTETerminal) Focus()                          {}
func (v *VTETerminal) Kill()                           {}
func (v *VTETerminal) SetOnOutput(f func())            { v.onOutput = f }
func (v *VTETerminal) SetOnTitleChange(f func(string)) { v.onTitleChange = f }
func (v *VTETerminal) SetOnUserInput(f func(uint16))   { v.onUserInput = f }

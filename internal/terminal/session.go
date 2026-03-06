package terminal

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/creack/pty"
)

// Session manages a single PTY process (one per agent).
// It is safe for concurrent use.
type Session struct {
	mu      sync.Mutex
	ptmx    *os.File // PTY master
	cmd     *exec.Cmd
	stopped bool

	OnOutput      func(data []byte)
	OnTitleChange func(title string)
	OnExit        func(exitCode int)
}

// NewSession spawns a shell command in a PTY and returns the running session.
// command is the full shell command string (passed to $SHELL -c).
// env is extra environment variables merged with the current environment.
func NewSession(command string, env []string) (*Session, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell, "-c", command)
	cmd.Env = append(os.Environ(), env...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	s := &Session{
		ptmx: ptmx,
		cmd:  cmd,
	}

	go s.readLoop()
	go s.waitLoop()

	return s, nil
}

// SendText writes text to the PTY input without a newline.
func (s *Session) SendText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return
	}
	_, _ = io.WriteString(s.ptmx, text)
}

// SendReturn sends a carriage return.
func (s *Session) SendReturn() {
	s.SendText("\r")
}

// InjectText sends text followed by a carriage return.
func (s *Session) InjectText(text string) {
	s.SendText(text)
	s.SendReturn()
}

// Resize informs the PTY of the new terminal dimensions.
func (s *Session) Resize(cols, rows uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped || s.ptmx == nil {
		return
	}
	_ = pty.Setsize(s.ptmx, &pty.Winsize{Cols: cols, Rows: rows})
}

// Kill terminates the process and closes the PTY.
func (s *Session) Kill() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Signal(syscall.SIGTERM)
	}
	if s.ptmx != nil {
		_ = s.ptmx.Close()
		s.ptmx = nil
	}
}

// IsRunning reports whether the underlying process is still alive.
func (s *Session) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.stopped
}

func (s *Session) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			// Check for OSC title sequence before calling OnOutput.
			if title, ok := extractTitle(chunk); ok && s.OnTitleChange != nil {
				s.OnTitleChange(title)
			}
			if s.OnOutput != nil {
				s.OnOutput(chunk)
			}
		}
		if err != nil {
			return
		}
	}
}

func (s *Session) waitLoop() {
	code := 0
	if err := s.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				code = status.ExitStatus()
			}
		}
	}
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()

	if s.OnExit != nil {
		s.OnExit(code)
	}
}

// extractTitle scans data for an OSC 0 or OSC 2 title escape sequence.
// Format: ESC ] 0 ; <title> BEL  or  ESC ] 2 ; <title> BEL
func extractTitle(data []byte) (string, bool) {
	const (
		esc = 0x1b
		bel = 0x07
	)
	for i := 0; i < len(data)-3; i++ {
		if data[i] != esc || data[i+1] != ']' {
			continue
		}
		if data[i+2] != '0' && data[i+2] != '2' {
			continue
		}
		if i+3 >= len(data) || data[i+3] != ';' {
			continue
		}
		end := bytes.IndexByte(data[i+4:], bel)
		if end < 0 {
			// Also accept ST (ESC \) terminator.
			end = bytes.Index(data[i+4:], []byte{esc, '\\'})
		}
		if end < 0 {
			continue
		}
		raw := data[i+4 : i+4+end]
		if !utf8.Valid(raw) {
			continue
		}
		return string(raw), true
	}
	return "", false
}

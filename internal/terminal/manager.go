package terminal

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const registrationDelay = 3 * time.Second

// Terminal is the interface that any terminal backend must implement.
type Terminal interface {
	// SendText sends text to the terminal without a trailing newline.
	SendText(text string)
	// SendReturn sends a carriage return.
	SendReturn()
	// InjectText sends text followed by a return.
	InjectText(text string)
	// Resize informs the terminal of a new character size.
	Resize(cols, rows int)
	// Focus requests keyboard focus for this terminal.
	Focus()
	// Kill terminates the running process.
	Kill()
	// SetOnOutput sets a callback invoked on any terminal output.
	SetOnOutput(func())
	// SetOnTitleChange sets a callback invoked when the terminal title changes.
	SetOnTitleChange(func(title string))
	// SetOnUserInput sets a callback invoked on user keypress with raw keycode.
	SetOnUserInput(func(keyCode uint16))
}

// Manager keeps all terminal instances alive and routes focus.
// Terminals are never destroyed when switching agents — only shown/hidden.
type Manager struct {
	mu        sync.RWMutex
	terminals map[uuid.UUID]Terminal

	OnRegistrationReady func(agentID uuid.UUID)
}

// NewManager creates an empty terminal manager.
func NewManager() *Manager {
	return &Manager{terminals: make(map[uuid.UUID]Terminal)}
}

// Register stores a terminal for the given agent.
func (m *Manager) Register(agentID uuid.UUID, t Terminal) {
	m.mu.Lock()
	m.terminals[agentID] = t
	m.mu.Unlock()

	// Schedule registration prompt injection after delay.
	if m.OnRegistrationReady != nil {
		go func() {
			time.Sleep(registrationDelay)
			m.OnRegistrationReady(agentID)
		}()
	}
}

// Remove removes a terminal from the manager.
func (m *Manager) Remove(agentID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.terminals[agentID]; ok {
		t.Kill()
		delete(m.terminals, agentID)
	}
}

// Get returns the terminal for the given agent.
func (m *Manager) Get(agentID uuid.UUID) (Terminal, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.terminals[agentID]
	return t, ok
}

// SendText sends text to the given agent's terminal.
func (m *Manager) SendText(agentID uuid.UUID, text string) {
	if t, ok := m.Get(agentID); ok {
		t.SendText(text)
	}
}

// InjectText sends text + return to the given agent's terminal.
func (m *Manager) InjectText(agentID uuid.UUID, text string) {
	if t, ok := m.Get(agentID); ok {
		t.InjectText(text)
	}
}

// Focus gives keyboard focus to the given agent's terminal.
func (m *Manager) Focus(agentID uuid.UUID) {
	if t, ok := m.Get(agentID); ok {
		t.Focus()
	}
}

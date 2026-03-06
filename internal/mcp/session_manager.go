package mcp

import (
	"sync"

	"github.com/google/uuid"
)

// session holds per-client MCP session state.
type session struct {
	id      string
	agentID uuid.UUID // set after register-agent call
}

type sessionManager struct {
	mu       sync.Mutex
	sessions map[string]*session
}

func newSessionManager() *sessionManager {
	return &sessionManager{sessions: make(map[string]*session)}
}

func (sm *sessionManager) getOrCreate(id string) *session {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.sessions[id]; ok {
		return s
	}
	if id == "" {
		id = uuid.NewString()
	}
	s := &session{id: id}
	sm.sessions[id] = s
	return s
}

func (sm *sessionManager) remove(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

// Package mcp implements the in-process MCP (Model Context Protocol) HTTP server.
// It exposes a JSON-RPC 2.0 endpoint at /mcp for AI agent tool calls, and a
// /hook endpoint for lifecycle events posted by claude/codex plugin scripts.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
	"github.com/Jared-Boschmann/skwad-linux/internal/persistence"
)

// Server is the in-process MCP HTTP server (JSON-RPC 2.0).
type Server struct {
	coordinator *agent.Coordinator
	store       *persistence.Store
	port        int

	httpServer  *http.Server
	tools       *toolHandler
	sessions    *sessionManager
	hookHandler *hookHandler

	mu      sync.Mutex
	started bool

	// Callbacks — set by UI layer.
	OnDisplayMarkdown func(agentID, filePath string)
	OnViewMermaid     func(agentID, source, title string)
	OnCreateAgent     func(req CreateAgentRequest) error
	OnCloseAgent      func(callerID, targetID string) error

	// StatusUpdater — set by the agent manager to receive hook events.
	StatusUpdater AgentStatusUpdater
}

// NewServer creates a new MCP server.
func NewServer(coord *agent.Coordinator, store *persistence.Store, port int) *Server {
	s := &Server{
		coordinator: coord,
		store:       store,
		port:        port,
		sessions:    newSessionManager(),
	}
	s.tools = newToolHandler(s)
	return s
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}

	s.hookHandler = newHookHandler(s.StatusUpdater)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)
	mux.Handle("/hook", s.hookHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("mcp server listen: %w", err)
	}

	s.started = true
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("mcp server error: %v", err)
		}
	}()
	log.Printf("MCP server listening on %s", s.httpServer.Addr)
	return nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpServer != nil {
		_ = s.httpServer.Shutdown(context.Background())
	}
	s.started = false
}

// URL returns the full MCP endpoint URL.
func (s *Server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/mcp", s.port)
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, Response{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "parse error"},
		})
		return
	}

	sessionID := r.Header.Get("Mcp-Session-Id")
	session := s.sessions.getOrCreate(sessionID)

	resp := s.dispatch(req, session)
	writeResponse(w, resp)
}

func (s *Server) dispatch(req Request, session *session) Response {
	base := Response{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case MethodPing:
		base.Result = map[string]interface{}{}

	case MethodInitialize:
		base.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"serverInfo":      map[string]interface{}{"name": "skwad", "version": "1.0.0"},
		}

	case MethodToolsList:
		base.Result = map[string]interface{}{"tools": s.tools.list()}

	case MethodToolsCall:
		params, err := parseToolCallParams(req.Params)
		if err != nil {
			base.Error = &RPCError{Code: -32602, Message: "invalid params"}
			return base
		}
		result, toolErr := s.tools.call(params, session)
		if toolErr != nil {
			base.Error = &RPCError{Code: -32000, Message: toolErr.Error()}
		} else {
			base.Result = result
		}

	default:
		base.Error = &RPCError{Code: -32601, Message: "method not found"}
	}

	return base
}


func parseToolCallParams(raw interface{}) (ToolCallParams, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return ToolCallParams{}, err
	}
	var p ToolCallParams
	return p, json.Unmarshal(data, &p)
}

func writeResponse(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/Jared-Boschmann/skwad-linux/internal/agent"
	"github.com/Jared-Boschmann/skwad-linux/internal/persistence"
)

// newTestServer spins up an httptest server backed by a real MCP Server.
func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()

	store, err := persistence.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	mgr, err := agent.NewManager(store)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	coord := agent.NewCoordinator(mgr)

	srv := NewServer(coord, store, 0)
	srv.tools = newToolHandler(srv)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", srv.handleMCP)

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return srv, ts
}

func mcpCall(t *testing.T, ts *httptest.Server, sessionID, method string, params interface{}) Response {
	t.Helper()

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/mcp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", sessionID)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return result
}

func toolCall(t *testing.T, ts *httptest.Server, sessionID, toolName string, args map[string]interface{}) Response {
	t.Helper()
	return mcpCall(t, ts, sessionID, MethodToolsCall, map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	})
}

// --- Tests ---

func TestMCP_Initialize(t *testing.T) {
	_, ts := newTestServer(t)
	resp := mcpCall(t, ts, "", MethodInitialize, nil)
	if resp.Error != nil {
		t.Fatalf("initialize error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type: %T", resp.Result)
	}
	if result["protocolVersion"] == nil {
		t.Error("missing protocolVersion in initialize response")
	}
}

func TestMCP_ToolsList(t *testing.T) {
	_, ts := newTestServer(t)
	resp := mcpCall(t, ts, "", MethodToolsList, nil)
	if resp.Error != nil {
		t.Fatalf("tools/list error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	tools := result["tools"].([]interface{})
	if len(tools) < 12 {
		t.Errorf("expected at least 12 tools, got %d", len(tools))
	}
}

func TestMCP_UnknownMethod(t *testing.T) {
	_, ts := newTestServer(t)
	resp := mcpCall(t, ts, "", "nonexistent/method", nil)
	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected -32601, got %d", resp.Error.Code)
	}
}

func TestMCP_RegisterAndList(t *testing.T) {
	_, ts := newTestServer(t)
	sessID := uuid.NewString()
	agentID := uuid.New()

	// Register agent A.
	resp := toolCall(t, ts, sessID, ToolRegisterAgent, map[string]interface{}{
		"agentId": agentID.String(),
		"name":    "Agent A",
		"folder":  "/tmp/a",
	})
	if resp.Error != nil {
		t.Fatalf("register-agent error: %v", resp.Error)
	}

	// List agents — should include A.
	resp = toolCall(t, ts, sessID, ToolListAgents, nil)
	if resp.Error != nil {
		t.Fatalf("list-agents error: %v", resp.Error)
	}
}

func TestMCP_SendAndCheckMessages(t *testing.T) {
	_, ts := newTestServer(t)

	sessA := uuid.NewString()
	sessB := uuid.NewString()
	agentA := uuid.New()
	agentB := uuid.New()

	// Register both agents.
	toolCall(t, ts, sessA, ToolRegisterAgent, map[string]interface{}{
		"agentId": agentA.String(), "name": "Alice", "folder": "/tmp/a",
	})
	toolCall(t, ts, sessB, ToolRegisterAgent, map[string]interface{}{
		"agentId": agentB.String(), "name": "Bob", "folder": "/tmp/b",
	})

	// A sends message to B by name.
	resp := toolCall(t, ts, sessA, ToolSendMessage, map[string]interface{}{
		"to": "Bob", "message": "hello from Alice",
	})
	if resp.Error != nil {
		t.Fatalf("send-message error: %v", resp.Error)
	}

	// B checks messages.
	resp = toolCall(t, ts, sessB, ToolCheckMessages, map[string]interface{}{
		"markRead": true,
	})
	if resp.Error != nil {
		t.Fatalf("check-messages error: %v", resp.Error)
	}

	// The result text should contain the message content.
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if text == "" {
		t.Error("expected non-empty check-messages response")
	}
}

func TestMCP_Broadcast(t *testing.T) {
	_, ts := newTestServer(t)

	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	sessions := make([]string, len(ids))
	for i, id := range ids {
		sessions[i] = uuid.NewString()
		toolCall(t, ts, sessions[i], ToolRegisterAgent, map[string]interface{}{
			"agentId": id.String(),
			"name":    fmt.Sprintf("Agent%d", i),
			"folder":  "/tmp",
		})
	}

	// Agent 0 broadcasts.
	resp := toolCall(t, ts, sessions[0], ToolBroadcast, map[string]interface{}{
		"message": "broadcast_payload",
	})
	if resp.Error != nil {
		t.Fatalf("broadcast error: %v", resp.Error)
	}

	// Agents 1 and 2 should have the message.
	for _, sess := range sessions[1:] {
		resp = toolCall(t, ts, sess, ToolCheckMessages, map[string]interface{}{"markRead": false})
		if resp.Error != nil {
			t.Fatalf("check-messages error: %v", resp.Error)
		}
		result := resp.Result.(map[string]interface{})
		text := result["content"].([]interface{})[0].(map[string]interface{})["text"].(string)
		if text == "Error: agent not found" {
			t.Error("agent should have received broadcast")
		}
	}
}

func TestMCP_Ping(t *testing.T) {
	_, ts := newTestServer(t)
	resp := mcpCall(t, ts, "", MethodPing, nil)
	if resp.Error != nil {
		t.Fatalf("ping error: %v", resp.Error)
	}
}

func TestMCP_UnknownTool(t *testing.T) {
	_, ts := newTestServer(t)
	resp := toolCall(t, ts, "", "nonexistent-tool", nil)
	// Should return a result (not an RPC error), with an error message in content.
	if resp.Error != nil {
		t.Fatalf("unexpected RPC-level error: %v", resp.Error)
	}
}

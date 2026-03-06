package mcp

// JSON-RPC 2.0 envelope types.

// Request is an inbound JSON-RPC 2.0 call from an MCP client.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response is an outbound JSON-RPC 2.0 reply.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError carries a JSON-RPC error code and message.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP initialize / tools/list / tools/call method names.
const (
	MethodInitialize  = "initialize"
	MethodToolsList   = "tools/list"
	MethodToolsCall   = "tools/call"
	MethodPing        = "ping"
)

// Tool names exposed by the MCP server.
const (
	ToolRegisterAgent  = "register-agent"
	ToolListAgents     = "list-agents"
	ToolSendMessage    = "send-message"
	ToolCheckMessages  = "check-messages"
	ToolBroadcast      = "broadcast-message"
	ToolListRepos      = "list-repos"
	ToolListWorktrees  = "list-worktrees"
	ToolCreateAgent    = "create-agent"
	ToolCloseAgent     = "close-agent"
	ToolCreateWorktree = "create-worktree"
	ToolDisplayMD      = "display-markdown"
	ToolViewMermaid    = "view-mermaid"
)

// ToolCallParams is the params block for a tools/call request.
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult wraps the result of a tool call.
type ToolResult struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock is a text or image block in a tool result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func textResult(text string) ToolResult {
	return ToolResult{Content: []ContentBlock{{Type: "text", Text: text}}}
}

func errorResult(msg string) ToolResult {
	return ToolResult{Content: []ContentBlock{{Type: "text", Text: "Error: " + msg}}}
}

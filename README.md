# Skwad Linux

A native Linux desktop application for running multiple AI coding agents simultaneously — each in its own embedded terminal, coordinated via a built-in MCP (Model Context Protocol) server.

This is a Go port of the [Skwad macOS app](https://skwad.ai), built for X11 and Wayland desktops.

## What It Does

- Runs multiple AI coding agents in parallel (Claude Code, Codex, Gemini CLI, GitHub Copilot, OpenCode, or custom shell commands), each in a persistent embedded terminal
- Organizes agents into named, color-coded **workspaces** with 1, 2, 3, and 4-pane split layouts
- Built-in **MCP HTTP server** (JSON-RPC 2.0) so agents can register, message each other, query worktrees, and spawn new agents programmatically
- **Git integration**: diff viewer, per-file staging, commit dialog, worktree creation, live git stats in the sidebar
- **Markdown** and **Mermaid** diagram preview panels rendered on demand via MCP tool calls
- **Fuzzy file finder**, agent personas, conversation history browser, and an autopilot service that uses an LLM to handle agent prompts automatically

## Screenshots

> Coming soon — the app is currently in active development.

## Getting Started

### Requirements

**Linux (for full functionality):**
```
sudo apt install libvte-2.91-dev libgtk-3-dev pkg-config
```

**macOS (for development/testing — VTE not available):**
```
brew install go
```

### Build

```bash
git clone https://github.com/Jared-Boschmann/skwad-linux
cd skwad-linux
make build
```

Or with `go` directly:
```bash
go build -o skwad ./cmd/skwad
./skwad
```

### Run

```bash
./skwad
```

Configuration is stored in `~/.config/skwad/`. On first launch with no agents configured, you'll see an empty workspace — use **Ctrl+N** (or **Cmd+N** on macOS) to create your first agent.

## Architecture

```
cmd/skwad/           Entry point
internal/
  models/            Pure data types (Agent, Workspace, Settings, Persona)
  agent/             Business logic: Manager, Coordinator, ActivityController
  mcp/               MCP HTTP server (JSON-RPC 2.0) + hook event handler
  terminal/          PTY session management (creack/pty) + Pool orchestrator
  git/               Git CLI wrapper: status, diff, stage, commit, worktrees
  persistence/       JSON file storage (~/.config/skwad/)
  search/            Fuzzy file path scorer
  ui/                Fyne v2 GUI components
  autopilot/         LLM-based autopilot (stub)
  notifications/     Desktop notifications via notify-send
  voice/             Push-to-talk voice input (stub)
plugin/
  claude/notify.sh   Hook script for Claude Code lifecycle events
  codex/notify.sh    Hook script for Codex lifecycle events
```

See [`DEVPLAN.md`](DEVPLAN.md) for the full phased build plan and [`AGENTS.md`](AGENTS.md) for the developer guide.

## Tech Stack

| Concern | Choice |
|---|---|
| Language | Go 1.22+ |
| GUI | [Fyne v2](https://fyne.io/) — OpenGL, X11 + Wayland |
| Terminal widget | VTE (libvte-2.91) via CGo on Linux |
| PTY sessions | [creack/pty](https://github.com/creack/pty) |
| MCP server | `net/http` stdlib, JSON-RPC 2.0 |
| File watching | [fsnotify](https://github.com/fsnotify/fsnotify) |
| Markdown | [goldmark](https://github.com/yuin/goldmark) |
| Persistence | JSON in `~/.config/skwad/` |

## MCP Server

Skwad exposes an MCP server at `http://127.0.0.1:8766/mcp` (port configurable in settings). AI agents that support MCP can use the following tools:

| Tool | Description |
|---|---|
| `register-agent` | Register with the Skwad session |
| `list-agents` | List all registered agents |
| `send-message` | Send a message to another agent |
| `check-messages` | Read messages in your inbox |
| `broadcast-message` | Send to all registered agents |
| `list-repos` | List recently used git repos |
| `list-worktrees` | List worktrees for a repo |
| `create-worktree` | Create a new git worktree |
| `display-markdown` | Open a markdown file in the preview panel |
| `view-mermaid` | Render a Mermaid diagram |
| `create-agent` | Spawn a new agent |
| `close-agent` | Close an agent |

Hook scripts in `plugin/claude/` and `plugin/codex/` post lifecycle events to `/hook` so Skwad can track agent status (running / idle / blocked).

## Keyboard Shortcuts

| Shortcut | Action |
|---|---|
| Ctrl+N | New agent |
| Ctrl+G | Toggle git panel |
| Ctrl+\ | Toggle sidebar |
| Ctrl+P | Fuzzy file finder |
| Ctrl+] | Next agent |
| Ctrl+[ | Previous agent |

> On macOS, replace Ctrl with Cmd.

## Development Status

| Phase | Status |
|---|---|
| Phase 1 — Data layer (models, persistence, manager) | ✅ Complete |
| Phase 2 — Git + file services | ✅ Complete |
| Phase 3 — Terminal + MCP server (headless) | ✅ Complete |
| Phase 4 — UI shell | 🔄 In progress |
| Phase 5 — Feature panels (git, markdown, mermaid, file finder) | Planned |
| Phase 6 — Autopilot, voice, notifications, polish | Planned |

## Testing

```bash
make test
# or
go test ./...
```

All packages through Phase 3 have unit tests. The `ui` package is manually verified.

## License

MIT

## Contributing

Issues and pull requests welcome. See [`DEVPLAN.md`](DEVPLAN.md) for the build plan and coding rules before contributing.

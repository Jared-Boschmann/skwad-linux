# Skwad Linux — Agent Development Guide

## Project Overview

Skwad Linux is a Go-based Linux desktop application replicating the Skwad macOS app.
It manages multiple AI coding agent CLIs simultaneously, each in an embedded PTY terminal,
coordinated via an in-process MCP HTTP server.

## Tech Stack

- **Language**: Go 1.23+
- **GUI**: [Fyne v2](https://fyne.io/) — cross-platform, OpenGL-accelerated
- **Terminal**: PTY via `creack/pty`; VTE (libvte-2.91) via CGo on Linux (`//go:build linux`)
- **MCP Server**: `net/http` stdlib (JSON-RPC 2.0, port 8766)
- **File watching**: fsnotify with 500ms debounce
- **Markdown**: goldmark + Fyne RichText
- **Persistence**: JSON files in `~/.config/skwad/`

## Project Structure

```
cmd/skwad/main.go              Entry point; MCP server failure is non-fatal

internal/
  models/                      Pure data types (no I/O or UI dependencies)
    agent.go                   Agent, AgentType, AgentStatus, ActivityTracking, GitStats
    workspace.go               Workspace, LayoutMode, WorkspaceColors
    settings.go                AppSettings, AppearanceMode, AutopilotAction
    persona.go                 Persona, PersonaType, PersonaState; fixed UUIDs (never change)
    bench.go                   BenchAgent + ToAgent()

  agent/
    manager.go                 Thread-safe owner of all agent + workspace state
    coordinator.go             Goroutine-safe MCP message queue and agent registry
    command_builder.go         Constructs shell commands per agent type
    activity.go                Per-agent status state machine + 10s input guard
    registration.go            RegistrationPrompt() — text injected ~3s after terminal start

  mcp/
    server.go                  HTTP server, JSON-RPC 2.0 dispatch (/mcp and /hook)
    hooks.go                   Hook event parsing; AgentStatusUpdater interface
    tools.go                   All 12 tool definitions + implementations
    types.go                   JSON-RPC structs, tool name constants
    session_manager.go         Per-client session tracking (agentID after register-agent)

  git/
    cli.go                     CLI — low-level git runner (30s timeout)
    repository.go              Repository — Branch, Status, Diff, Stage, Commit, NumStat, LsFiles
    worktree.go                WorktreeManager — List, Create, SuggestedPath
    watcher.go                 Watcher — fsnotify with 500ms debounce
    types.go                   FileStatus, DiffLine, BranchInfo, Worktree, RepoStats
    discovery.go               AutoDetectSourceDir, DiscoverRepos, ExcludedDirs, IsExcluded

  terminal/
    pool.go                    Pool — central orchestrator (Session + ActivityController per agent)
    session.go                 Session — PTY process via creack/pty
    cleaner.go                 StripANSI, CleanTitle; ANSI/OSC regex
    manager.go                 Terminal interface + keep-alive registry; registrationDelay = 3s
    vte.go                     VTE CGo bindings (//go:build linux)
    vte_stub.go                No-op VTETerminal stub (//go:build !linux)
    vte_impl.h                 C shim header for VTE integration

  persistence/
    store.go                   Store — JSON persistence; NewStoreAt(dir) for test isolation
                               Includes migration defaults for old config files

  search/
    fuzzy.go                   FuzzySearch — subsequence scorer with consecutive/separator bonuses

  ui/
    app.go                     Fyne app + main window layout
    workspace_bar.go           Left workspace badge strip
    sidebar.go                 Agent list with status indicators
    terminal_area.go           Split-pane layout (single/2/3/4 panes)
    terminal_pane.go           Single pane slot + VTE overlay coordination
    git_panel.go               Sliding git status panel
    markdown_panel.go          Markdown preview (goldmark + Fyne RichText)
    file_finder.go             Fuzzy file search overlay
    agent_sheet.go             New/edit agent dialog
    settings_window.go         Settings window

  autopilot/
    autopilot.go               LLM-based autopilot (analyzes idle agent output)

  notifications/
    service.go                 Desktop notifications via notify-send (libnotify)

  voice/
    service.go                 Push-to-talk voice input stub

plugin/
  claude/notify.sh             Hook script: POSTs lifecycle events to /hook
  codex/notify.sh              Hook script: POSTs lifecycle events to /hook
```

## Test Coverage

Tests live alongside the packages they test:

| Package | Test file | What's covered |
|---------|-----------|----------------|
| `search` | `fuzzy_test.go` | scoring, empty query, separator bonuses |
| `models` | `agent_test.go`, `workspace_test.go`, `persona_test.go` | ActivityMode, WorstStatus, fixed persona UUIDs |
| `agent` | `command_builder_test.go`, `manager_test.go` | all agent types, fork, resume, companion removal |
| `persistence` | `store_test.go` | round-trip, migration defaults, recent repos |
| `mcp` | `integration_test.go` | initialize, tools/list, register+list, send+check, broadcast, ping |
| `terminal` | `cleaner_test.go`, `session_test.go` | StripANSI, CleanTitle, PTY spawn/exit/resize |
| `git` | `repository_test.go`, `watcher_test.go` | Branch, Status, Stage, Commit, Diff, debounce |

Run all tests:
```
make test
```

On macOS: all packages except `terminal/vte.go` (Linux-only) build and test cleanly.
The VTE stub (`vte_stub.go`) satisfies the `Terminal` interface on non-Linux platforms.

## Build Requirements (Linux)

```
sudo apt install libvte-2.91-dev libgtk-3-dev pkg-config
make build
```

On macOS for development:
```
go run ./cmd/skwad        # launches Fyne window with placeholder panes
go test ./...             # all tests pass natively
```

## Key Architectural Decisions

### PTY Sessions vs VTE
`terminal.Session` (`creack/pty`) drives all PTY I/O on both platforms.
VTE (`terminal/vte.go`, Linux only) is a higher-level GTK terminal widget used for
the actual rendered display inside the Fyne window.

### VTE + Fyne Embedding
Fyne uses its own OpenGL renderer and cannot host GTK widgets directly.
Two strategies:
1. **XEmbed (X11)**: embed VTE's GtkPlug into Fyne's native window via GtkSocket.
2. **Overlay window**: a borderless child X11 window containing VTE, kept in sync
   with the Fyne pane's screen coordinates.

`internal/terminal/vte.go` is the CGo stub. `vte_impl.h` declares the C interface.

### Terminal Pool
`terminal.Pool` is the single owner that wires together:
- `agent.Manager` (persisted state)
- `terminal.Session` (PTY I/O)
- `agent.ActivityController` (status state machine)
- `agent.Coordinator` (message queue)

Shell agents use a staggered startup delay to avoid hammering the system on restore.

### Agent Status State Machine
`agent.ActivityController` owns the status transitions per agent:
- `idle` → `running`: terminal output (non-hook mode) or hook `PreToolUse`/`Start`
- `running` → `idle`: no output for 5s timeout, or hook `PostToolUse`/`Stop`
- any → `input`: hook `Notify`/`ask` signals permission prompt
- `input` → `running`: user presses Return
- `input` → `idle`: user presses Escape

### MCP Message Flow
1. Agent terminal spawns
2. After 3s (`registrationDelay`), registration prompt is injected
3. Agent calls `register-agent` → `Coordinator.RegisterAgent()`
4. Agent can call `send-message`, `check-messages`, `broadcast-message`
5. When agent goes idle, `Coordinator.NotifyIdleAgent()` delivers any queued messages

### Input Protection Guard
User keypresses activate a 10-second guard that blocks `InjectText`.
Texts queued during the guard are delivered when the guard expires or on next idle.

### Hook Scripts
`plugin/claude/notify.sh` and `plugin/codex/notify.sh` POST lifecycle events to
`POST /hook` with `SKWAD_AGENT_ID`. The `hookHandler` normalises Claude's
`hook_event_name` field and Codex's `eventType` field to a common `HookEventType`.

### MCP Server Non-Fatal Startup
If port 8766 is already in use (e.g., the macOS Skwad app is also running),
the MCP server logs a warning and the app continues without it. Agents will not
receive registration prompts or be able to exchange messages.

### Persona Fixed UUIDs
System personas (Kent Beck, Martin Fowler, etc.) use hardcoded UUIDs that must
never change across installs — agents refer to them by ID in persisted config.

## Common Tasks

### Adding a new MCP tool
1. Add constant to `mcp/types.go`
2. Add definition in `toolHandler.list()` in `mcp/tools.go`
3. Add switch case in `toolHandler.call()`
4. Implement handler method
5. Wire any needed `Server` callback for UI interaction

### Adding a new agent type
1. Add constant to `models/agent.go`
2. Add command construction in `agent/command_builder.go`
3. Update `Agent.SupportsHooks()`, `SupportsResume()`, `ActivityMode()` if needed
4. Add to `AgentSheet` type selector in `ui/agent_sheet.go`

### Adding a git operation
1. Add low-level command in `git/cli.go` if needed
2. Add method to `git/repository.go`
3. Expose in `ui/git_panel.go`

### Adding a new persistence field
Fields on `models.Agent` or `models.Workspace` tagged `json:"-"` are runtime-only.
Persisted fields must have a migration default in `persistence/store.go`
(`LoadAgents` or `LoadWorkspaces`) to handle old config files.

## Reference Implementation
See `Skwad-Mac/` for the Swift/macOS reference. The MCP tools, agent types,
command construction, and data models are direct ports.

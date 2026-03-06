# Skwad Linux — Agent Development Plan

This document is the primary reference for any AI agent working on this codebase.
Read it before starting any task. It defines the build order, module boundaries,
testing strategy, and decision rules that keep parallel work coherent.

---

## Project Goal

Port the Skwad macOS app to a Linux desktop application written in Go.
All existing functionality must be preserved. The target runtime is Linux
desktop (X11 and Wayland). See `REQUIREMENTS.md` for the full feature specification.

---

## Repository Layout (target)

```
skwad-linux/
├── cmd/
│   └── skwad/
│       └── main.go              # Entry point
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Agent model + status types
│   │   ├── manager.go           # AgentManager (lifecycle, layout, git stats)
│   │   └── bench.go             # Bench agent templates
│   ├── workspace/
│   │   ├── workspace.go         # Workspace model
│   │   └── colors.go            # WorkspaceColor palette
│   ├── mcp/
│   │   ├── server.go            # HTTP JSON-RPC 2.0 server
│   │   ├── coordinator.go       # AgentCoordinator (thread-safe message queue)
│   │   ├── tools.go             # Tool handler implementations
│   │   └── types.go             # MCP protocol structs
│   ├── terminal/
│   │   ├── session.go           # TerminalSession (process lifecycle)
│   │   ├── controller.go        # TerminalSessionController (status state machine)
│   │   ├── command.go           # Command builder (agent CLI flags)
│   │   └── cleaner.go           # Terminal text/title cleaning utilities
│   ├── git/
│   │   ├── cli.go               # Low-level git command runner (30s timeout)
│   │   ├── repository.go        # High-level git operations (status, diff, stage, commit)
│   │   ├── worktree.go          # Worktree discovery and creation
│   │   ├── watcher.go           # inotify/fsnotify file watcher with debounce
│   │   ├── parser.go            # Output parsers (status, diff, numstat)
│   │   └── types.go             # FileStatus, DiffLine, GitLineStats, etc.
│   ├── history/
│   │   ├── service.go           # ConversationHistoryService
│   │   ├── claude.go            # Claude session file parser
│   │   ├── codex.go             # Codex session file parser
│   │   ├── gemini.go            # Gemini session file parser
│   │   └── copilot.go           # Copilot session file parser
│   ├── autopilot/
│   │   ├── service.go           # AutopilotService (LLM classification + actions)
│   │   └── providers.go         # OpenAI / Anthropic / Google API calls
│   ├── voice/
│   │   └── service.go           # VoiceInputManager (push-to-talk, STT)
│   ├── persona/
│   │   └── persona.go           # Persona model + default personas
│   ├── discovery/
│   │   └── repos.go             # Background git repo discovery service
│   ├── filesearch/
│   │   └── service.go           # FileSearchService (fuzzy, git-aware)
│   ├── notify/
│   │   └── service.go           # Desktop notification service
│   ├── settings/
│   │   └── settings.go          # AppSettings (load/save JSON config)
│   └── ui/
│       ├── app.go               # App entry, window setup
│       ├── workspacebar.go      # Workspace switcher bar
│       ├── sidebar.go           # Agent list sidebar
│       ├── terminal_view.go     # Terminal pane(s) with layout modes
│       ├── split.go             # Split divider logic
│       ├── git_panel.go         # Sliding git status panel
│       ├── markdown_panel.go    # Markdown preview panel
│       ├── mermaid_panel.go     # Mermaid diagram panel
│       ├── file_finder.go       # Fuzzy file finder overlay
│       ├── settings_view.go     # Settings window
│       ├── agent_sheet.go       # New/edit agent dialog
│       └── autopilot_sheet.go   # Autopilot decision sheet
├── plugin/
│   ├── claude/                  # Claude hook plugin (notify.sh)
│   └── codex/                   # Codex hook plugin (notify.sh)
├── REQUIREMENTS.md              # Full feature specification
├── DEVPLAN.md                   # This file
├── go.mod
└── go.sum
```

---

## Technology Decisions

| Concern | Choice | Rationale |
|---|---|---|
| GUI | [Fyne v2](https://fyne.io/) | Pure Go, cross-platform (X11 + Wayland), no CGo required for UI layer |
| Terminal emulation | VTE via CGo OR embed [go-vte](https://github.com/nicowillis/go-vte) | VTE is the standard Linux terminal widget; Fyne custom widget wraps it |
| MCP HTTP server | `net/http` stdlib | No external dependency needed for JSON-RPC 2.0 |
| File watching | [fsnotify](https://github.com/fsnotify/fsnotify) | Cross-platform, inotify on Linux |
| Markdown render | [goldmark](https://github.com/yuin/goldmark) → HTML → WebView | Renders to HTML; display in a Fyne WebView or `webkit2gtk` |
| Mermaid render | Embed `mermaid.js` in WebView | Avoids `mmdc` dependency; renders in-process |
| Fuzzy search | Custom `FuzzyScorer` (port from Swift) | Lightweight, no dependency |
| Persistence | JSON files in `~/.config/skwad/` | Simple, human-readable, easy to migrate |
| Logging | `log/slog` (stdlib, Go 1.21+) | Structured, no dependency |
| Speech-to-text | `speech-dispatcher` via exec OR `vosk` library | Linux-native STT |

**Go version target:** 1.22+

---

## Build Phases

Work is divided into six phases. Each phase produces a runnable or testable artifact.
Do not start a phase until all items in the previous phase are complete and passing tests.

---

### Phase 1 — Core Data Layer (no UI)

**Goal:** All models, persistence, and business logic compile and are tested.
No UI, no terminal, no HTTP server.

**Tasks:**

- [x] Initialize `go.mod` — module `github.com/kochava-studios/skwad-linux`, deps: fyne, fsnotify, uuid, goldmark
- [x] `internal/models/agent.go` — `Agent`, `AgentStatus`, `AgentType`, `ActivityTracking`; all fields from REQUIREMENTS §2.1; `SupportsHooks`, `SupportsResume`, `ActivityMode`
- [x] `internal/models/workspace.go` — `Workspace`, `WorkspaceColors`, `LayoutMode`; `PaneCount()`, `WorstStatus()`
- [x] `internal/models/persona.go` — `Persona`, `PersonaType`, `PersonaState`; all six default personas with fixed UUIDs
- [x] `internal/models/bench.go` — `BenchAgent` struct; `ToAgent()`
- [x] `internal/models/settings.go` — `AppSettings`; all fields from REQUIREMENTS §3.18; `DefaultSettings()`
- [x] `internal/persistence/store.go` — load/save agents, workspaces, personas, bench, settings, recent repos to `~/.config/skwad/`; migration defaults; auto-detect source folder in `git/discovery.go`
- [x] `internal/agent/manager.go` — `Manager`; CRUD (add, remove, restart, duplicate, move); workspace CRUD; companion cleanup on remove; default workspace auto-create
- [x] Unit tests: agent activity modes, hook/resume/system-prompt support; workspace `WorstStatus` priority, `PaneCount`; persona fixed UUIDs, no duplicates; settings round-trip, migration defaults; recent repos ordering, dedup, cap; command builder for all agent types; fuzzy search
- [x] Companion rules test (hidden in sidebar, removed with creator)
- [x] `agent.Manager.ForkAgent()`, `ResumeAgent()`, `Companions()` — fork/resume session wiring
- [x] Manager tests: companion rules, fork, resume, duplicate, restart, default workspace auto-create

**Exit criteria:** `go test ./internal/...` passes. ✅

---

### Phase 2 — Git + File Services (no UI)

**Goal:** All git operations and background services work end-to-end.

**Tasks:**

- [x] `internal/git/cli.go` — shell out to `git` binary; 30-second timeout
- [x] `internal/git/types.go` — `FileStatus`, `DiffLine`, `BranchInfo`, `Worktree`, `RepoStats`
- [x] `internal/git/repository.go` — `Branch()`, `Status()`, `Diff()`, `Stage()`, `StageAll()`, `Unstage()`, `UnstageAll()`, `Discard()`, `Commit()`, `NumStat()`, `LsFiles()`
- [x] `internal/git/worktree.go` — `List()`, `Create()`, `SuggestedPath()`
- [x] `internal/git/watcher.go` — fsnotify watcher with 500ms debounce
- [x] `internal/git/discovery.go` — `AutoDetectSourceDir()`, `DiscoverRepos()`
- [x] `internal/search/fuzzy.go` — fuzzy scorer with consecutive/separator bonuses, match indices
- [x] `internal/notifications/service.go` — `notify-send` wrapper
- [x] Tests: git operations against temp repo — Branch, Status, Stage/Unstage, Commit, Diff, NumStat, LsFiles, IsRepo, RootOf, SuggestedPath
- [x] Tests: file watcher debounce and stop
- [x] File search exclusion list (`ExcludedDirs`, `IsExcluded`) in `git/discovery.go`
- [x] `persistence.NewStoreAt(dir)` for test isolation
- [ ] `internal/history/` — claude, codex, gemini, copilot session file parsers
- [ ] Extract `parseStatus`/`parseDiff` to `git/parser.go` (currently inline; low priority)

**Exit criteria:** ✅ git tests pass; watcher test passes

---

### Phase 3 — Terminal + MCP Server (headless)

**Goal:** Terminals spawn agents and agents can communicate via MCP. Runnable as a headless daemon for integration testing.

**Tasks:**

- [x] `internal/agent/command_builder.go` — all agent types; MCP flags; persona injection; resume/fork flags; `shellQuote`, `shellEscapeDouble`
- [x] `internal/agent/activity.go` — `ActivityController`; activity tracking modes (all/userInput/none); status state machine; idle timeout; input protection guard (10s); `QueueText()` with delivery
- [x] `internal/mcp/types.go` — JSON-RPC 2.0 structs; all tool names
- [x] `internal/mcp/server.go` — `net/http` server; JSON-RPC 2.0 dispatch; configurable port; start/stop; graceful non-fatal on port conflict
- [x] `internal/mcp/session_manager.go` — per-client session tracking
- [x] `internal/mcp/tools.go` — all 12 tool definitions + stub implementations; UI callbacks for display-markdown, view-mermaid, create-agent, close-agent
- [x] `internal/agent/coordinator.go` — goroutine-safe message queue; `RegisterAgent`, `SendMessage`, `BroadcastMessage`, `CheckMessages`, `NotifyIdleAgent`, `UnregisterAgent`
- [x] `internal/terminal/session.go` — PTY process via `creack/pty`; read/write; resize; OSC title parsing; callbacks: `onOutput`, `onTitleChange`, `onExit`
- [x] `internal/terminal/cleaner.go` — ANSI/OSC escape stripping; spinner char removal; status prefix stripping
- [x] Hook handler (`internal/mcp/hooks.go`) — parse claude/codex hook events; `AgentStatusUpdater` interface; dispatch running/idle/blocked/error
- [x] `plugin/claude/notify.sh` and `plugin/codex/notify.sh` — hook scripts posting to `/hook` endpoint with `SKWAD_AGENT_ID`
- [x] MCP integration tests — initialize, tools/list, register+list, send+check messages, broadcast, ping, unknown method/tool ← **DONE**

**Exit criteria:** ✅ all tests passing (`go test ./...`)

---

### Phase 4 — UI Shell (no content yet)

**Goal:** A working window with correct layout structure, workspace bar, sidebar, and multi-pane terminal area. Terminals display and accept input. No git panel, no markdown panel yet.

**Tasks:**

- [x] Fyne v2 application skeleton scaffolded (`ui/app.go`, `ui/workspace_bar.go`, `ui/sidebar.go`, `ui/terminal_area.go`, `ui/terminal_pane.go`)
- [x] `internal/terminal/pool.go` — `Pool` bridges AgentManager ↔ PTY sessions ↔ ActivityControllers; shell staggered startup; registration prompt scheduling; hook event routing
- [x] `internal/agent/registration.go` — `RegistrationPrompt()` per agent type
- [x] `agent.Manager.ActiveSettings()`, `agent.Manager.Persona()` helpers
- [x] `cmd/skwad/main.go` — `hookBridge` wires MCP hook events → Pool; `pluginDirectory()` auto-detects plugin path; Pool created after MCP server with correct mcpURL
- [x] Workspace bar: `parseHexColor` (#RRGGBB → color.NRGBA); status dot (green/orange/red per WorstStatus); rounded badge; initials
- [x] Sidebar: collapsible (Toggle); right-click context menu (Restart, Remove); `OpenNewAgentSheet()`; `OnAddAgent`/`OnRemoveAgent`/`OnRestartAgent` callbacks; `agentRow.SecondaryTapped`
- [x] Terminal pool wired: `App.NewApp` spawns existing agents on startup; `OnAddAgent` calls `pool.Spawn` outside manager lock (avoids deadlock); `hookBridge` routes MCP hook events to pool
- [x] Keyboard shortcuts: Ctrl/Cmd+N (new agent), +G (git panel), +\ (toggle sidebar), +P (file finder stub), +]/[ (next/prev agent)
- [x] Main layout uses `container.NewHSplit` for sidebar/terminal split — gives built-in drag handle
- [ ] Terminal pane: show/hide VTE overlay per layout; focus management (Linux-only)
- [ ] Split dividers: persist ratio changes back to Workspace on drag end
- [ ] Agent drag-to-reorder in sidebar (requires custom List widget)
- [ ] Keyboard shortcuts: Cmd+1-9 select agent by index; Cmd+Shift+]/[ switch workspace

**Exit criteria:** launch app, create two agents, verify both terminals spawn and stay alive on switch; drag sidebar resize handle; switch workspaces; reorder agents via drag.

---

### Phase 5 — Feature Panels

**Goal:** Git panel, markdown panel, Mermaid panel, file finder all functional.

**Tasks:**

- [ ] `internal/ui/git_panel.go` — sliding panel (bottom of active pane); file list with status icons; stage/unstage per file; stage all / unstage all; diff view (syntax-highlighted +/- lines); commit dialog; toggle staged/unstaged diff; ahead/behind/branch display
- [ ] `internal/ui/markdown_panel.go` — right-side sliding panel; WebView with goldmark-rendered HTML; file path history navigation; text highlight → review comment injection; maximized mode; themed (dark/light)
- [ ] `internal/ui/mermaid_panel.go` — WebView with embedded mermaid.js; theme + scale settings; open alongside markdown panel
- [ ] `internal/ui/file_finder.go` — Cmd+P overlay; fuzzy search input; results list with match highlighting; open file in configured editor; uses `FileSearchService`
- [ ] `internal/ui/agent_sheet.go` — new/edit agent dialog: folder picker, worktree picker (from `RepoDiscoveryService`), name, avatar (emoji picker + image upload), agent type, persona selector
- [ ] `internal/ui/settings_view.go` — all settings sections from REQUIREMENTS §3.18; live apply where possible (e.g., MCP port requires restart)

**Exit criteria:** open git panel on a repo with changes; stage a file; commit; open markdown panel from MCP tool call; search files in file finder.

---

### Phase 6 — Autopilot, Voice, Notifications, Polish

**Goal:** All remaining features complete. App is shippable.

**Tasks:**

- [ ] `internal/autopilot/service.go` — `analyze()`, `classify()`, `dispatchAction()`; all four action modes (mark/ask/continue/custom); three LLM providers
- [ ] `internal/autopilot/providers.go` — OpenAI, Anthropic, Google API calls; configurable model per provider
- [ ] `internal/ui/autopilot_sheet.go` — decision sheet: show agent's last message, classification, approve/decline/custom reply buttons
- [ ] `internal/voice/service.go` — push-to-talk monitor (global key listener); STT via `speech-dispatcher` or `vosk`; waveform overlay during recording; auto-inject on release
- [ ] Conversation history UI in sidebar — session list per agent; resume / fork / delete actions
- [ ] Desktop notifications — wire `NotificationService` to agent status changes; `notify-send` or `libnotify`
- [ ] Appearance mode — auto (detect from terminal bg luminance), system, light, dark
- [ ] Agent metadata display — show model name, cwd from hook metadata in sidebar tooltip or detail view
- [ ] Persona management UI — create/edit/delete personas; restore defaults button
- [ ] Bench UI — bench section in sidebar or settings; deploy / rename / remove entries
- [ ] MCP `display-markdown` and `view-mermaid` wired to UI panels
- [ ] Full keyboard shortcut pass — verify all shortcuts from REQUIREMENTS §7 work
- [ ] System tray / menu bar mode (keep-in-tray setting)
- [ ] Packaging — produce `.deb`, `.rpm`, and AppImage via GitHub Actions

**Exit criteria:** full manual test checklist (see below) passes on Ubuntu 24.04 LTS and Fedora 40.

---

## Manual Test Checklist

Run before any release:

1. Launch app fresh (no config); verify source folder auto-detection
2. Create agent from repo picker; verify terminal spawns and agent command runs
3. Create agent with new worktree; verify worktree created on disk
4. Switch agents; verify terminal state/history preserved
5. Reorder agents via drag; verify order persists after restart
6. Open git panel; stage file; view diff; commit; verify commit appears in git log
7. Send message between two agents via MCP; verify delivery notification
8. Broadcast message; verify all agents receive it
9. Open markdown panel via `display-markdown` MCP tool; highlight text; submit review; verify text injected
10. Open Mermaid panel; verify diagram renders
11. Fuzzy file search: open file finder; search for a file; open in editor
12. Resume a Claude session; verify `--resume <id>` in spawned command
13. Fork a Codex session; verify `codex fork <id>` in spawned command
14. Restart agent; verify same ID kept but fresh terminal
15. Autopilot: enable; run agent to completion; verify no action taken (completed)
16. Autopilot mark: trigger binary question; verify status turns red + notification
17. Voice: hold push-to-talk; speak; verify transcription injected on release
18. Workspace: create second workspace; move agent; switch workspaces; verify layout saved
19. Bench: save agent to bench; remove agent; deploy from bench; verify agent restored
20. Companion: add shell companion to agent; verify split layout; verify companion removed when creator removed
21. Change settings (font, colors, MCP port); restart app; verify settings restored
22. Quit and relaunch; verify all agents, workspaces, and layout restored

---

## Coding Rules for Agents

1. **Read `REQUIREMENTS.md` before implementing any feature.** It is the source of truth.
2. **One package per concern.** Do not put UI code in `internal/agent` or vice versa.
3. **Thread safety:** `AgentManager` must be accessed from a single goroutine or protected with a mutex. `AgentCoordinator` is the only place that manages the MCP message queue — never access it directly from the UI goroutine without going through its public API.
4. **No global mutable state** except `AppSettings.shared` (singleton, read-only after init from non-main goroutines) and the two service singletons (`AgentCoordinator`, `RepoDiscoveryService`).
5. **Terminals are never destroyed on agent switch** — only hidden. This is a hard requirement.
6. **Agent IDs are stable across restarts.** Never regenerate an ID on restart; only `restartToken` changes.
7. **MCP messages are intentionally in-memory only.** Do not persist them.
8. **Default persona UUIDs are fixed** — hardcoded in `internal/persona/persona.go`. Do not generate new ones.
9. **Registration prompt delay is ~3 seconds** after terminal ready. Do not inject it immediately.
10. **Input protection guard is 10 seconds.** Do not shorten or remove it.
11. **Shell agents use deferred staggered startup** on restore. Main shells before companions.
12. **Do not add features not in `REQUIREMENTS.md`** without updating the requirements file first.
13. **Write tests for all business logic** in `internal/` packages. UI code (in `internal/ui/`) does not require unit tests but should be manually verified.
14. **Phase order is strict.** Do not start Phase N+1 work until Phase N tests pass.

---

## Dependency List (go.mod)

```
require (
    fyne.io/fyne/v2          latest
    github.com/creack/pty    latest   // PTY process spawning
    github.com/fsnotify/fsnotify latest // File watching
    github.com/yuin/goldmark  latest   // Markdown → HTML
)
```

All other needs (HTTP server, JSON, logging, sync primitives) are covered by the Go standard library.

---

## Decision Log

| Date | Decision | Reason |
|---|---|---|
| 2026-03-06 | Use Fyne v2 for GUI | Pure Go, avoids CGo complexity for UI layer, cross-platform |
| 2026-03-06 | VTE for terminal widget | Industry-standard Linux terminal; GPU-accelerated path available |
| 2026-03-06 | Persist to `~/.config/skwad/` | XDG-compliant, standard for Linux desktop apps |
| 2026-03-06 | MCP server on `net/http` stdlib | No external framework needed; JSON-RPC 2.0 is simple to implement |
| 2026-03-06 | Mermaid via embedded WebView | Avoids `mmdc` binary dependency; renders identically to Mac version |

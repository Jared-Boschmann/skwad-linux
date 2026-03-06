# Skwad Linux — Requirements & Functionality Specification

A Go-based Linux desktop application that replicates all functionality from the Skwad macOS app.
The target platform is Linux desktop (X11/Wayland) using a cross-platform GUI toolkit.

---

## 1. Overview

Skwad is a multi-agent orchestration desktop app. It runs multiple AI coding agent CLIs simultaneously, each in its own embedded terminal, and lets them coordinate with each other via a built-in MCP (Model Context Protocol) HTTP server. Users manage agents across named workspaces, view git status, preview markdown/Mermaid diagrams, and optionally use voice input or an autopilot service to manage agent interactions.

---

## 2. Core Concepts

### 2.1 Agent

An agent is a running AI CLI (or plain shell) inside an embedded terminal. Each agent has:

| Field | Type | Description |
|---|---|---|
| `id` | UUID | Stable identifier, persisted across restarts |
| `name` | string | Display name |
| `avatar` | string | Emoji or base64-encoded PNG (`data:image/png;base64,...`) |
| `folder` | string | Working directory the agent operates in |
| `agentType` | string | One of: `claude`, `codex`, `opencode`, `gemini`, `copilot`, `custom1`, `custom2`, `shell` |
| `shellCommand` | string | Custom command for shell agent type |
| `personaId` | UUID | Optional reference to a persona (system prompt modifier) |
| `createdBy` | UUID | ID of the agent that created this one (nil = created by user) |
| `isCompanion` | bool | Companion agents are hidden in the sidebar; their lifecycle is tied to their creator |

**Runtime-only state (not persisted):**
- `status`: `idle` (green), `running` (orange), `input` (red), `error` (red)
- `isRegistered`: true after calling MCP `register-agent`
- `sessionId`: the AI session ID reported via hooks
- `resumeSessionId`: session to resume or fork at next launch
- `terminalTitle`: live title from terminal escape sequences
- `gitStats`: `{ insertions, deletions, files }` from `git diff --numstat`
- `markdownFilePath`: path currently shown in markdown preview panel
- `markdownMaximized`: whether markdown panel is maximized
- `markdownFileHistory`: list of recently shown markdown files (most recent first)
- `mermaidSource` / `mermaidTitle`: diagram source and optional title
- `metadata`: key-value map populated by hooks (e.g., `cwd`, `model`, `transcript_path`)
- `workingFolder`: hook-reported `cwd` if it differs from `folder` (used when agent changed to a different worktree root)

### 2.2 Agent Types

| Type | Command | Notes |
|---|---|---|
| `claude` | `claude` | Supports hook-based activity detection, system prompt, session resume/fork |
| `codex` | `codex` | Supports hook-based activity detection, system prompt, session resume/fork (`codex resume <id>`, `codex fork <id>`) |
| `opencode` | `opencode` | No system prompt support |
| `gemini` | `gemini` | Session resume via `--resume`, no system prompt |
| `copilot` | `gh copilot` | Session resume via `--interactive`, no system prompt |
| `custom1` / `custom2` | User-configured | Configurable command + options |
| `shell` | User-configured | Plain terminal, no AI, no activity tracking |

### 2.3 Workspace

A workspace groups agents together with their own layout state.

| Field | Type | Description |
|---|---|---|
| `id` | UUID | Stable identifier |
| `name` | string | Display name |
| `colorHex` | string | Accent color from a predefined palette |
| `agentIds` | []UUID | Ordered list of agents in this workspace |
| `layoutMode` | enum | `single`, `splitVertical`, `splitHorizontal`, `threePane`, `gridFourPane` |
| `activeAgentIds` | []UUID | Agents currently displayed in panes (1-4) |
| `focusedPaneIndex` | int | Index of focused pane |
| `splitRatio` | float | Primary split divider position (0.0-1.0) |
| `splitRatioSecondary` | float | Secondary split divider (grid/three-pane) |

**Workspace colors (predefined palette):** blue, purple, magenta, lavender, pink, rosePink, skyBlue, cyan, aqua, lime, green, teal, mauve, coral, red, orange, amber, tan.

---

## 3. Feature Modules

### 3.1 Workspace Management

- Create, rename, delete, and reorder workspaces
- Each workspace has a name and color accent
- Workspace switcher bar always visible on the left edge
- Switching workspaces saves the current layout state first
- Workspace status badge shows the "worst" status across all agents (input > running > nil)
- Agent can be moved from one workspace to another
- Auto-create a default "Skwad" workspace when the first agent is added

### 3.2 Agent Lifecycle

- **Create agent**: pick folder/worktree, name, avatar (emoji or image), agent type, optional persona
  - Insert after a specific agent (positional insertion)
  - Agents can also be created programmatically via MCP tool `create-agent`
- **Remove agent**: terminates companions first, unregisters from MCP, updates layout
- **Restart agent**: keeps same ID, increments a restart token to force terminal recreation; clears `sessionId`
- **Resume session**: pass an existing `sessionId` to the agent's CLI via `--resume <id>` (or `codex resume <id>`)
- **Fork session**: resume + pass `--fork-session` (or `codex fork <id>`) to branch from an existing session
- **Duplicate agent**: creates a copy with `" (copy)"` suffix, same folder and type
- **Reorder agents**: drag-and-drop within the sidebar
- **Add to Bench / Deploy from Bench**: user can save agent configurations as templates ("bench agents") and deploy them into the active workspace

### 3.3 Terminal Management

Each agent gets a persistent embedded terminal. Requirements:

- Terminals must remain alive (not destroyed) when switching between agents; use z-ordering / opacity toggle
- Terminal engine should be pluggable. Primary target: a GPU-accelerated terminal (e.g., VTE, Alacritty library, or equivalent). Fallback: VTE with standard rendering
- Each terminal runs the user's default shell, then sends `cd '<folder>' && clear && <agent-command>`
- Commands are prefixed with a space to avoid shell history pollution (`HISTCONTROL=ignorespace`)
- Agent ID injected as `SKWAD_AGENT_ID=<uuid>` environment variable
- Terminal title updates propagate to the agent's `terminalTitle` field (strip leading spinner/status characters)
- Keyboard focus management: focus correct terminal when switching panes

**Text injection:**
- `sendText(text)`: send text without Return
- `sendReturn()`: send Return key
- `injectText(text)`: send text + Return
- **Input protection**: user keypresses activate a 10-second guard that blocks automatic `injectText` delivery; queued messages are delivered after the guard expires or on next idle

### 3.4 Layout System

Five layout modes for the main terminal area:

| Mode | Panes | Description |
|---|---|---|
| `single` | 1 | Full area, one agent |
| `splitVertical` | 2 | Left \| Right, draggable divider |
| `splitHorizontal` | 2 | Top / Bottom, draggable divider |
| `threePane` | 3 | Left full-height \| Right-top / Right-bottom |
| `gridFourPane` | 4 | 2x2 grid |

**Split dividers** are draggable (ratio stored per workspace).

**Companion layout**: when selecting an agent that has companions, automatically enter split mode showing the agent and its companions.

**Auto-layout transitions:**
- Adding a companion to a single-pane view enters `splitVertical`
- Adding to a split view enters `threePane`
- Adding to three-pane enters `gridFourPane`
- Removing from a pane collapses the layout as needed

### 3.5 Sidebar

- Lists non-companion agents for the current workspace
- Shows agent name, avatar, status indicator (color dot), terminal title, git stats (`+N -N` in N files)
- Compact mode when sidebar width < 160px (icons only)
- Resizable sidebar (drag handle, 80px–400px)
- Collapsible sidebar (toggle)
- Context menu per agent:
  - Start split / enter split mode with this agent
  - Restart
  - Resume session (if agent supports conversation history)
  - Fork session (if supported)
  - Edit (name, avatar, folder)
  - Duplicate
  - Add shell companion
  - Add to Bench
  - Register (force re-inject registration prompt)
  - Move to workspace (submenu)
  - Close

### 3.6 MCP Server

An in-process HTTP server implementing the MCP (Model Context Protocol) JSON-RPC 2.0 protocol.

**Default port:** 8766 (configurable in settings)
**Endpoint:** `POST /mcp`

**Session management:** tracks client sessions, dispatches tool calls.

**Available tools:**

| Tool | Description |
|---|---|
| `register-agent` | Register agent with Skwad crew. Returns skwad member list and unread message count |
| `list-agents` | List all registered agents with name, folder, status |
| `send-message` | Send message to another agent by name or ID |
| `check-messages` | Read inbox messages, optionally mark as read |
| `broadcast-message` | Send message to all registered agents |
| `list-repos` | List all git repos in the configured source folder |
| `list-worktrees` | List worktrees for a given repo path |
| `create-agent` | Create a new agent (optionally with new worktree), can mark as companion |
| `close-agent` | Close an agent created by the caller |
| `create-worktree` | Create a new git worktree from a repo |
| `display-markdown` | Show a markdown file in the preview panel |
| `view-mermaid` | Render a Mermaid diagram in the preview panel |

**Agent coordinator (actor-safe message queue):**
- Maintains agent registry (registered agents only)
- Per-agent message inbox (in-memory, lost on restart)
- Deduplication: last-notified message ID per agent to avoid double-notifying
- When agent goes idle, checks for unread messages and injects notification text

**Registration flow:**
1. Terminal starts
2. After ~3 seconds, inject registration prompt with agent ID
3. Agent calls `register-agent` tool with its UUID + optional session ID
4. `isRegistered` flag set to true; session ID stored

**Hook support (claude, codex):**
- Agents emit lifecycle events via plugin/hook scripts
- Hook events update agent status (running/idle/blocked/error) and metadata
- Hook-based agents: `claude` and `codex`
- Hook data populates `metadata` map (keys: `cwd`, `model`, `transcript_path`, etc.)
- Hooks call a notify script with `SKWAD_AGENT_ID` to identify the agent

### 3.7 Activity Detection

Each agent has an `ActivityTracking` mode:

| Mode | Used by | Behavior |
|---|---|---|
| `.all` | non-hook agents | Terminal output + user input drive running/idle transitions |
| `.userInput` | hook-managed agents | Only user input tracked locally; hooks drive running/idle |
| `.none` | shell agents | No status tracking |

**Status state machine:**
- `idle` → `running`: terminal output activity or hook event
- `running` → `idle`: no output for idle timeout (default: 5s for non-hook, longer fallback for hook agents)
- Any → `input` (blocked): hook signals permission prompt; only Return (→ running) or Escape (→ idle) unblocks
- Any → `error`: hook signals error condition

**Idle timeout:** configurable per agent type via timing constants.

### 3.8 Git Integration

Git panel slides in from the right side of the active terminal.

**Operations:**
- View current branch, upstream tracking info, ahead/behind counts
- View modified/staged/untracked files list
- Click file to view diff (syntax-highlighted, with `+`/`-` colored lines)
- Stage individual files, stage all
- Unstage individual files, unstage all
- Commit with message (text input)
- Discard changes (unstaged files)

**Git diff viewer:**
- Shows hunk headers, context lines, additions (+, green), deletions (-, red)
- Toggle staged vs. unstaged diff

**Git stats in sidebar:**
- Shows combined `insertions + deletions + files` count per agent when idle
- Refreshed automatically when agent transitions to idle

**File watcher:**
- FSEvent/inotify monitoring of the agent's folder
- Debounced auto-refresh of git status

**Worktree support:**
- List worktrees for a repo (`git worktree list`)
- Create new worktree (branch name + destination path)
- `suggestedWorktreePath`: derives sibling path from repo path + branch name

**Repo discovery:**
- Background scan of configured source base folder
- Discovers all git repositories
- Auto-detects source folder on first launch from common locations: `~/src`, `~/dev`, `~/code`, `~/projects`, `~/repos`, `~/source`, `~/sources`, `~/workspace`, `~/workspaces`, `~/git`, `~/github`, `~/Development`, `~/work`, `~/coding`

### 3.9 Conversation History

View and manage past agent sessions without leaving the app.

**Supported agent types:** claude, codex, gemini, copilot

**Operations:**
- List sessions sorted by date descending (up to 20)
- Session summary shows: title (first meaningful user message), timestamp, message count
- Delete a session (removes session files)
- Resume a session (passes session ID to agent CLI)
- Fork a session (branch from an existing session)

Each agent type has its own history provider that knows where to find and parse session files.

### 3.10 Markdown Preview Panel

- Slides in from the right of the terminal area (or maximized)
- Renders markdown with themed styling (font size configurable, dark mode aware)
- Can be opened by MCP tool `display-markdown` or manually
- File path history: tracks recently shown files, navigable
- User can highlight text in the preview and submit it as a review comment back to the agent (text injected into agent terminal)
- Maximized mode: fills available space

### 3.11 Mermaid Diagram Panel

- Renders Mermaid diagram source natively
- Shown alongside markdown panel or standalone
- Supports: flowcharts (`graph TD/LR`), state diagrams, sequence diagrams, class diagrams, ER diagrams
- Configurable theme (`auto`, `light`, `dark`) and scale (zoom)
- Limitations: no HTML node labels, no tooltips, no `<br>` multiline labels, no `style`/`linkStyle` directives, no subgraph styling

### 3.12 File Finder

- Fuzzy file search within the active agent's folder
- Keyboard-invocable (Cmd+P equivalent)
- Uses `git ls-files` for git repos, filesystem enumeration otherwise
- Excludes: `.git`, `node_modules`, `.build`, `__pycache__`, `.DS_Store`, `.svn`, `.hg`, `Pods`, `DerivedData`
- Max 50,000 files indexed; max 50 results returned
- Results ranked by fuzzy score with match index highlighting
- Opens selected file in the configured editor app

### 3.13 Autopilot

An optional feature that uses an LLM to analyze agent messages and take automated actions.

**Trigger:** when a hook-managed agent (claude, codex) transitions to idle, the last terminal message is analyzed.

**Classification (three categories):**
- `completed`: agent finished work, no input needed
- `binary`: agent is asking for simple yes/no approval
- `open`: agent is asking an open-ended question

**Configured actions:**
- `mark`: set agent status to `input`, send desktop notification
- `ask`: mark + show a decision sheet with the last message and options (approve / decline / custom reply)
- `continue`: if `binary`, auto-inject "yes, continue"; if `open`, fall back to `mark`
- `custom`: user-provided prompt fed to the LLM with the agent's message; the LLM response is injected directly

**Supported AI providers:** OpenAI (`gpt-5-mini`), Anthropic (`claude-haiku-4-5`), Google Gemini (`gemini-flash-lite-latest`)

**Configuration:**
- Provider (openai / anthropic / google)
- API key
- Action (mark / ask / continue / custom)
- Custom prompt (for `custom` action)

### 3.14 Voice Input

- Push-to-talk voice input using OS speech recognition
- Configurable push-to-talk key (default: Right Shift)
- Auto-insert transcription into focused agent terminal when key released
- Visual overlay during recording (waveform display)
- Configurable: enable/disable, engine (only "apple" / OS-native currently), push-to-talk key, auto-insert toggle

### 3.15 Personas

Named system prompt modifiers that change how an agent behaves.

**Fields:** id, name, instructions, type (`system`/`user`), state (`enabled`/`disabled`/`deleted`)

**System-shipped personas (fixed UUIDs for cross-install consistency):**
- Kent Beck — TDD, simplest code, red-green-refactor
- Martin Fowler — readability, design patterns, continuous refactoring
- Linus Torvalds — simplicity, performance, pragmatism
- Uncle Bob — SOLID principles, clean code
- John Carmack — deep technical focus, linear code, hardware awareness
- Dave Farley — continuous delivery, testing, incremental steps

**Operations:**
- Create, edit, delete user personas
- Soft-delete system personas (mark as deleted, not removable)
- Restore default system personas to original state
- Assign persona to an agent at creation time or via edit
- Persona instructions injected into agent's system prompt at launch

### 3.16 Bench (Agent Templates)

- User saves agents as templates ("bench agents")
- Bench is a persistent list of agent configurations (name, avatar, folder, type, command, persona)
- Deploy a bench agent into the current workspace (validates folder exists first, removes from bench if invalid)
- Rename bench entries
- Remove bench entries

### 3.17 Notifications

- Desktop notifications when an agent enters `input` (blocked/awaiting) status
- Notification includes agent name and a short message
- Configurable: enable/disable

### 3.18 Settings

**General:**
- Restore layout on launch (bool)
- Keep in menu bar / system tray (bool)
- Terminal engine selection (primary / fallback)
- Source base folder path

**Terminal:**
- Font name and size
- Background color (hex)
- Foreground color (hex)

**Coding:**
- Default "open with" app (Cmd+Shift+O shortcut)
- Per-agent-type CLI options (extra flags passed to agent commands)
- Custom agent 1: command + options
- Custom agent 2: command + options

**MCP Server:**
- Enable/disable
- Port number (default 8766)

**Autopilot:** (see section 3.13)

**Voice Input:** (see section 3.14)

**Personas:** (see section 3.15)

**Appearance:**
- Mode: `auto` (derives from terminal background luminance), `system`, `light`, `dark`
- Markdown font size
- Mermaid theme (auto/light/dark) and scale

---

## 4. Persistence

All settings and state persisted across launches:

| Data | Storage |
|---|---|
| Agents | JSON file (`savedAgents`) |
| Workspaces + layout | JSON file (`savedWorkspaces`) |
| Current workspace ID | Config entry |
| Bench agents | JSON file |
| Personas | JSON file |
| App settings | Config file (equivalent of UserDefaults) |
| Recent repos | JSON list (last 5) |

**Migration notes from Mac version:**
- Agents without `isCompanion`/`createdBy` default to `false`/`nil`
- Agents without `agentType` default to `"claude"`
- Workspaces without `splitRatioSecondary` default to `0.5`
- If agents exist but no workspaces, auto-create default "Skwad" workspace

---

## 5. Terminal Command Construction

Agent launch commands follow this pattern:

```
 cd '<folder>' && clear && SKWAD_AGENT_ID=<uuid> <agent-cmd> [resume-flags] [user-opts] [mcp-flags] [registration-flags]
```

**MCP flags per agent type:**

| Type | MCP Config Flag |
|---|---|
| `claude` | `--mcp-config '{"mcpServers":{"skwad":{"type":"http","url":"<url>"}}}' --allowed-tools 'mcp__skwad__*' --plugin-dir "<plugin-path>"` |
| `codex` | `-c 'notify=["bash","<notify-script>"]'` |
| `gemini` | `--allowed-mcp-server-names skwad` |
| `copilot` | `--additional-mcp-config '...' --allow-tool 'skwad(<tool>)'` (for each tool) |

**Registration flags per agent type:**

| Type | System Prompt Flag | User Prompt Flag |
|---|---|---|
| `claude` | `--append-system-prompt "<prompt>"` | Passed as last positional arg (skipped on resume) |
| `codex` | `-c 'developer_instructions="<prompt>"'` | Passed as last positional arg (skipped on resume) |
| `opencode` | N/A | `--prompt "<prompt>"` |
| `gemini` | N/A | `--prompt-interactive "<prompt>"` |
| `copilot` | N/A | `--interactive "<prompt>"` |

**Shell escaping:** backslash, double quote, `$`, backtick, `!` are all escaped inside double-quoted strings.

---

## 6. Deferred Shell Agent Startup

When multiple shell agents are restored from persistence on app launch, they use a staggered startup queue to avoid overwhelming the system:

1. Show a "Starting soon..." banner in the terminal immediately
2. Main shell agents queue before companion shells
3. Initial delay before first agent starts (lets AI agents settle)
4. Each subsequent shell agent starts after a stagger delay

---

## 7. Keyboard Shortcuts

| Action | Shortcut |
|---|---|
| New agent | Cmd+N |
| Toggle git panel | Cmd+G |
| Toggle sidebar | Cmd+\ |
| Toggle file finder | Cmd+P |
| Next agent | Cmd+] |
| Previous agent | Cmd+[ |
| Select agent 1-9 | Cmd+1 through Cmd+9 |
| Next workspace | Cmd+Shift+] |
| Previous workspace | Cmd+Shift+[ |
| Switch to workspace N | Cmd+Ctrl+1 through Cmd+Ctrl+9 |
| Open with default editor | Cmd+Shift+O |
| Fork agent | available from context menu |

---

## 8. UI Layout

```
+---+----------+----------------------------------+
| W |          |                                  |
| o | Sidebar  |   Terminal Pane(s)               |
| r | (agent   |   (1, 2, 3, or 4 split panes)   |
| k | list)    |                    +------------+ |
| s |          |                    | Markdown / | |
| p |          |                    | Mermaid    | |
| a |          |                    | Panel      | |
| c |          |                    +------------+ |
| e |          +----------------------------------+ |
| B |          | Git Panel (slides up from bottom)|
| a |          +----------------------------------+
| r |
+---+
```

- **Workspace bar**: vertical strip on far left, one badge per workspace
- **Sidebar**: resizable, collapsible, shows agent list
- **Terminal area**: main content, supports split layouts
- **Markdown/Mermaid panel**: right-side sliding panel within terminal area
- **Git panel**: bottom-sliding panel below the active terminal pane

---

## 9. Go Implementation Notes

### Recommended Libraries

| Need | Candidate |
|---|---|
| GUI framework | [Fyne](https://fyne.io/) or [GTK4 via gotk4](https://github.com/diamondburned/gotk4) |
| Terminal emulation | VTE (via CGo bindings) or [tcell](https://github.com/gdamore/tcell) for TUI fallback |
| HTTP server (MCP) | `net/http` (stdlib) |
| JSON-RPC | Manual or `github.com/gorilla/rpc` |
| File watching | [fsnotify](https://github.com/fsnotify/fsnotify) |
| Speech recognition | OS-level (via CGo or `exec.Command` to external tool) |
| Markdown rendering | [goldmark](https://github.com/yuin/goldmark) + HTML render |
| Mermaid | Embed a WebView with mermaid.js, or shell out to `mmdc` CLI |
| Persistence | JSON files in `~/.config/skwad/` |
| Fuzzy search | Implement FuzzyScorer (scored character-match algorithm) |

### Architecture Recommendations

- Central `AgentManager` struct (thread-safe, single source of truth)
- `AgentCoordinator` as a goroutine-safe actor (channel-based or mutex-protected) for MCP message queue
- Separate goroutine per agent terminal session
- MCP server runs as a goroutine with `net/http`
- File watchers per git repo with debounced refresh
- Settings stored in `~/.config/skwad/config.json` (or equivalent XDG path)
- Agent state (non-terminal) stored in `~/.config/skwad/agents.json`
- Workspaces stored in `~/.config/skwad/workspaces.json`
- Personas stored in `~/.config/skwad/personas.json`

### Key Behavioral Invariants to Preserve

1. Terminal instances are never destroyed on agent switch — only hidden/shown
2. Agent IDs are stable across restarts (persisted)
3. `restartToken` pattern: increment on restart to force terminal re-creation while keeping the same agent ID
4. MCP messages are in-memory only (intentionally, lost on restart)
5. Registration prompt injected ~3 seconds after terminal starts (not immediately)
6. Input protection: user keypress blocks auto-inject for 10 seconds
7. Shell agents use deferred/staggered startup on restore
8. Companion agents are invisible in the sidebar; their layout is linked to their creator
9. Personas with fixed UUIDs must keep those exact UUIDs for cross-install compatibility

---

## 10. Known Limitations from Mac Version (to address or carry forward)

- MCP messages are in-memory only (lost on app restart) — carry forward by design
- Single window only — carry forward initially
- Terminal colors set by agent CLIs may override app theme colors
- Ghostty-specific: reads `~/.config/ghostty/config` for background color — Linux equivalent: read terminal config or provide a color picker

---

## 11. Out of Scope (Future / Not Required for v1)

- Split pane view for multiple agents (partially implemented on Mac — the layout system is required, see section 3.4)
- Agent templates/presets (Bench is required; "templates" as a richer concept is future)
- GitHub PR integration
- Multi-window support
- Cloud sync of agents/settings

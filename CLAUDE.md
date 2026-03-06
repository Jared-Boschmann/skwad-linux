@DEVPLAN.md

# Skwad Linux

## Purpose

This project ports the Skwad macOS application to a Linux desktop application written in Go. Skwad is a multi-agent orchestration tool — it runs multiple AI coding agent CLIs (Claude Code, Codex, Gemini CLI, GitHub Copilot, OpenCode, or custom) simultaneously, each in its own embedded terminal, and enables them to coordinate work via a built-in MCP (Model Context Protocol) server.

The goal is a native Linux desktop app that is functionally identical to the Mac version, cross-platform (X11 and Wayland), and built entirely in Go.

## What This App Does

- Runs multiple AI coding agents in parallel, each in a persistent embedded terminal
- Organizes agents into named, color-coded workspaces
- Supports 1, 2, 3, and 4-pane split layouts so multiple agents are visible at once
- Provides an in-process MCP HTTP server so agents can register, message each other, query worktrees, and spawn new agents programmatically
- Integrates with git: diff viewer, staging, committing, worktree creation, and live git stats per agent
- Renders markdown files and Mermaid diagrams in side panels on demand
- Includes a fuzzy file finder, conversation history browser, agent personas, and an autopilot service that uses an LLM to automatically handle agent prompts

## Key Documents

- `REQUIREMENTS.md` — complete feature specification derived from the Mac source code; the source of truth for what must be built
- `DEVPLAN.md` — phased build plan, module layout, technology decisions, coding rules, and test checklist for agents working on this codebase

## Technology

- **Language:** Go 1.22+
- **GUI:** Fyne v2
- **Terminal widget:** VTE (via CGo)
- **MCP server:** `net/http` stdlib, JSON-RPC 2.0
- **File watching:** fsnotify
- **Markdown:** goldmark → HTML → WebView
- **Mermaid:** embedded mermaid.js in WebView
- **Persistence:** JSON files in `~/.config/skwad/`

## Development Approach

Follow the six phases defined in `DEVPLAN.md` in strict order. Each phase has an exit criteria that must pass before the next phase begins. The coding rules in `DEVPLAN.md` are hard requirements — particularly around terminal persistence, stable agent IDs, thread safety, and MCP message handling.

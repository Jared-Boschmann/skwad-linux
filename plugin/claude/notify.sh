#!/usr/bin/env bash
# Claude Code hook plugin for Skwad Linux.
# Installed via: claude --plugin-dir <path>
#
# Claude passes the hook event as JSON on stdin.
# We forward it to the Skwad hook endpoint, injecting SKWAD_AGENT_ID.

set -euo pipefail

SKWAD_PORT="${SKWAD_MCP_PORT:-8766}"
SKWAD_HOOK_URL="http://127.0.0.1:${SKWAD_PORT}/hook"

# Read full stdin (the hook event JSON from Claude).
EVENT_JSON=$(cat)

# Inject agentId and hook_event_name into the payload.
HOOK_NAME="${CLAUDE_HOOK_NAME:-}"
AGENT_ID="${SKWAD_AGENT_ID:-}"

PAYLOAD=$(printf '%s' "$EVENT_JSON" | \
  python3 -c "
import sys, json
data = json.load(sys.stdin)
data['agentId'] = '$AGENT_ID'
if '$HOOK_NAME':
    data['hook_event_name'] = '$HOOK_NAME'
print(json.dumps(data))
" 2>/dev/null || echo "{\"agentId\":\"$AGENT_ID\",\"hook_event_name\":\"$HOOK_NAME\"}")

curl -s -X POST \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" \
  "$SKWAD_HOOK_URL" \
  --max-time 2 \
  --silent \
  --output /dev/null || true

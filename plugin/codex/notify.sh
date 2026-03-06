#!/usr/bin/env bash
# Codex hook plugin for Skwad Linux.
# Configured via: codex -c 'notify=["bash","<path>/notify.sh"]'
#
# Codex calls the notify script with the event type as $1 and
# optional JSON context on stdin.

set -euo pipefail

SKWAD_PORT="${SKWAD_MCP_PORT:-8766}"
SKWAD_HOOK_URL="http://127.0.0.1:${SKWAD_PORT}/hook"

EVENT_TYPE="${1:-}"
AGENT_ID="${SKWAD_AGENT_ID:-}"

# Read optional JSON context from stdin (may be empty).
CONTEXT_JSON=$(cat 2>/dev/null || echo "{}")

PAYLOAD=$(printf '%s' "$CONTEXT_JSON" | \
  python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
except Exception:
    data = {}
data['agentId'] = '$AGENT_ID'
data['eventType'] = '$EVENT_TYPE'
print(json.dumps(data))
" 2>/dev/null || echo "{\"agentId\":\"$AGENT_ID\",\"eventType\":\"$EVENT_TYPE\"}")

curl -s -X POST \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" \
  "$SKWAD_HOOK_URL" \
  --max-time 2 \
  --silent \
  --output /dev/null || true

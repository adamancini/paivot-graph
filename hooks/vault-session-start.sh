#!/usr/bin/env bash
# vault-session-start.sh -- Consult the Obsidian vault for project context on session start.
#
# Reads the SessionStart hook JSON from stdin, extracts cwd, detects the project name,
# searches the vault, and outputs relevant context to stdout (injected into Claude's awareness).
# Reads the operating mode from the vault note (or uses static fallback).
#
# Always exits 0 -- never blocks session start.

set -euo pipefail

VAULT_DIR="$HOME/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"

# ---------------------------------------------------------------------------
# 1. Read hook input and extract cwd
# ---------------------------------------------------------------------------
hook_input="$(cat)"
cwd="$(printf '%s' "$hook_input" | python3 -c "import sys,json; print(json.load(sys.stdin).get('cwd',''))" 2>/dev/null || echo "")"

if [ -z "$cwd" ]; then
    cwd="$(pwd)"
fi

# ---------------------------------------------------------------------------
# 2. Detect project name (git remote basename > directory name)
# ---------------------------------------------------------------------------
project=""
if [ -d "$cwd/.git" ] || git -C "$cwd" rev-parse --git-dir >/dev/null 2>&1; then
    remote_url="$(git -C "$cwd" remote get-url origin 2>/dev/null || echo "")"
    if [ -n "$remote_url" ]; then
        project="$(basename "$remote_url" .git)"
    fi
fi

if [ -z "$project" ]; then
    project="$(basename "$cwd")"
fi

# ---------------------------------------------------------------------------
# 3. Check vault directory exists
# ---------------------------------------------------------------------------
if [ ! -d "$VAULT_DIR" ]; then
    echo "[VAULT] Vault directory not found -- vault consultation skipped."
    exit 0
fi

# ---------------------------------------------------------------------------
# 4. Search vault for project context (grep for project name in .md files)
# ---------------------------------------------------------------------------
search_results=""
if command -v rg >/dev/null 2>&1; then
    search_results="$(rg -l --type md "$project" "$VAULT_DIR" 2>/dev/null \
        | sed "s|$VAULT_DIR/||" \
        || echo "")"
else
    search_results="$(grep -rl "$project" "$VAULT_DIR" --include='*.md' 2>/dev/null \
        | sed "s|$VAULT_DIR/||" \
        || echo "")"
fi

# Trim whitespace
search_results="$(printf '%s' "$search_results" | sed '/^$/d')"

if [ -z "$search_results" ]; then
    search_results="(none found -- this is a new project to the vault)"
fi

# ---------------------------------------------------------------------------
# 5. Output structured context
# ---------------------------------------------------------------------------
cat <<CONTEXT
[VAULT] Project: $project
Relevant vault notes:

$search_results

CONTEXT

# ---------------------------------------------------------------------------
# 6. Read operating mode from vault (or use static fallback)
# ---------------------------------------------------------------------------
mode_file="$VAULT_DIR/conventions/Session Operating Mode.md"
operating_mode=""
if [ -f "$mode_file" ]; then
    operating_mode="$(cat "$mode_file")"
fi

if [ -n "$operating_mode" ]; then
    echo "[VAULT] Operating mode for this session (from vault):"
    echo ""
    echo "$operating_mode"
else
    # Static fallback
    cat <<'MODE'
[VAULT] Operating mode for this session:

BEFORE STARTING: Read the vault notes listed above. Do not rediscover what is already known.
  Use the Read tool on the vault file path shown above.

WHILE WORKING: Capture knowledge as it emerges -- do not wait for the end.
  - After making a decision (chose X over Y): create a decision note
  - After solving a non-obvious bug: create a debug note
  - After discovering a reusable pattern: create a pattern note
  Use the Write tool to create files in: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/_inbox/

BEFORE ENDING: Update the project index note with what was accomplished.
  Append to the project note in: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/projects/

This is not optional. Knowledge that is not captured is knowledge that will be rediscovered at cost.
MODE
fi

exit 0

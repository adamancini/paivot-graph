#!/usr/bin/env bash
# vault-pre-compact.sh -- Remind Claude to capture knowledge before context compaction.
#
# Reads the Pre-Compact Checklist from the vault (or uses static fallback).
# This is the last chance to save what was learned before memory is lost.
# Outputs a structured reminder to stdout. Always exits 0.

set -euo pipefail

VAULT_DIR="$HOME/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"
TODAY="$(date +%Y-%m-%d)"

# ---------------------------------------------------------------------------
# Try to read the checklist from the vault
# ---------------------------------------------------------------------------
checklist=""
checklist_file="$VAULT_DIR/conventions/Pre-Compact Checklist.md"
if [ -f "$checklist_file" ]; then
    checklist="$(cat "$checklist_file")"
fi

# ---------------------------------------------------------------------------
# Output checklist (vault or fallback)
# ---------------------------------------------------------------------------
if [ -n "$checklist" ]; then
    echo "[VAULT] Context compaction imminent -- capture knowledge now."
    echo ""
    echo "$checklist"
else
    # Static fallback
    cat <<EOF
[VAULT] Context compaction imminent -- capture knowledge now.

Before this context is compacted, save anything worth remembering:

1. DECISIONS made this session (with rationale and alternatives considered):
   Use Write tool to create: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/_inbox/<Decision Title>.md

2. PATTERNS discovered (reusable solutions):
   Use Write tool to create: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/_inbox/<Pattern Name>.md

3. DEBUG INSIGHTS (problems solved):
   Use Write tool to create: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/_inbox/<Bug Title>.md

4. PROJECT UPDATES (progress, state changes):
   Use Edit tool to append to: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/projects/<Project>.md

All notes must have frontmatter: type, project, status, created ($TODAY).

Do this NOW -- after compaction, the details will be lost.
EOF
fi

exit 0

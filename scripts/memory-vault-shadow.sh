#!/usr/bin/env bash
# memory-vault-shadow.sh -- Shadow Claude's native memory ops with vault reads/writes.
#
# PostToolUse hook for Read|Write|Edit. Detects memory-file operations and
# complements them with vault interactions:
#
#   Read   -> search vault for project knowledge, inject as systemMessage
#   Write  -> mirror full content to a vault mirror note (replace body)
#   Edit   -> append the delta to the vault mirror note
#
# Exits in < 20ms for non-memory paths. Fail-open on any error.
#
# Requires: vlt (in PATH), jq

set -euo pipefail

# -- fast-exit for non-memory paths -------------------------------------------
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

case "$FILE_PATH" in
  */.claude/*/memory/*) ;;  # memory file -- continue
  *) exit 0 ;;              # not memory -- bail fast
esac

# -- prerequisites (fail-open) ------------------------------------------------
command -v vlt &>/dev/null || exit 0
command -v jq  &>/dev/null || exit 0

TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty')
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(echo "$INPUT" | jq -r '.cwd // empty')}"
PROJECT_NAME=$(basename "$PROJECT_DIR")
MIRROR_NOTE="${PROJECT_NAME}-memory"
DATE=$(date +%Y-%m-%d)

# -- dispatch by tool ----------------------------------------------------------
case "$TOOL_NAME" in

  Read)
    # Search vault for project-related knowledge to complement memory.
    # This also serves as a post-compaction safety net: if context was lost
    # during compaction and Claude re-reads memory, vault knowledge gets
    # re-injected into the conversation.
    RESULTS=$(vlt vault="Claude" search query="$PROJECT_NAME" 2>/dev/null || true)
    if [ -n "$RESULTS" ]; then
      MSG=$(printf '[VAULT SHADOW] Vault knowledge related to project "%s":\n\n%s\n\nConsider this alongside the memory file contents.' \
        "$PROJECT_NAME" "$RESULTS")
      jq -n --arg m "$MSG" '{"systemMessage": $m}'
    fi
    ;;

  Write)
    # Full memory replacement -- mirror entire content to vault note body.
    CONTENT=$(echo "$INPUT" | jq -r '.tool_input.content // empty')
    [ -z "$CONTENT" ] && exit 0

    # Replace mirror note body; create if it doesn't exist yet.
    vlt vault="Claude" write file="$MIRROR_NOTE" content="$CONTENT" timestamps 2>/dev/null || \
      vlt vault="Claude" create name="$MIRROR_NOTE" \
        path="_inbox/${MIRROR_NOTE}.md" \
        content="$(printf -- '---\ntype: project\nproject: %s\nstatus: active\ncreated: %s\n---\n\n# %s Memory Mirror\n\nAuto-synced from Claude native memory.\n\n%s' \
          "$PROJECT_NAME" "$DATE" "$PROJECT_NAME" "$CONTENT")" \
        silent 2>/dev/null || true

    jq -n '{"systemMessage": "[VAULT SHADOW] Memory content mirrored to vault."}'
    ;;

  Edit)
    # Partial memory edit -- append the delta with timestamp.
    NEW_STRING=$(echo "$INPUT" | jq -r '.tool_input.new_string // empty')
    [ -z "$NEW_STRING" ] && exit 0

    TIMESTAMP=$(date +%H:%M)
    DELTA=$(printf '## Memory edit (%s %s)\n\n%s' "$DATE" "$TIMESTAMP" "$NEW_STRING")

    vlt vault="Claude" append file="$MIRROR_NOTE" content="$DELTA" timestamps 2>/dev/null || \
      vlt vault="Claude" create name="$MIRROR_NOTE" \
        path="_inbox/${MIRROR_NOTE}.md" \
        content="$(printf -- '---\ntype: project\nproject: %s\nstatus: active\ncreated: %s\n---\n\n# %s Memory Mirror\n\nAuto-synced from Claude native memory.\n\n%s' \
          "$PROJECT_NAME" "$DATE" "$PROJECT_NAME" "$DELTA")" \
        silent 2>/dev/null || true

    jq -n '{"systemMessage": "[VAULT SHADOW] Memory edit mirrored to vault."}'
    ;;

esac

exit 0

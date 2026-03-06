# Hook Variable Substitution: PostToolUse:Read Error Resolution

## Problem

Users were seeing "PostToolUse:Read hook error" messages during file reads, with 4-5 errors per read operation.

## Root Cause

The PostToolUse hook in `hooks.json` used:

```json
"command": "bash ${CLAUDE_PLUGIN_ROOT}/scripts/memory-vault-shadow.sh"
```

Claude Code's hook system does NOT expand shell variables like `${VAR}`. The command literally became:

```
bash /scripts/memory-vault-shadow.sh
```

This file does not exist, so the hook failed. The script's error handling (`exit 0`) made it fail silently, but Claude Code still reported the error.

## Solution

Changed the hook to dynamically discover the script location:

```json
"command": "bash -c 'SCRIPT=$(find ~/.claude/plugins/cache/paivot-graph -name \"memory-vault-shadow.sh\" -type f 2>/dev/null | head -1); [ -n \"$SCRIPT\" ] && bash \"$SCRIPT\" || true'"
```

This:
1. Finds the script in the plugin cache (where Claude Code installs plugins)
2. Runs the script if found
3. Silently succeeds if not found (`|| true`)
4. Works regardless of `CLAUDE_PLUGIN_ROOT` environment variable

## Why This Works

Claude Code installs plugins to `~/.claude/plugins/cache/<plugin-name>/<version>/`. The `make install` command copies the plugin there. The hook now searches for the script in this standard location instead of relying on environment variables.

## Testing

Tested with:
```bash
bash -c 'SCRIPT=$(find ~/.claude/plugins/cache/paivot-graph -name "memory-vault-shadow.sh" -type f 2>/dev/null | head -1); [ -n "$SCRIPT" ] && bash "$SCRIPT" <<< "{...json input...}"'
```

Result: Successfully executes without errors.

## Related

- `scripts/hook-diagnostics.sh` — Diagnostic tool to identify hook environment issues
- `hooks/hooks.json` — PostToolUse hook configuration
- `scripts/memory-vault-shadow.sh` — Memory sync to vault script

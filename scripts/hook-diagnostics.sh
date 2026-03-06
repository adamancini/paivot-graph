#!/usr/bin/env bash
# Diagnostic script to identify hook execution issues
# Run this to understand what's happening in the PostToolUse:Read hook

echo "=== Hook Diagnostics ==="
echo "Timestamp: $(date)"
echo "Shell: $SHELL"
echo "User: $(whoami)"
echo ""

echo "=== Environment Variables ==="
echo "CLAUDE_PLUGIN_ROOT: ${CLAUDE_PLUGIN_ROOT:-NOT SET}"
echo "PATH: $PATH"
echo ""

echo "=== Command Availability ==="
command -v bash && echo "bash: $(command -v bash)" || echo "bash: NOT FOUND"
command -v vlt && echo "vlt: $(command -v vlt)" || echo "vlt: NOT FOUND"
command -v jq && echo "jq: $(command -v jq)" || echo "jq: NOT FOUND"
echo ""

echo "=== Script Paths ==="
SCRIPT_PATH="${CLAUDE_PLUGIN_ROOT}/scripts/memory-vault-shadow.sh"
echo "Expected script path: $SCRIPT_PATH"
if [ -f "$SCRIPT_PATH" ]; then
  echo "Script exists: YES"
  echo "Script executable: $([ -x "$SCRIPT_PATH" ] && echo YES || echo NO)"
  echo "Script size: $(wc -c < "$SCRIPT_PATH") bytes"
else
  echo "Script exists: NO"
fi
echo ""

echo "=== Testing Script Execution ==="
TEST_INPUT='{"tool_name":"Read","tool_input":{"file_path":"/tmp/test.md"},"cwd":"/tmp"}'
echo "Test input: $TEST_INPUT"
echo ""
echo "Execution output:"
bash "$SCRIPT_PATH" <<< "$TEST_INPUT" 2>&1 | head -20
echo ""
echo "Exit code: $?"
echo ""

echo "=== Vault Connectivity ==="
if command -v vlt &>/dev/null; then
  echo "Testing vlt search..."
  vlt vault="Claude" search query="test" 2>&1 | head -5
  echo "vlt search exit code: $?"
else
  echo "vlt not available in hook environment"
fi

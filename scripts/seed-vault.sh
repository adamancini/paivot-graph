#!/usr/bin/env bash
# seed-vault.sh -- DEPRECATED: use seed-vault-direct.sh instead.
#
# This script uses the obsidian CLI which is slow and hangs on large content.
# Kept for reference only. The Makefile 'seed' target uses seed-vault-direct.sh.
#
# Original purpose: reads agent prompts from paivot-claude v1.38.0, the
# vault-knowledge skill, and behavioral content from the current plugin,
# then creates corresponding Obsidian vault notes via the obsidian CLI.
#
# Idempotent: checks if each note exists before creating. Re-running is safe.

set -euo pipefail

VAULT="Claude"
TODAY="$(date +%Y-%m-%d)"

# Source directories
AGENT_SRC="${AGENT_SRC:-$HOME/.claude/plugins/cache/paivot-claude/paivot/1.38.0/agents}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_DIR="$(dirname "$SCRIPT_DIR")"

created=0
skipped=0
failed=0

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

note_exists() {
    local name="$1"
    local output
    output="$(obsidian vault="$VAULT" read file="$name" 2>/dev/null)"
    # obsidian CLI exits 0 even on error; check output for "Error:" prefix
    [[ "$output" != *"Error:"* ]] && [ -n "$output" ]
}

# Extract the markdown body from a paivot-claude agent file (strip YAML frontmatter)
extract_body() {
    local file="$1"
    awk 'BEGIN{c=0} /^---[[:space:]]*$/{c++; next} c>=2{print}' "$file"
}

# Create a vault note from a temp file. Uses a temp file to avoid shell
# argument length limits with very long agent prompts.
create_note() {
    local vault_name="$1"
    local vault_path="$2"
    local tmp_file="$3"

    if note_exists "$vault_name"; then
        echo "  SKIP: $vault_name (already exists)"
        skipped=$((skipped + 1))
        return 0
    fi

    local content
    content="$(cat "$tmp_file")"

    if obsidian vault="$VAULT" create name="$vault_name" path="$vault_path" content="$content" silent 2>/dev/null; then
        echo "  CREATED: $vault_path"
        created=$((created + 1))
    else
        echo "  FAIL: $vault_path"
        failed=$((failed + 1))
    fi
}

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------

echo "paivot-graph vault seeder"
echo "========================="
echo ""

if ! command -v obsidian >/dev/null 2>&1; then
    echo "ERROR: obsidian CLI not found. Install: https://github.com/Acylation/obsidian-cli"
    exit 1
fi

if [ ! -d "$AGENT_SRC" ]; then
    echo "ERROR: Agent source not found at $AGENT_SRC"
    echo "Install paivot-claude v1.38.0 first, or set AGENT_SRC to the agents/ directory."
    exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

# ---------------------------------------------------------------------------
# 1. Agent prompts (8 agents)
# ---------------------------------------------------------------------------

echo "Seeding agent prompts..."

declare -A agent_map=(
    ["paivot-sr-pm"]="Sr PM Agent"
    ["paivot-pm"]="PM Acceptor Agent"
    ["paivot-developer"]="Developer Agent"
    ["paivot-architect"]="Architect Agent"
    ["paivot-designer"]="Designer Agent"
    ["paivot-business-analyst"]="Business Analyst Agent"
    ["paivot-anchor"]="Anchor Agent"
    ["paivot-retro"]="Retro Agent"
)

for slug in "${!agent_map[@]}"; do
    vault_name="${agent_map[$slug]}"
    src_file="$AGENT_SRC/${slug}.md"

    if [ ! -f "$src_file" ]; then
        echo "  WARN: $src_file not found, skipping $vault_name"
        skipped=$((skipped + 1))
        continue
    fi

    body="$(extract_body "$src_file")"

    cat > "$tmp_dir/note.md" <<AGENT_EOF
---
type: methodology
project: paivot
stack: [claude-code]
domain: developer-tools
status: active
created: $TODAY
---

$body
AGENT_EOF

    create_note "$vault_name" "methodology/${vault_name}.md" "$tmp_dir/note.md"
done

# ---------------------------------------------------------------------------
# 2. Skill content (vault-knowledge)
# ---------------------------------------------------------------------------

echo ""
echo "Seeding skill content..."

skill_src="$PLUGIN_DIR/skills/vault-knowledge/SKILL.md"
if [ -f "$skill_src" ]; then
    skill_body="$(extract_body "$skill_src")"

    cat > "$tmp_dir/note.md" <<SKILL_EOF
---
type: convention
project: paivot-graph
stack: [claude-code, obsidian]
domain: developer-tools
status: active
created: $TODAY
---

$skill_body
SKILL_EOF

    create_note "Vault Knowledge Skill" "conventions/Vault Knowledge Skill.md" "$tmp_dir/note.md"
else
    echo "  WARN: $skill_src not found, skipping Vault Knowledge Skill"
    skipped=$((skipped + 1))
fi

# ---------------------------------------------------------------------------
# 3. Behavioral notes (session operating mode, pre-compact, stop)
# ---------------------------------------------------------------------------

echo ""
echo "Seeding behavioral notes..."

# 3a. Session Operating Mode
cat > "$tmp_dir/note.md" <<'SOM_EOF'
---
type: convention
project: paivot-graph
stack: [claude-code, obsidian]
domain: developer-tools
status: active
created: TODAY_PLACEHOLDER
---

# Session Operating Mode

BEFORE STARTING: Read the vault notes listed above. Do not rediscover what is already known.
  obsidian vault="Claude" read file="<note>"

WHILE WORKING: Capture knowledge as it emerges -- do not wait for the end.
  - After making a decision (chose X over Y): create a decision note
  - After solving a non-obvious bug: create a debug note
  - After discovering a reusable pattern: create a pattern note
  Use: obsidian vault="Claude" create name="<Title>" path="_inbox/<Title>.md" content="..." silent

BEFORE ENDING: Update the project index note with what was accomplished.
  obsidian vault="Claude" append file="<Project>" content="## Session update (<date>)
  - <what was done>"

This is not optional. Knowledge that is not captured is knowledge that will be rediscovered at cost.
SOM_EOF
sed -i '' "s/TODAY_PLACEHOLDER/$TODAY/" "$tmp_dir/note.md"
create_note "Session Operating Mode" "conventions/Session Operating Mode.md" "$tmp_dir/note.md"

# 3b. Pre-Compact Checklist
cat > "$tmp_dir/note.md" <<PCL_EOF
---
type: convention
project: paivot-graph
stack: [claude-code, obsidian]
domain: developer-tools
status: active
created: $TODAY
---

# Pre-Compact Checklist

Context compaction is imminent. Save anything worth remembering NOW.

## 1. DECISIONS made this session

Record any decisions with rationale and alternatives considered:

    obsidian vault="Claude" create name="<Decision Title>" path="_inbox/<Decision Title>.md" content="---
    type: decision
    project: <project>
    status: active
    confidence: high
    created: <YYYY-MM-DD>
    ---

    # <Decision Title>

    ## Decision
    <what was decided>

    ## Rationale
    <why>

    ## Alternatives considered
    - <alt 1>
    - <alt 2>" silent

## 2. PATTERNS discovered

Record reusable solutions:

    obsidian vault="Claude" create name="<Pattern Name>" path="_inbox/<Pattern Name>.md" content="---
    type: pattern
    project: <project>
    stack: []
    status: active
    created: <YYYY-MM-DD>
    ---

    # <Pattern Name>

    ## When to use
    <context>

    ## Implementation
    <how>" silent

## 3. DEBUG INSIGHTS

Record problems solved:

    obsidian vault="Claude" create name="<Bug Title>" path="_inbox/<Bug Title>.md" content="---
    type: debug
    project: <project>
    status: active
    created: <YYYY-MM-DD>
    ---

    # <Bug Title>

    ## Symptoms
    <what happened>

    ## Root cause
    <why>

    ## Fix
    <how it was resolved>" silent

## 4. PROJECT UPDATES

Update the project index note:

    obsidian vault="Claude" append file="<Project>" content="
    ## Session update (<YYYY-MM-DD>)
    <what was accomplished>"

Do this NOW -- after compaction, the details will be lost.
PCL_EOF
create_note "Pre-Compact Checklist" "conventions/Pre-Compact Checklist.md" "$tmp_dir/note.md"

# 3c. Stop Capture Checklist
cat > "$tmp_dir/note.md" <<SCL_EOF
---
type: convention
project: paivot-graph
stack: [claude-code, obsidian]
domain: developer-tools
status: active
created: $TODAY
---

# Stop Capture Checklist

Before ending this session, confirm you have considered each of these:

- [ ] Did you capture any DECISIONS made this session? (chose X over Y, established a convention)
- [ ] Did you capture any PATTERNS discovered? (reusable solutions, idioms, workflows)
- [ ] Did you capture any DEBUG INSIGHTS? (non-obvious bugs, sharp edges, environment issues)
- [ ] Did you update the PROJECT INDEX NOTE with what was accomplished?

If none of the above apply (e.g., quick fix, trivial session), that is fine -- but confirm it was considered, not forgotten.

Use: obsidian vault="Claude" create name="<Title>" path="_inbox/<Title>.md" content="..." silent
SCL_EOF
create_note "Stop Capture Checklist" "conventions/Stop Capture Checklist.md" "$tmp_dir/note.md"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo ""
echo "Done. Created: $created, Skipped: $skipped, Failed: $failed"
if [ "$failed" -gt 0 ]; then
    echo "WARNING: Some notes failed to create. Check Obsidian is running."
    exit 1
fi

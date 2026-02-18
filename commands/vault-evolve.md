---
description: Refine vault-backed content based on session experience. Review what happened, identify improvements to agent prompts, skill content, or operating mode, and update the relevant vault notes.
allowed-tools: ["Bash", "Read", "Write", "Edit", "Glob", "Grep"]
---

# Vault Evolve -- Refine Vault Content from Experience

Review the current session's work and refine the vault notes that power paivot-graph. This closes the feedback loop: work produces experience, experience refines the vault, refined vault improves future work.

**Vault path:** `/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude`

## Step 1: Assess What Happened

Review the conversation so far. Identify:
- What tasks were completed
- What friction was encountered (agent prompts that were unclear, missing context, wrong defaults)
- What patterns emerged that should be codified
- What decisions were made that should be recorded

## Step 2: Identify Vault Notes to Update

Check which vault-backed content could be improved:

### Agent prompts (methodology/)

Use Glob to find agent notes:
```
Glob: pattern="methodology/*Agent*.md" path="/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"
```

Read any that need review:
```
Read: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/methodology/<Agent Name>.md
```

Look for:
- Instructions that were unclear or missing (agent got confused or went off-track)
- Workflow steps that should be reordered
- Quality checks that should be added
- Mode descriptions that need refinement

### Skill content (conventions/)

```
Read: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/conventions/Vault Knowledge Skill.md
```

Look for:
- Capture patterns that should be updated
- Search strategies that worked well
- Frontmatter conventions that evolved

### Behavioral notes (conventions/)

```
Read: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/conventions/Session Operating Mode.md
Read: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/conventions/Pre-Compact Checklist.md
Read: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/conventions/Stop Capture Checklist.md
```

Look for:
- Operating mode instructions that were ignored (make them more explicit)
- Checklist items that were not useful (remove or rephrase)
- Missing checklist items (add them)

## Step 3: Apply Updates

For each improvement identified, use Read to get the current content, then Edit to make targeted changes or Write to replace the full note.

When updating agent prompts, be conservative:
- Add clarifying instructions, do not remove existing ones without good reason
- Add examples of what went wrong and how to avoid it
- Preserve the overall structure

## Step 4: Report Changes

Summarize what was changed:
- Which vault notes were updated
- What the specific improvements were
- Why each change was made (what session experience motivated it)

## Constraints

- Only modify vault notes, never modify the plugin's static files (those are fallbacks)
- Keep changes grounded in actual session experience, not hypothetical improvements
- If unsure whether a change is warranted, describe it to the user and ask

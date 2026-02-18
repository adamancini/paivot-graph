---
name: vault-knowledge
description: This skill should be used when working on any project to understand how to effectively interact with the Obsidian knowledge vault. It teaches when to capture knowledge, what to capture, how to format vault notes, and how to search effectively. Use when you need to "save to vault", "update vault", "capture a decision", "record a pattern", "log a debug insight", or when starting/ending a significant work session.
version: 0.3.0
---

# Vault Knowledge (Vault-Backed)

The Obsidian vault ("Claude") lives on disk. Interact with it directly using Read, Write, Grep, and Glob tools -- this is much faster than the obsidian CLI.

**Vault path:** `/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude`

Read the full skill content from the vault:

    /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/conventions/Vault Knowledge Skill.md

The vault version is authoritative. Follow it completely.

If the vault is unavailable, use these minimal instructions:

## Fallback: Core Vault Interaction Patterns

The Obsidian vault ("Claude") is your persistent knowledge layer. It survives across sessions, projects, and context compactions.

**Vault path:** `/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude`

### Vault Structure

```
methodology/  # Agent prompts, paivot methodology
conventions/  # Working conventions (testing, python, communication)
decisions/    # Architectural and design decisions with rationale
patterns/     # Reusable solutions and idioms
debug/        # Problems and their resolutions
concepts/     # Language, framework, and tool knowledge
projects/     # One index note per project
people/       # User preferences and team conventions
_inbox/       # Unsorted capture, triage into proper folders
_templates/   # Note templates
```

### When to Capture

- **Decisions**: chose X over Y, established a convention, made a trade-off
- **Debug insights**: solved a non-obvious bug, found a sharp edge
- **Patterns**: found a reusable solution, identified an anti-pattern
- **Session boundaries**: start (read), before compaction (save), end (update)

### How to Read

Use the Read tool directly on the vault file:

    Read: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/<folder>/<Note Title>.md

### How to Search

Use the Grep tool on the vault directory:

    Grep: pattern="<term>" path="/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"

To find notes by filename:

    Glob: pattern="**/<partial-name>*.md" path="/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"

### How to Create Notes

Use the Write tool to create a new file:

    Write: /Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude/_inbox/<Title>.md

    ---
    type: decision | pattern | debug
    project: <project>
    status: active
    created: <YYYY-MM-DD>
    ---

    # <Title>

    <content>

### How to Append to Notes

Use the Edit tool to add content at the end of an existing note, or Read the note and Write it back with additions.

### How to Move/Triage Notes

Use Bash `mv` to move notes from `_inbox/` to their proper folder:

    mv "<vault-path>/_inbox/<Note>.md" "<vault-path>/decisions/<Note>.md"

### Frontmatter Requirements

Every note MUST have: type, project, status, created. Optional: stack, domain, confidence.

### The Rule

Knowledge not captured is knowledge rediscovered at cost. Capture as you go, not at the end.

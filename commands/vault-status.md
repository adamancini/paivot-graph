---
description: Show Obsidian vault health -- note counts by folder, recent notes, and overall state
allowed-tools: ["Bash", "Read", "Glob", "Grep"]
---

# Vault Status

Show the current state and health of the Obsidian knowledge vault.

**Vault path:** `/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude`

## Steps

1. **Check if vault directory exists**:
   ```bash
   test -d "$HOME/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"
   ```
   If not, report and exit.

2. **Gather vault statistics** by counting files per folder:

   Use Glob to count notes in each folder:
   ```
   Glob: pattern="methodology/*.md" path="/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"
   Glob: pattern="conventions/*.md" path="..."
   Glob: pattern="decisions/*.md" path="..."
   Glob: pattern="patterns/*.md" path="..."
   Glob: pattern="debug/*.md" path="..."
   Glob: pattern="concepts/*.md" path="..."
   Glob: pattern="projects/*.md" path="..."
   Glob: pattern="people/*.md" path="..."
   Glob: pattern="_inbox/*.md" path="..."
   ```

   List recently modified notes (Glob results are sorted by modification time).

3. **Search for potential issues**:

   Notes still in inbox (need triage):
   ```
   Glob: pattern="_inbox/*.md" path="/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude"
   ```

   Notes with missing frontmatter:
   ```
   Grep: pattern="^type:" path="/Users/ramirosalas/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude" output_mode="count"
   ```

4. **Present the report**:

   ```
   ## Vault Status

   ### Note Inventory
   | Folder        | Count | Purpose                              |
   |---------------|-------|--------------------------------------|
   | methodology/  | N     | Paivot methodology (atomic concepts) |
   | conventions/  | N     | Working conventions                  |
   | decisions/    | N     | Architectural decisions              |
   | patterns/     | N     | Reusable solutions                   |
   | debug/        | N     | Problems and resolutions             |
   | concepts/     | N     | Language/framework knowledge         |
   | projects/     | N     | Project index notes                  |
   | people/       | N     | Team preferences                     |
   | _inbox/       | N     | Unsorted (needs triage)              |
   | **Total**     | **N** |                                      |

   ### Health
   - Inbox items: N (triage needed if > 0)
   - Active projects: <list>
   - Most recent notes: <list of last 5>

   ### Recommendations
   - <any actionable suggestions based on the data>
   ```

5. **Provide actionable recommendations** based on what was found:
   - If inbox has items: "N notes in _inbox/ need triage -- move them to proper folders"
   - If a folder is empty: "No <type> notes yet -- consider capturing <type> knowledge"
   - If vault is healthy: "Vault is well-organized. Keep capturing knowledge as you work."

## If vault directory is missing

```
## Vault Status

Vault directory not found at expected path.
Expected: ~/Library/Mobile Documents/iCloud~md~obsidian/Documents/Claude

Ensure Obsidian is installed and the "Claude" vault exists.
```

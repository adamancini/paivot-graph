# Vault Seeding: How Agent Playbooks Get Deployed

## Overview

**paivot-graph** uses a vault-backed architecture where agent instructions live in the Obsidian Claude vault and are seeded from the repo.

```
Repository (seed/Sr PM Playbook.md)
    ↓ make seed / pvg seed
Obsidian Vault (.vault/knowledge/methodology/Sr PM Agent.md)
    ↓ vlt vault="Claude" read file="Sr PM Agent"
Running Agent (loads full instructions at execution time)
```

## How Seeding Works

### 1. Source Location

Agent playbooks live in `/seed/` directory in the repo:

```
paivot-graph/
├── seed/
│   ├── Sr PM Playbook.md      ← Source of truth for Sr PM Agent
│   ├── Developer Playbook.md  ← (If created)
│   ├── PM-Acceptor Playbook.md ← (If created)
│   └── ... other playbooks
├── agents/
│   ├── sr-pm.md               ← Agent definition (lightweight loader)
│   ├── developer.md           ← Agent definition
│   └── ... other agents
```

### 2. Seeding Command

```bash
# Seed vault (idempotent - only creates if missing)
make seed

# Force-update vault (overwrites existing notes)
make reseed
```

Under the hood:

```bash
CLAUDE_PLUGIN_ROOT=/path/to/paivot-graph pvg seed
```

### 3. Vault Destination

Vault notes are created in the **system vault** (`Claude`), in the `methodology/` folder:

```
~/.claude/
├── Obsidian/
│   └── Documents/Claude/
│       ├── methodology/
│       │   ├── Sr PM Agent.md      ← Seeded from seed/Sr PM Playbook.md
│       │   ├── Developer Agent.md  ← Seeded from seed/Developer Playbook.md
│       │   └── ... other methodology notes
│       ├── conventions/
│       ├── decisions/
│       └── ...
```

### 4. Agent Loading at Runtime

When an agent is spawned:

1. Agent definition (agents/sr-pm.md) runs
2. Agent tries: `vlt vault="Claude" read file="Sr PM Agent"`
3. **If successful:** loads full vault content (most common case)
4. **If vault unavailable:** falls back to inline instructions in agents/sr-pm.md

Example from sr-pm.md:

```markdown
# Senior Product Manager (Vault-Backed)

Read your full instructions from the vault (via Bash):

    vlt vault="Claude" read file="Sr PM Agent"

The vault version is authoritative. Follow it completely.

If the vault is unavailable, use these minimal instructions:

## Fallback: Core Responsibilities
[minimal fallback content]
```

## Seeding Mechanics

### How `pvg seed` Discovers Content

`pvg` looks for playbook files in `${CLAUDE_PLUGIN_ROOT}/seed/` and uploads them to the vault with specific naming:

**Naming Convention:**

```
seed/<AgentRole> Playbook.md  →  Vault note: <AgentRole> Agent
└─────────────────────────────────────────────────────────────┘
         Filename in repo              Name of vault note

seed/Sr PM Playbook.md         →  "Sr PM Agent"
seed/Developer Playbook.md     →  "Developer Agent"
seed/PM-Acceptor Playbook.md   →  "PM-Acceptor Agent"
seed/Anchor Playbook.md        →  "Anchor Agent"
```

**Vault placement:** All notes created in `Claude/methodology/` folder.

### Idempotence

`make seed` is **idempotent** — safe to run multiple times:

- If vault note doesn't exist: creates it
- If vault note exists with **same content**: no-op
- If vault note exists with **different content**: skipped (won't overwrite local edits)

To force an update:

```bash
make reseed    # Overwrites vault notes with latest repo content
```

### Version Alignment

Vault seeding is tied to **plugin version** in VERSION file:

```
paivot-graph/VERSION = 1.26.0

When seed runs:
1. Reads VERSION
2. Seeds playbooks for this version
3. Agents load from vault when spawned
```

When upgrading paivot-graph:

```bash
make bump v=1.27.0    # Updates VERSION + plugin.json + marketplace.json
make update           # Updates plugin in Claude Code
make reseed           # Reseeds vault with new playbooks
```

## Workflow for Adding New Playbooks

If you create a new agent role and want to seed its playbook:

1. **Create seed file:**
   ```
   seed/New Agent Playbook.md
   ```
   With frontmatter:
   ```yaml
   ---
   type: methodology
   project: paivot-graph
   stack: [...]
   domain: ...
   status: active
   created: 2026-03-07
   ---
   ```

2. **Create agent definition:**
   ```
   agents/new-agent.md
   ```
   With vault loader:
   ```markdown
   vlt vault="Claude" read file="New Agent"
   ```
   (Note: `new-agent.md` loads from vault as `"New Agent"`)

3. **Seed the vault:**
   ```bash
   make seed
   ```

4. **Test:**
   Spawn the agent and verify it loads from vault

## Troubleshooting

### Vault Note Not Found

If agent tries to load but vault note doesn't exist:

```
Agent tries: vlt vault="Claude" read file="Sr PM Agent"
Error: File not found
Result: Agent uses fallback instructions
```

**Solution:**
```bash
make reseed    # Force update vault
```

### Stale Vault Content

If vault note exists but is outdated:

```bash
# Update repo content
git pull

# Reseed vault with latest
make reseed
```

### Vault Unreachable

If `vlt` command fails (vault corrupted or unavailable):

Agent automatically falls back to inline instructions in agents/*.md file. System continues to work but with minimal Sr PM guidance.

### Testing Seeding Locally

```bash
# Before seeding
vlt vault="Claude" read file="Sr PM Agent" 2>&1  # Should fail

# Seed
make seed

# After seeding
vlt vault="Claude" read file="Sr PM Agent"  # Should return content
```

## Design Rationale

### Why Seed?

1. **Version control:** Playbooks are in git, versioned with releases
2. **Reproducibility:** Every plugin version has matching vault content
3. **Offline capability:** Vault is local Obsidian, works without internet
4. **Evolution:** Playbooks improve over time, changes captured in git history
5. **Distribution:** Users who install plugin automatically get current playbooks

### Why Vault Instead of Inline?

1. **Updatable:** Agents can reference vault without restarting Claude Code
2. **Living documents:** Playbooks can be refined in-session via `vlt append`
3. **Discoverable:** Users can read playbooks in Obsidian, learn the methodology
4. **Referenced:** Related notes can link to playbooks via `[[wikilinks]]`
5. **Fallback:** If vault unavailable, inline instructions provide core guidance

### Why Both?

- **Vault (primary):** Full, comprehensive, living document
- **Inline fallback:** Minimal but functional, works offline or if vault corrupted

This dual approach maximizes reliability while keeping rich guidance available.

---

## Related

- [[Sr PM Playbook]] — The actual Sr PM agent instructions (seeded content)
- [[Session Operating Mode]] — Dispatcher orchestration and vault usage
- [[Vault Knowledge Skill]] — How to interact with vault from agents

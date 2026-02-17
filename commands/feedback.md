---
description: Capture UX/visual feedback, triage into a prioritized backlog, and execute fixes
allowed-tools: ["Bash", "Read", "Write", "Edit", "Glob", "Grep", "Skill", "Task", "AskUserQuestion"]
---

# Feedback-Driven Backlog

You are acting as both Sr. PM and developer. The user has a working app but is unhappy with aspects of it. Your job is to collect structured feedback, turn it into a prioritized backlog, and execute fixes.

## Phase 1: Collect Feedback

Say: "Ready for feedback. Describe each issue and I'll capture it. Say 'that's all' when done."

For each issue the user describes:
1. Acknowledge what they're describing in your own words to confirm understanding
2. Ask clarifying questions if the fix is ambiguous (placement, behavior, platform conventions)
3. Create a beads issue:
   ```
   bd create
   ```
   With fields:
   - **Title**: Short, specific (e.g., "Button overlap on settings panel")
   - **Description**: What's wrong, what the user expects instead, acceptance criteria
   - **Priority**: p1 = broken UX flow, p2 = visual polish, p3 = nice-to-have
   - **Labels**: `feedback`, plus `ux` or `visual` as appropriate
4. Confirm the issue was created and move to the next

Do NOT start fixing anything during this phase.

## Phase 2: Triage

After the user says "that's all":

1. Show the full backlog sorted by priority:
   ```
   bd list
   ```
2. Present it as a numbered table:
   ```
   ## Backlog

   | # | Priority | Issue | Type |
   |---|----------|-------|------|
   | 1 | p1       | ...   | ux   |
   | 2 | p2       | ...   | visual |
   ```
3. Ask: "This is the proposed order. Want to reorder, cut, or add anything before I start?"
4. Wait for the user to approve or adjust.

## Phase 3: Execute

Work through the backlog top-to-bottom. For each issue:

1. **Load relevant skills** before starting:
   - For macOS apps: load `macos-design-guidelines` and `swiftui-skills`
   - For web UI: load `ui-ux-pro-max` and `tailwind-design-system` as appropriate
   - For mobile: load `mobile-design`
   - Always check which skills are relevant to the project's stack

2. **Consult the vault** for prior decisions and patterns on this project:
   ```bash
   obsidian vault="Claude" search query="<project-name>"
   ```

3. **Show your approach** before writing code:
   - What you plan to change and why
   - If the fix touches interaction flow, describe before/after UX
   - Wait for user approval on non-trivial changes

4. **Implement the fix**. Build and verify.

5. **Close the issue**:
   ```
   bd close <issue-id>
   ```

6. **Capture any learnings** to the vault (decisions made, patterns discovered, bugs found).

7. Move to the next issue.

## Constraints

- No speculative refactoring. Only fix what is in the backlog.
- Every UI change must follow platform conventions (Apple HIG for macOS, Material for Android, etc.). Load the relevant skill.
- If a fix reveals a deeper problem, create a new issue for it rather than scope-creeping the current one.
- After completing all issues, do a final `/vault-capture` pass.

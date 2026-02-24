# Skill Lifecycle Checklist

**Extracted:** 2026-02-24
**Context:** When removing, renaming, or consolidating skills in the .claude/ ecosystem

## Problem

Removing or renaming a skill directory is only one step. Skills are referenced
from multiple locations — hooks, commands, other skills, rules, and agents.
Missing any reference leaves broken cross-references that silently degrade
the system.

In practice, two categories were missed on the first pass:
1. `.claude/settings.json` hooks pointing to deleted skill scripts
2. 19 files across commands/ and skills/ still referencing old skill names

## Solution

When removing, renaming, or consolidating any skill, complete ALL steps:

### 1. Update the skill directory
- Delete or rename the directory in `.claude/skills/`
- If merging, copy unique content into the target skill

### 2. Update hooks in .claude/settings.json
- Search for the old skill path in ALL hook arrays (PreToolUse, PostToolUse,
  Stop, UserPromptSubmit, SessionStart)
- Remove or replace with the new skill's hook command
- Validate JSON after editing: `python3 -m json.tool .claude/settings.json`

### 3. Update .claude/rules/agents.md
- Update skill tables (Core Development, Code Review, Analysis, etc.)
- Update counts and loading guidance

### 4. Grep ALL cross-references
```bash
# Search for every removed/renamed skill name across the entire .claude/ tree
grep -r "old-skill-name" .claude/
```

Update matches in:
- `.claude/commands/*.md` — Required Skills sections, frontmatter skills lists
- `.claude/skills/*/SKILL.md` — Cross-references to related skills
- `.claude/agents/*.md` — Agent instructions referencing skills
- `.claude/rules/*.md` — Rule files referencing skills
- `CLAUDE.md` — Top-level project instructions

### 5. Verify zero remaining references
```bash
# Final verification — must return no matches
grep -rE "skill-a|skill-b|skill-c" .claude/ CLAUDE.md
```

### 6. Run project checks
```bash
task check  # Ensure Go codebase unaffected
```

## When to Use

- Removing a skill directory
- Renaming a skill (directory or SKILL.md name change)
- Consolidating multiple skills into one
- Migrating a skill to a new version (e.g., v1 → v2)

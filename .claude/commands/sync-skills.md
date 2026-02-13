---
description: Sync skills from .claude/skills/ to CLAUDE.md and .claude/rules/agents.md. Run after adding, modifying, or removing skills.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Write
  - Edit
argument-hint: "[--dry-run]"
model: haiku
---

# Sync Skills to Documentation

Synchronize the skills defined in `.claude/skills/` with the Skills section in CLAUDE.md and `.claude/rules/agents.md`.

## Scope: $ARGUMENTS

Parse the argument:

- `--dry-run` - Show what would change without making modifications
- (no argument) - Apply changes to CLAUDE.md and agents.md

## Phase 1: Discover Skills

### Step 1.1: Find All Skills

```bash
ls -d .claude/skills/*/
```

### Step 1.2: Extract Metadata

For each skill directory, read the `SKILL.md` file and extract:

- `name` from YAML frontmatter
- `description` from YAML frontmatter

Example SKILL.md structure:

```yaml
---
name: skill-name
description: When to use this skill
---
```

## Phase 2: Categorize Skills

Group skills by their function based on description keywords:

| Category | Keywords in Description |
| -------- | ----------------------- |
| Core Go Development | "Go code", "writing", "implementing", "Cobra", "Viper", "interfaces", "goroutines", "channels" |
| Code Review | "reviewing", "code review", "pre-commit", "BubbleTea" |
| Architecture & Planning | "architectural", "planning", "PRPs", "LEVER", "prompts" |
| A1-ASCEG System | "exploit", "vulnerability", "TABLE IX", "A1-ASCEG" |
| Advanced Features | "RAG", "vector", "multi-model", "multi-actor" |

**Categorization Rules:**

1. Match description against keywords (case-insensitive)
2. If multiple matches, use the FIRST matching category
3. If no match, place in "Other Skills" category

## Phase 3: Generate Skills Tables

### Table Format

For each category, generate a markdown table:

```markdown
### Category Name

| Skill | Triggers On |
| ----- | ----------- |
| `skill-name` | Short description from frontmatter |
```

### Trigger Description

Transform the full description into a short trigger phrase:

- Remove "Use when" prefix if present
- Truncate to first sentence or first 60 characters
- Remove trailing periods

Example:

- Full: "Apply Go idioms and Quanta project coding standards. Use when writing Go code, reviewing implementations..."
- Trigger: "Writing Go code, reviewing implementations"

## Phase 4: Update CLAUDE.md

### Step 4.1: Locate Skills Section

Find the section between:

- Start marker: `## Skills (Auto-Triggered)`
- End marker: Next `##` heading

### Step 4.2: Replace Content

Replace the content between markers with generated tables.

Preserve:

- The section header
- The intro line about `.claude/skills/`

### Step 4.3: Verify Structure

Ensure the output follows this structure:

```markdown
## Skills (Auto-Triggered)

Domain knowledge is in `.claude/skills/`. Skills auto-trigger based on context:

### Core Go Development

| Skill | Triggers On |
| ----- | ----------- |
| `go-coding-standards` | Writing Go code, reviewing implementations |
...

### Code Review

| Skill | Triggers On |
| ----- | ----------- |
...

### Architecture & Planning
...

### A1-ASCEG System
...

### Advanced Features
...
```

## Phase 5: Update agents.md

### Step 5.1: Locate Skills Section in agents.md

Find and update the skills tables in `.claude/rules/agents.md`.

### Step 5.2: Sync Core Development Skills

Update the "Core Development Skills" table to match CLAUDE.md.

### Step 5.3: Sync Code Review Skills

Update the "Code Review Skills" table to match CLAUDE.md.

### Step 5.4: Sync A1-ASCEG Skills

Update the "A1-ASCEG Skills" table to match CLAUDE.md.

## Phase 6: Report Changes

Generate a summary report:

```markdown
# Skills Sync Report

**Skills Found:** {count}
**Categories:** {count}

## Skills by Category

### Core Go Development ({count})
- skill-1
- skill-2

### Code Review ({count})
- skill-1
- skill-2

...

## Files Updated

| File | Status |
| ---- | ------ |
| CLAUDE.md | Updated/No changes |
| .claude/rules/agents.md | Updated/No changes |

## New Skills Added
- {skill-name} (category)

## Skills Removed
- {skill-name} (was in category)
```

## Dry Run Mode

If `--dry-run` is specified:

1. Perform all discovery and categorization
2. Generate the report showing what WOULD change
3. Show diff preview of CLAUDE.md changes
4. DO NOT write any files

## Error Handling

| Error | Resolution |
| ----- | ---------- |
| SKILL.md missing | Skip skill, warn in report |
| Invalid YAML frontmatter | Skip skill, warn in report |
| Missing name/description | Use directory name as fallback |
| CLAUDE.md section not found | Create section at end of file |

## Validation

After sync, verify:

- [ ] All skills from `.claude/skills/` are in CLAUDE.md
- [ ] No duplicate entries
- [ ] Tables are properly formatted
- [ ] Category assignments are correct

## Usage Examples

```bash
# Preview changes
/sync-skills --dry-run

# Apply changes
/sync-skills

# After adding a new skill
mkdir .claude/skills/new-skill
# ... create SKILL.md ...
/sync-skills
```

## Integration

This command maintains consistency between:

- `.claude/skills/*/SKILL.md` - Source of truth
- `CLAUDE.md` - User-facing documentation
- `.claude/rules/agents.md` - Agent guidance

Run this command:

- After adding new skills
- After modifying skill descriptions
- After removing skills
- As part of PR review for skill changes

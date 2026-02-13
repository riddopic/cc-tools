---
description: Audit and refactor CLAUDE.md to follow progressive disclosure, leveraging existing skills and docs infrastructure. Run periodically to prevent bloat.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Edit
  - Task
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "[audit|refactor|full]"
model: sonnet
---

# Refactor CLAUDE.md for Progressive Disclosure

Audit and refactor CLAUDE.md to maintain it as a lean navigation hub that leverages existing skills and documentation infrastructure.

## Scope: $ARGUMENTS

Parse the argument to determine operation:

- `audit` - Analyze only, report findings without changes
- `refactor` - Apply changes after audit
- `full` (default) - Full audit + refactor + verification

## Context

This is a **Go CLI project**. The existing infrastructure includes:

- **Skills** in `.claude/skills/` with YAML frontmatter and progressive disclosure
- **Detailed docs** in `docs/examples/` organized by philosophy/patterns/standards
- Skills auto-trigger based on context (testing → testing-patterns skill)

## Phase 1: Gather Current State

### Step 1.1: Measure Current Size

```bash
wc -l CLAUDE.md
```

Target: **≤150 lines**. If over, refactoring is needed.

### Step 1.2: List Available Skills

```bash
ls -la .claude/skills/*/SKILL.md
```

### Step 1.3: List Available Docs

```bash
ls -la docs/examples/**/*.md
```

## Phase 2: Audit for Issues

### Issue Type 1: Contradictions

Compare CLAUDE.md content against:

- Content in existing skills
- Content in docs/examples/
- Internal contradictions (same topic covered differently)

For each contradiction found, document:

- Location in CLAUDE.md (line numbers)
- Conflicting source
- Which version is correct

### Issue Type 2: Redundancy with Skills

Check for content that duplicates these skills:

| Skill | Topics It Covers |
| ------- | ------------------ |
| `go-coding-standards` | Go idioms, error handling, interfaces, zero values |
| `tdd-workflow` | TDD checklist, Red-Green-Refactor cycle |
| `testing-patterns` | Table-driven tests, mockery, security in tests |
| `cli-development` | Cobra/Viper patterns, flag validation |
| `coding-philosophy` | LEVER principles, Karpathy execution guidelines |
| `interface-design` | Interface segregation, composition |
| `concurrency-patterns` | Goroutines, channels, context |

**Action**: Content covered by skills should be REMOVED from CLAUDE.md.

### Issue Type 3: Agent-Known Content

Flag content that Claude already knows without explicit instruction:

- Standard Go idioms (errors are values, early returns)
- Basic naming conventions (PascalCase, camelCase)
- Standard import organization
- Basic TDD concepts
- Generic best practices ("write clean code")

**Action**: DELETE agent-known content.

### Issue Type 4: Verbose Formatting

Check for:

- Code blocks that could be tables
- Repeated command listings
- Redundant examples
- Long explanatory paragraphs

**Action**: Compact to tables or single-line entries.

## Phase 3: Define Essential Content

Content that MUST remain in CLAUDE.md:

| Section | Purpose | Target Lines |
| --------- | --------- | -------------- |
| Project Overview | Domain context | 5-10 |
| Critical Commands | Project-specific make targets | 30-40 |
| Skills Reference | Navigation table | 10-15 |
| Documentation Reference | Navigation table | 5-10 |
| Global Rules | Rules for EVERY task | 10-15 |
| Search Requirements | rg vs grep | 5-10 |
| Project-Specific Patterns | Learned gotchas | 15-25 |

## Phase 4: Refactor (if scope includes refactor)

### Step 4.1: Create Backup

```bash
cp CLAUDE.md CLAUDE.md.backup
```

### Step 4.2: Write Refactored File

Structure:

```markdown
# CLAUDE.md

## Project Overview
[1 paragraph domain context]

## Critical Commands
[Compact table format]

## Skills (Auto-Triggered)
[Table pointing to skills]

## Documentation
[Table pointing to docs]

## Global Rules
[Numbered list, 10 max]

## Search Requirements
[rg examples]

## Project-Specific Patterns
[Learned patterns and gotchas only]

## Dependencies
[Main deps only]
```

### Step 4.3: Verify Line Count

```bash
wc -l CLAUDE.md
```

Must be ≤150 lines. If over, further compaction needed.

## Phase 5: Generate Report

Output format for audit results:

```markdown
# CLAUDE.md Audit Report

**Date:** {timestamp}
**Initial Lines:** {count}
**Final Lines:** {count}
**Reduction:** {percentage}%

## Findings

### Contradictions Found
| Location | CLAUDE.md Says | Source Says | Resolution |
| ---------- | --------------- | ------------- | ------------ |

### Redundancy with Skills
| Lines | Content | Duplicated In | Action |
| ------- | --------- | --------------- | -------- |

### Agent-Known Content Removed
| Lines | Content | Reason |
| ------- | --------- | -------- |

### Verbose Content Compacted
| Before | After | Lines Saved |
| -------- | ------- | ------------- |

## Skills Audit

Verified skills are up-to-date:
- [ ] go-coding-standards
- [ ] tdd-workflow
- [ ] testing-patterns
- [ ] cli-development
- [ ] coding-philosophy
- [ ] interface-design
- [ ] concurrency-patterns

## Recommendations

1. {any skills that need updating}
2. {any docs that need updating}
3. {any new patterns to document}
```

## Anti-Patterns to Avoid

### DO NOT add to CLAUDE.md

- Standard Go conventions (use skills)
- TDD workflow details (use skills)
- Testing patterns (use skills)
- CLI patterns (use skills)
- Generic best practices
- Long code examples
- Verbose explanations

### ALWAYS add to CLAUDE.md

- Project-specific gotchas discovered during work
- Commands unique to this project
- Patterns specific to this codebase
- Global rules that apply to every task

## Maintenance Schedule

Run `/refactor-claude-md audit` when:

- CLAUDE.md exceeds 150 lines
- After adding new skills
- After significant project changes
- Monthly maintenance

Run `/refactor-claude-md refactor` when:

- Audit shows >20% redundancy
- New skills make content obsolete
- Contradictions are found

## Integration

This command works with:

- `.claude/skills/` - Authoritative domain knowledge
- `docs/examples/` - Deep dive documentation
- `LEARNINGS.md` - Recent learnings (separate from CLAUDE.md)

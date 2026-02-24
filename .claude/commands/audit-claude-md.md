---
description: Evidence-based audit of CLAUDE.md against AGENTS.md research findings. Prevents bloat, identifies harmful content, and optionally applies fixes.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Edit
  - Task
  - Bash
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "[audit|apply|full]"
model: sonnet
---

# Audit CLAUDE.md Against Research Evidence

Measure CLAUDE.md against empirical findings from "Evaluating AGENTS.md" (arXiv 2602.11988) and SkillsBench. Flag content that increases cost without improving outcomes. Optionally apply fixes.

## Evidence Base

Audit criteria derive from empirical research:

| Finding | Source | Impact |
|---------|--------|--------|
| Architecture overviews don't reduce steps to relevant files | AGENTS.md paper, 4 models | Remove overviews |
| Tool mentions amplify usage 10-50x | AGENTS.md paper, Fig 11 | Intentional placement only |
| Context files add +14-22% reasoning tokens | AGENTS.md paper, Fig 7 | Minimize always-on content |
| 2-3 skills optimal; 4+ degrades performance | SkillsBench, 7,308 trajectories | Don't list skills in CLAUDE.md |
| Context redundant with existing docs | AGENTS.md paper, Fig 5 | Remove duplicated content |
| Minimal requirements only | AGENTS.md paper, conclusion | Keep only non-discoverable content |

## Scope: $ARGUMENTS

Parse the argument to determine operation:

- `audit` — Analyze only, report findings without changes
- `apply` — Apply changes after audit
- `full` (default) — Full audit + apply + verification

## Phase 1: Gather Current State

### Step 1.1: Measure CLAUDE.md Size

```bash
wc -l CLAUDE.md
```

Target: **≤80 lines**. If over, audit will likely find removable content.

### Step 1.2: Measure Total Always-On Context

```bash
wc -l CLAUDE.md .claude/rules/*.md
```

Report total lines. Rules are path-scoped (not all load every session), but visibility helps track growth.

### Step 1.3: Estimate Token Overhead

Heuristic: 1 markdown line ≈ 12 tokens. Report "estimated per-session token overhead" for CLAUDE.md alone.

### Step 1.4: List Available Skills and Rules

```bash
ls .claude/skills/*/SKILL.md
ls .claude/rules/*.md
```

## Phase 2: Audit for Issues

Audit in order of evidence strength. For each issue found, record the line numbers, content summary, evidence source, and recommended action.

### Issue Type 1: Architecture Overviews (strongest evidence)

Flag: file trees, directory listings, package descriptions, "Architecture" sections, entry point descriptions, execution path narratives.

**Action**: REMOVE — agents navigate via Glob/Grep, not descriptions.

**Evidence**: Tested across 4 models, zero reduction in steps to find relevant files.

### Issue Type 2: Instruction Amplification (+14-22% cost)

Flag: "Always check X before Y" patterns, pre-commit checklists that duplicate rules/skills, skill-discovery instructions, search-workflow instructions.

**Action**: REMOVE if covered by a skill or rule file.

**Evidence**: Instructions cause more reasoning work without better outcomes.

### Issue Type 3: Redundancy with Rules Files (double-loading penalty)

Flag: content in CLAUDE.md that also appears in `.claude/rules/*.md`.

**Action**: REMOVE from CLAUDE.md — rules file is the authoritative source.

**Evidence**: Double-loaded content costs tokens twice with zero benefit.

Compare against all rules files:

| Rule File | Topics |
|-----------|--------|
| `coding-style.md` | Go idioms, formatting, naming |
| `testing.md` | TDD, mocks, test patterns |
| `security.md` | Security practices |
| `performance.md` | Benchmarks, profiling, model selection |
| `git-workflow.md` | Conventional commits, PR workflow |
| `comments.md` | Comment guidelines |
| `hooks.md` | Hook types, configuration |
| `agents.md` | Agent/skill orchestration |

### Issue Type 4: Contradictions

Compare CLAUDE.md content against skills, rules, and docs. For each contradiction, document which source is correct.

### Issue Type 5: Redundancy with Skills

Check for content duplicating these skills:

| Skill | Topics It Covers |
|-------|------------------|
| `go-coding-standards` | Go idioms, error handling, interfaces, zero values |
| `tdd-workflow` | TDD checklist, Red-Green-Refactor cycle |
| `testing-patterns` | Table-driven tests, mockery, security in tests |
| `go-coding-standards` (LEVER section) | LEVER decision framework |
| `code-review` | Go idioms, testing patterns, project standards |

**Action**: Content covered by skills should be REMOVED from CLAUDE.md.

### Issue Type 6: Agent-Known Content

Flag content that Claude already knows without instruction:

- Standard Go idioms (errors are values, early returns)
- Basic naming conventions
- Standard import organization
- Generic best practices ("write clean code")

**Action**: DELETE agent-known content.

### Issue Type 7: Verbose Formatting

Check for code blocks that could be tables, repeated command listings, redundant examples, long explanatory paragraphs.

**Action**: Compact to tables or single-line entries.

### Issue Type 8: Casual Tool Mentions (10-50x amplification)

Flag tool/library names outside the Build Commands section.

**Action**: Move to Build Commands section (intentional amplification) or REMOVE.

**Evidence**: Agents use mentioned tools 10-50x more than baseline.

### Issue Type 9: Cost-Inducing Patterns (summary category)

Flag: code blocks > 3 lines, sections replaceable by a single skill/rule reference.

**Action**: Estimate token cost, prioritize removal by cost-per-unique-signal ratio.

## Phase 3: Define Essential Content

CLAUDE.md should contain **only content not discoverable from code or covered by skills/rules**.

| Section | Purpose | Target Lines |
|---------|---------|--------------|
| Project Identity | What this project IS — not discoverable from code | 5-8 |
| Build Commands | Intentional tool amplification for critical commands | 10-15 |
| Minimal Requirements | Hard constraints that break builds if wrong (Go version, build tags, toolchain) | 5-8 |
| Project-Specific Gotchas | Learned patterns — the only category with positive research effect | 5-10 |

**Removed sections with rationale:**

- **Skills Reference** → already in `rules/agents.md` and `using-superpowers` skill; adding = instruction amplification
- **Documentation Reference** → navigation tables don't reduce steps (paper evidence)
- **Search Requirements** → agent-known; `search-first` skill covers this
- **Global Rules** → every current rule is already in a `.claude/rules/` file

## Phase 4: Apply Changes (if scope includes apply)

### Step 4.1: Create Backup

```bash
cp CLAUDE.md CLAUDE.md.backup
```

### Step 4.2: Write Audited File

Structure (target ≤80 lines):

```markdown
# CLAUDE.md

[1-2 sentence project description]

## Project
[Module path, language version, key toolchain constraints]

## Build Commands
[Compact table or code block — intentional tool amplification]

## Minimal Requirements
[Build tags, gotestsum requirement, other non-obvious constraints]

## Project-Specific Gotchas
[Learned patterns and gotchas only — the evidence-backed section]
```

### Step 4.3: Verify Line Count

```bash
wc -l CLAUDE.md
```

Must be ≤80 lines. If over, further compaction needed.

## Phase 5: Generate Report

Output format for audit results:

```markdown
# CLAUDE.md Audit Report

**Date:** {timestamp}
**Initial Lines:** {count}
**Final Lines:** {count} (if apply mode)
**Reduction:** {percentage}% (if apply mode)

## Cost Analysis

| File | Lines | Est. Tokens |
|------|-------|-------------|
| CLAUDE.md | {n} | {n × 12} |
| rules/coding-style.md | {n} | {n × 12} |
| rules/testing.md | {n} | {n × 12} |
| ... | ... | ... |
| **Total always-on** | {sum} | {sum × 12} |

**Per-session overhead**: {CLAUDE.md tokens} tokens before any user prompt.

## Evidence-Based Findings

| Type | Issue | Lines | Evidence Source | Action |
|------|-------|-------|----------------|--------|
| 1 - Architecture Overview | {description} | {lines} | AGENTS.md, 4 models | REMOVE |
| 2 - Instruction Amplification | {description} | {lines} | AGENTS.md, Fig 7 | REMOVE |
| 3 - Rules Redundancy | {description} | {lines} | Token cost analysis | REMOVE |
| ... | ... | ... | ... | ... |

## Detailed Findings

### Architecture Overviews Found
| Lines | Content | Action |
|-------|---------|--------|

### Instruction Amplification Found
| Lines | Content | Covered By | Action |
|-------|---------|------------|--------|

### Redundancy with Rules Files
| Lines | CLAUDE.md Content | Also In | Action |
|-------|-------------------|---------|--------|

### Contradictions Found
| Location | CLAUDE.md Says | Source Says | Resolution |
|----------|----------------|-------------|------------|

### Redundancy with Skills
| Lines | Content | Duplicated In | Action |
|-------|---------|---------------|--------|

### Agent-Known Content
| Lines | Content | Reason |
|-------|---------|--------|

### Verbose Content
| Before | After | Lines Saved |
|--------|-------|-------------|

### Casual Tool Mentions
| Lines | Tool/Library | Context | Action |
|-------|-------------|---------|--------|

## Recommendations

1. {any skills that need updating}
2. {any new gotchas to document}
3. {any rules files that need consolidation}
```

## Anti-Patterns to Avoid

### DO NOT add to CLAUDE.md

- Architecture overviews (file trees, package descriptions, directory layouts)
- "Always check X before Y" instructions (instruction amplification)
- Content already in `.claude/rules/` files (double-loading)
- Casual tool/library mentions outside Build Commands section (10-50x amplification)
- Skill discovery instructions (use `using-superpowers` skill)
- Navigation tables pointing to files (agents use Glob/Grep)
- Standard Go conventions (use skills)
- TDD workflow details (use skills)
- Testing patterns (use skills)
- Generic best practices
- Long code examples (>3 lines)
- Verbose explanations

### ALWAYS add to CLAUDE.md

- Project-specific gotchas discovered during work
- Build commands unique to this project
- Hard constraints that break builds if wrong

## Maintenance Schedule

Run `/audit-claude-md audit` when:

- CLAUDE.md exceeds 80 lines
- After adding new skills or rules files
- After significant project changes
- Monthly maintenance

Run `/audit-claude-md apply` when:

- Audit shows Architecture Overview or Instruction Amplification issues
- Any redundancy with rules files found
- New learned gotchas to add to Gotchas section

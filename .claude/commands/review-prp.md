---
description: Comprehensive read-only review of PRP completion status
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "<prp-name> [prp-name-2 ...]"
model: opus
---

# Review PRP Completion

## Arguments: $ARGUMENTS

Parse the input arguments (space-separated PRP names):

| Component     | Description                              | Example                        |
|---------------|------------------------------------------|--------------------------------|
| **PRP Names** | One or more PRP filenames without extension | `feedback-driven-exploitation` |

**PRP File Path**: `docs/PRPs/{prp-name}.md`

### Parsing Examples

| Input                                              | PRP Paths                                                                             |
|----------------------------------------------------|---------------------------------------------------------------------------------------|
| `feedback-driven-exploitation`                     | `docs/PRPs/feedback-driven-exploitation.md`                                           |
| `multi-model-ensemble rag-knowledge-system`        | `docs/PRPs/multi-model-ensemble.md`, `docs/PRPs/rag-knowledge-system.md`              |
| `unified-tui-interface structured-planning`        | `docs/PRPs/unified-tui-interface.md`, `docs/PRPs/structured-planning.md`              |

## Required Skills

This command uses the following skills (auto-loaded based on context):

| Skill | Role | When |
|-------|------|------|
| `prp-workflow` | Evaluation framework, scoring rubric (1-10), quality gates | Always |
| `review-verification-protocol` | MANDATORY before reporting ANY findings - reduces false positives | Always |
| `code-review` | Quanta project review standards, red flags, pre-commit checklist | Always |
| `go-code-review` | Go idiomatic patterns, error handling, concurrency, interfaces | Always |
| `go-testing-code-review` | Table-driven tests, assertions, coverage patterns | Always |
| `go-coding-standards` | Reference for Go idioms and project conventions | Always |
| `verification-before-completion` | Ensure verification evidence exists before claims | Always |
| `recursive-decomposition` | Large task decomposition | If PRP spans 10+ files |
| `discovery-oriented-prompts` | LLM prompt review patterns | If PRP involves LLM prompts |
| `exploit-debugging` | Exploit failure modes | If PRP involves exploit generation |

## Purpose

Read-only review that checks whether all PRP objectives were met, code is correct, and documentation is up-to-date. This command produces a structured go/no-go recommendation without modifying any files.

**This command is read-only.** It does not create evaluation reports, fix documents, or modify any code. Use `/apply-review-feedback` to act on findings.

## Execution Flow

### Step 1: Load PRP Objectives

For each PRP name in `$ARGUMENTS`:

1. Read `docs/PRPs/{prp-name}.md`
2. Extract all objectives, tasks, and acceptance criteria
3. Note any cross-PRP dependencies

If a PRP file is not found, report the error and continue with remaining PRPs.

### Step 2: Examine Code Changes

Gather the implementation delta:

```bash
# Full diff against main branch
git diff main...HEAD --stat
git diff main...HEAD

# Commit history on this branch
git log main..HEAD --oneline

# Files changed
git diff main...HEAD --name-only
```

Map changed files to PRP objectives. Identify:
- Files that correspond to specific PRP tasks
- Changed files that are NOT covered by any PRP objective (unexpected changes)
- PRP objectives with NO corresponding file changes (missing implementation)

### Step 3: Per-Objective Assessment

For each PRP objective/task:

1. **Locate implementation**: Find the code that implements this objective
2. **Read the code**: Examine the actual implementation for correctness
3. **Check tests**: Verify tests exist and cover the objective
4. **Assess quality**: Apply `go-code-review` and `go-coding-standards` checks
5. **Verdict**: Mark as Complete, Partial, or Missing

### Step 4: Run Quality Checks (Read-Only)

Run build and test commands to verify current state:

```bash
# Build validation
make build

# Linting
make lint

# Unit tests
make test

# Race detection
make test-race
```

Record pass/fail status for each check. Do NOT attempt to fix any failures.

### Step 5: Documentation Staleness Check

Check whether documentation needs updating based on the code changes:

1. **README.md**: Does it reference features/commands added or changed by this PRP?
2. **docs/USER-GUIDE/**: Are user-facing changes reflected in the guide?
3. **docs/REGRESSION_TEST.md**: Are new test scenarios documented?
4. **CLAUDE.md**: Do new patterns or learnings need recording?
5. **docs/CODING_GUIDELINES.md**: Are new coding patterns established that should be documented?

For each document, report: Up-to-date, Needs Update (with specific gaps), or N/A.

### Step 6: Apply Review Verification Protocol

**MANDATORY** before finalizing any findings:

- Re-read every finding and verify it against actual code
- Confirm line numbers and file paths are correct
- Remove any false positives
- Ensure severity ratings are justified
- Verify that "missing" items are truly missing (not implemented differently)

### Step 7: Score Using PRP Workflow Rubric

Apply the `prp-workflow` scoring rubric (1-10 scale):

### Code Quality (3 points)

- [ ] Follows Go idioms and patterns (1.5 points)
- [ ] Clean golangci-lint output (1.5 points)

### Test Coverage (2 points)

- [ ] Table-driven tests (1 point)
- [ ] Coverage >= 80% (1 point)

### Documentation (2 points)

- [ ] Complete godoc coverage (1 point)
- [ ] Clear package documentation (0.5 points)
- [ ] Examples provided (0.5 points)

### Performance (1 point)

- [ ] Benchmarks pass (0.5 points)
- [ ] No race conditions (0.5 points)

### Error Handling (1 point)

- [ ] Proper error wrapping (0.5 points)
- [ ] No ignored errors (0.5 points)

### Architecture (1 point)

- [ ] Clean package boundaries (0.5 points)
- [ ] Interface-first design (0.5 points)

## Output Format

```markdown
# PRP Review: {prp-name(s)}

## Branch: {current-branch}
## Commits: {N} commits since main
## Score: X/10

---

## Quality Checks

| Check | Status | Notes |
|-------|--------|-------|
| `make build` | PASS/FAIL | ... |
| `make lint` | PASS/FAIL | N warnings |
| `make test` | PASS/FAIL | N/M tests pass |
| `make test-race` | PASS/FAIL | ... |

---

## Per-Objective Assessment

### PRP: {prp-name}

| # | Objective | Status | Evidence | Notes |
|---|-----------|--------|----------|-------|
| 1 | {objective text} | Complete/Partial/Missing | {file:line} | ... |
| 2 | ... | ... | ... | ... |

**Completion**: X/Y objectives complete (Z%)

---

## Documentation Status

| Document | Status | Gap |
|----------|--------|-----|
| README.md | Up-to-date / Needs Update / N/A | {specific gap} |
| docs/USER-GUIDE/ | ... | ... |
| docs/REGRESSION_TEST.md | ... | ... |
| CLAUDE.md | ... | ... |

---

## Issues Found

### Critical (blocks merge)

- {issue with file:line reference}

### Major (should fix before merge)

- {issue with file:line reference}

### Minor (consider fixing)

- {issue}

---

## Unexpected Changes

Files changed that are NOT covered by any PRP objective:
- {file} — {description of change}

---

## Recommendation

**GO** / **NO-GO** / **CONDITIONAL GO**

{Justification with specific items that need attention}

### If NO-GO, required actions:
1. {action item}
2. {action item}

### If CONDITIONAL GO, recommended actions before merge:
1. {action item}
```

## Important Rules

1. **Read-only**: Do NOT modify any files, create reports, or write output files
2. **Evidence-based**: Every finding must reference a specific file and line number
3. **Verification protocol**: ALWAYS apply `review-verification-protocol` before finalizing
4. **No false positives**: Remove findings that cannot be confirmed in actual code
5. **Score honestly**: Apply the rubric objectively, do not inflate scores
6. **Check all PRPs**: If multiple PRP names given, review each one
7. **Unexpected changes**: Flag files changed that don't map to any PRP objective

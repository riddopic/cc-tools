---
description: Apply externally-provided review feedback to PRP implementation
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Edit
  - Task
  - Bash
  - TaskCreate
  - TaskUpdate
  - TaskList
  - mcp__sequential-thinking__sequentialthinking
skills:
  - tdd-workflow
  - verification-before-completion
  - code-review
  - go-code-review
  - go-testing-code-review
  - go-coding-standards
  - coding-philosophy
  - dispatching-parallel-agents
  - recursive-decomposition
  - testing-patterns
  - exploit-debugging
  - pattern-management
argument-hint: "<prp-name> [prp-name-2 ...] <paste feedback below>"
model: opus
---

# Apply Review Feedback

## Arguments: $ARGUMENTS

Parse the input:

- **PRP names**: One or more PRP filenames without extension (space-separated, before the feedback text)
- **Feedback text**: The pasted review feedback (everything after the PRP names)
- **PRP File Path**: `docs/PRPs/{prp-name}.md`

| Component     | Description                                         | Example                                    |
| --------------- | ----------------------------------------------------- | -------------------------------------------- |
| **PRP Names** | PRP filenames without extension                     | `feedback-driven-exploitation`             |
| **Feedback**  | Externally-provided review text (from codex, human) | Critical: missing error handling in foo.go |

### Parsing Examples

| Input | PRP | Feedback |
| ------- | ----- | ---------- |
| `flash-loan-rag Critical: missing validation` | `docs/PRPs/flash-loan-rag.md` | `Critical: missing validation` |
| `ensemble rag Fix the test timeout in...` | `docs/PRPs/ensemble.md`, `docs/PRPs/rag.md` | `Fix the test timeout in...` |

**Heuristic**: PRP names are short hyphenated tokens. The feedback starts at the first token that looks like a sentence or severity label (Critical, Major, Minor, Fix, Add, Remove, etc.).

## Required Skills

| Skill | Role | When |
| ------- | ------ | ------ |
| `tdd-workflow` | MANDATORY - every code fix follows Red-Green-Refactor | Always |
| `verification-before-completion` | MANDATORY - run `make lint`/`make test` before claiming done | Always |
| `code-review` | Project review standards, validate fixes meet quality bar | Always |
| `go-code-review` | Go-specific validation of fixes | Always |
| `go-testing-code-review` | Validate test fixes follow patterns | When feedback involves tests |
| `go-coding-standards` | Reference during fix implementation | Always |
| `coding-philosophy` | Decision framework (LEVER) + behavioral guardrails (Karpathy) | Always |
| `dispatching-parallel-agents` | Scope each agent to one independent failure domain | When 2+ independent items |
| `recursive-decomposition` | Handle large-scale feedback across many files | When feedback spans 10+ files |
| `testing-patterns` | Detailed test patterns and Mockery usage | When feedback involves test structure |
| `exploit-debugging` | Exploit failure modes and debugging | When feedback involves exploit failures |
| `pattern-management` | Vulnerability pattern updates | When feedback involves pattern detection |

## CRITICAL: This Command Always Starts in Plan Mode

**Before touching any code**, this command MUST:

1. Parse and categorize all feedback items
2. Assess each item's validity
3. Run the Integration & Wiring Checklist to detect gaps
4. Flag any concerns or disagreements
5. Present a prioritized plan for user approval
6. Wait for explicit approval before executing fixes

This ensures the user can validate, modify, or reject items before any code changes occur.

---

## Phase 1: Plan (Before Any Code Changes)

### Step 1: Parse Feedback

Read PRP(s) for context, then parse the pasted feedback into discrete items:

For each feedback item, extract:

- **Category**: Code fix, test fix, doc update, quality fix, architecture change
- **Severity**: Critical, Major, Minor (infer from language if not explicit)
- **Target**: File(s) and line(s) affected
- **Description**: What needs to change
- **Rationale**: Why the reviewer flagged this (if stated)

### Step 2: Assess Validity

For each feedback item, read the actual code and classify:

| Assessment | Meaning | Action |
| ------------ | --------- | -------- |
| **Valid** | Reviewer is correct, code needs fixing | Include in plan |
| **Questionable** | Reviewer may be wrong or feedback is ambiguous | Flag for user, include tentatively |
| **Incorrect** | Reviewer is factually wrong (e.g., code already handles this) | Flag for user with evidence, exclude |

**IMPORTANT**: Provide evidence for questionable and incorrect assessments. Reference specific file:line where the code already handles the concern or where the reviewer's assumption is wrong.

### Step 3: Integration & Wiring Checklist

Beyond the explicit feedback items, proactively check for integration gaps:

#### Documentation Staleness

Check whether the PRP's changes require updates to:

| Document | Check | Action if Stale |
| ---------- | ------- | ----------------- |
| `README.md` | Does it reference new commands, flags, or features added by this PRP? | Add as doc update item |
| `docs/USER-GUIDE/` | Are user-facing workflows or instructions affected? | Add as doc update item |
| `docs/REGRESSION_TEST.md` | Are new test scenarios or regression checks needed? | Add as doc update item |

#### Command Wiring

Check whether new features need to be wired into existing CLI commands:

| Command | Check |
| --------- | ------- |
| `analyze` | Does this PRP add new analysis capabilities (vulnerability patterns, static checks) that should be callable from `analyze`? |
| `run` | Does this PRP add exploit strategies, execution modes, or runtime features that `run` should invoke? |
| `regression` | Does this PRP add contracts, patterns, or test cases that should be included in regression testing? |
| Other commands | Grep `cmd/*.go` for commands that reference packages modified by this PRP — do they need updates? |

#### Surfacing Gaps

For each integration gap found:

- Add it as a new feedback item with category "integration" and severity "Major"
- Include it in the plan alongside reviewer feedback items
- Mark its source as "auto-detected" (vs "reviewer") so the user can distinguish

### Step 4: Identify Work Streams

Using `dispatching-parallel-agents` patterns:

- Group feedback items by independence (can they be fixed in parallel?)
- Identify sequential dependencies (item B requires item A to be fixed first)
- Map items to specialized agents:

| Feedback Type | Agent |
| --------------- | ------- |
| Core Go code fixes | `backend-systems-engineer` |
| CLI command fixes | `cli-design-architect` |
| Test fixes/additions | `qa-test-engineer` |
| Security issues | `security-threat-analyst` |
| Performance issues | `performance-optimizer` |
| Documentation updates | `technical-docs-writer` |

### Step 5: Present Plan

Output a structured plan for user approval:

```markdown
## Review Feedback Plan

### PRP(s): {names}
### Feedback Items: {N} total ({X} valid, {Y} questionable, {Z} incorrect)

---

### Flagged Concerns

> **Item {N}** (Questionable): "{reviewer's feedback}"
> **Assessment**: {why this is questionable}
> **Evidence**: {file:line showing current code}
> **Recommendation**: Include / Skip / Modify

> **Item {M}** (Incorrect): "{reviewer's feedback}"
> **Assessment**: {why this is incorrect}
> **Evidence**: {file:line showing code already handles this}
> **Recommendation**: Skip

---

### Approved Items (Pending Your Confirmation)

#### Critical
1. {item description} — {target file(s)} — Agent: {agent-type}
2. ...

#### Major
3. {item description} — {target file(s)} — Agent: {agent-type}
4. ...

#### Minor
5. {item description} — {target file(s)} — Agent: {agent-type}

---

### Auto-Detected Integration Items

Items discovered by the Integration & Wiring Checklist (not from reviewer):

| # | Category | Description | Target | Severity |
| --- | ---------- | ------------- | -------- | ---------- |
| A1 | integration | {description} | {file(s)} | Major |
| A2 | integration | {description} | {file(s)} | Major |

---

### Work Streams

**Parallel Stream A**: Items 1, 3 (independent, can run concurrently)
**Parallel Stream B**: Items 2, 5 (independent, can run concurrently)
**Sequential**: Item 4 depends on Item 1

---

### Awaiting Approval

Please review the plan above and:
- Approve all items
- Remove specific items by number
- Modify specific items
- Add additional items
```

**STOP HERE and wait for user approval before proceeding to Phase 2.**

---

## Phase 2: Execute (After User Approval)

### Step 1: Apply Karpathy Guidelines

Before writing any code:

- Think through the fix before implementing
- Make minimal, targeted changes
- Do not refactor surrounding code
- Verify assumptions by reading code first

### Step 2: Execute Fixes with TDD

For each approved item, following `tdd-workflow`:

1. **RED**: Write or update a failing test that captures the expected behavior
2. **GREEN**: Implement the minimum fix to make the test pass
3. **REFACTOR**: Clean up only if necessary, keeping tests green

Use TaskCreate to track progress per feedback item:

```
TaskCreate("Item 1: {description}")  # Creates trackable task
TaskUpdate(taskId, status: "in_progress")  # Mark when starting RED phase
# ... RED → GREEN → VERIFY ...
TaskUpdate(taskId, status: "completed")  # Mark when verified
TaskList()  # Show remaining items
```

### Step 3: Dispatch Parallel Agents

For independent work streams identified in the plan, dispatch agents in parallel using `dispatching-parallel-agents` patterns:

- Each agent gets ONE independent failure domain
- Each agent receives full context: PRP objectives, feedback item, target files, coding standards
- Each agent follows TDD workflow independently

### Step 4: Validate Fixes

Apply `verification-before-completion` after all fixes:

```bash
# Format
make fmt

# Lint
make lint

# Unit tests
make test

# Race detection
make test-race

# Build
make build
```

ALL checks must pass before claiming completion.

### Step 5: Output Report

```markdown
## Feedback Applied

### PRP(s): {names}
### Items Fixed: {N}/{M} approved items

---

### Changes Made

| # | Source | Item | File(s) Changed | Test Added/Modified | Status |
| --- | -------- | ------ | ----------------- | --------------------- | -------- |
| 1 | reviewer | {description} | {files} | {test file} | Done |
| 2 | reviewer | {description} | {files} | {test file} | Done |
| 3 | reviewer | {description} | — | — | Skipped (reason) |
| A1 | auto-detected | {description} | {files} | {test file} | Done |
| A2 | auto-detected | {description} | — | — | Skipped (reason) |

---

### Verification Results

| Check | Status |
| ------- | -------- |
| `make fmt` | PASS/FAIL |
| `make lint` | PASS/FAIL |
| `make test` | PASS/FAIL (N/M tests) |
| `make test-race` | PASS/FAIL |
| `make build` | PASS/FAIL |

---

### Items Not Addressed

- Item {N}: {reason — user rejected, blocked by dependency, etc.}

### Follow-up Needed

- {any remaining work or concerns}
```

## Important Rules

1. **Plan first, always**: Never modify code before user approves the plan
2. **TDD discipline**: Every fix starts with a failing test
3. **Minimal changes**: Fix what was asked, nothing more
4. **Flag concerns**: If the reviewer is wrong, say so with evidence
5. **Verify before claiming done**: All make targets must pass
6. **Parallel when possible**: Use `dispatching-parallel-agents` for independent items
7. **Evidence-based**: Reference file:line for all assessments and changes

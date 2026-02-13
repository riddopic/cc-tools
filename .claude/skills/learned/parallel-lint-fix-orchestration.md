# Parallel Lint Fix Orchestration

**Extracted:** 2026-02-09
**Context:** Fixing multiple independent lint violations and test failures across a Go codebase

## Problem

Running `make lint` + `make test` reveals 20+ violations across 6+ files in different packages. Fixing sequentially takes 30+ minutes as each fix requires reading, editing, building, and testing. Many violations are independent (different files, different root causes) and can be fixed concurrently.

## Solution

### Step 1: Classify Violations by Independence

Group lint/test failures into independent clusters:

- **Independent:** Different files, different root causes, no shared code paths
- **Dependent:** Same file, or fix in one file affects another

### Step 2: Create One Agent Per Cluster

For each independent cluster, dispatch a background agent with:

- Exact file path(s) and line numbers
- The specific lint rule or test error
- Clear fix instructions (what to remove/rename/refactor)
- A verification command (`go build ./... && go test -short -count=1 ./pkg/...`)

```
Agent 1: Remove unused consts from codex_client.go
Agent 2: Remove unused consts from gemini_client.go
Agent 3: Refactor worker.go complexity (gocognit)
Agent 4: Fix test mock missing chain ID
Agent 5: Fix unparam in test helper
```

### Step 3: Launch All in Parallel

Use `run_in_background: true` on all Task calls in a single message. This runs them concurrently.

### Step 4: Handle Stragglers

While agents run:

- Fix trivial issues directly (single-line removals)
- Watch for NEW diagnostics revealed by agent changes
- Stale diagnostics may show pre-fix state — verify with grep before acting

### Step 5: Final Verification

After all agents complete:

```bash
make lint   # Expect 0 issues
make test   # Expect 0 failures
```

## Key Insights

- **gocognit refactoring:** Extract logical blocks (profit parsing, error handling) into pure helper functions with clear input/output signatures
- **Mock coverage gaps:** When production code adds a new enum value (e.g., new chain ID), ALL mock helpers that iterate known values must be updated
- **Stale diagnostics:** IDE diagnostics refresh on file save, not on agent completion — always re-run `make lint` rather than trusting in-flight diagnostics
- **Agent scope:** Give each agent exactly ONE domain. Agents that touch multiple unrelated files tend to have more conflicts.

## When to Use

- `make lint` shows 5+ violations across 3+ independent files
- `make test` shows failures in multiple test packages
- Violations are in different packages with no shared dependencies
- Time pressure requires parallel fixing rather than sequential

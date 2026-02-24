---
name: prp-workflow
description: Apply PRP (Product Requirements Plan) workflow patterns. Use when working with PRPs, generating PRPs, executing PRPs, evaluating implementations, or fixing PRP issues. Auto-triggers for /generate-initial, /generate-prp, /execute-prp, /evaluate-prp, /fix-prp commands.
context: fork
---

# PRP Workflow Patterns

The PRP (Product Requirements Plan) workflow is a multi-stage development process for implementing Go features with high first-pass success rates.

## Workflow Overview

```
Sprint Doc → INITIAL.md → PRP → Execute → Evaluate → Fix (if needed)
```

### Stage 1: Generate INITIAL.md (`/generate-initial`)
Transform sprint documentation into a structured INITIAL.md with:
- Feature description
- Go code examples (interfaces, implementations, tests)
- Documentation references
- Implementation checklist

### Stage 2: Generate PRP (`/generate-prp`)
Deep research phase that produces a comprehensive PRP document:
- Codebase analysis for existing patterns
- External research (Go docs, libraries, best practices)
- Architecture review
- Validation gates (bash commands for quality checks)

**Key Principle**: The more research, the better the one-pass implementation success rate.

### Stage 2.5: Refinement Interview
Before writing the PRP, conduct a structured AskUserQuestion phase to surface:
- Scope boundaries (in vs out of scope)
- Design decisions when multiple approaches exist
- Integration points with existing commands/packages
- Priority trade-offs (MVP vs nice-to-have)

This reduces rework during execution by capturing intent accurately upfront.

### Stage 3: Execute PRP (`/execute-prp`)
Implementation phase using sub-agent orchestration:

**Claude Code Role**: Orchestrator ONLY - never writes code directly.

**Task Subsystem** (TaskCreate/TaskUpdate/TaskList):
1. Create all PRP tasks via TaskCreate with `blockedBy` dependencies at start
2. Mark each task `in_progress` before starting, `completed` after quality gates pass
3. Run TaskList after each completion to show progress and unblocked tasks

**Atomic Commits Per Task**:
After each task passes quality gates, commit only that task's files with the PRP's suggested commit message. This creates a clean git history where each commit maps to one PRP task.

**Sub-Agent Delegation**:
| Task Type | Agent |
|-----------|-------|
| CLI (Cobra/Viper) | cli-tool-developer |
| Core business logic | backend-systems-engineer |
| Concurrency | concurrency-specialist |
| HTTP/gRPC services | api-backend-engineer |
| Performance-critical | performance-optimizer |

**Recovery/Continuation**:
If a session fails mid-execution, start a new session with the PRP pinned. Check git log for completed task commits, then resume from the next incomplete task. The PRP is the stable recovery point.

### Stage 4: Evaluate PRP (`/evaluate-prp`)
Quality assessment against Go standards:

**Scoring Rubric (1-10)**:
- Code Quality: 3 points (Go idioms + clean lint)
- Test Coverage: 2 points (table-driven + ≥80%)
- Documentation: 2 points (godoc + examples)
- Performance: 1 point (benchmarks + no races)
- Error Handling: 1 point (wrapping + no ignores)
- Architecture: 1 point (clean packages + interfaces)

**Pass Threshold**: Score ≥ 8/10

### Stage 5: Fix PRP (`/fix-prp`)
Systematic remediation if score < 8:
1. Load evaluation report and fixes document
2. Categorize issues (Critical/Major/Minor)
3. Fix in TDD order (test first, then implementation)
4. Validate after each category
5. Auto-chain to `/evaluate-prp` for re-evaluation

## Quality Gates

Standard validation commands:
```bash
task fmt          # Format check
task lint         # golangci-lint
task test-race    # Tests with race detector
task coverage     # Test coverage report
task build        # Build validation
```

## Go Standards Enforced

All PRP implementations must follow:
- Interface-first design (small, focused interfaces)
- Table-driven tests with t.Run subtests
- Error wrapping with context (`fmt.Errorf("...: %w", err)`)
- Context as first parameter
- No ignored errors (`_ = someFunc()`)
- No panic for normal error handling
- Proper godoc on all exported items

## Related Skills

The PRP workflow leverages these skills:
- `go-coding-standards` - Go idioms and patterns
- `tdd-workflow` - Test-first development
- `testing-patterns` - Table-driven tests, mocking
- `code-review` - Quality review standards

## Common Issues

### Low Evaluation Scores
- Missing godoc comments on exports
- Ignored errors (`_ = err`)
- Interfaces too large (>5 methods)
- Magic numbers without constants
- Missing table-driven tests

### Execution Failures
- Claude Code writing code directly (must delegate)
- Missing validation gates in PRP
- Insufficient research context

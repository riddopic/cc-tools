# Agent Orchestration

Guidelines for using specialized agents and skills.

## Available Skills (Auto-Triggered)

Skills in `.claude/skills/` auto-trigger based on context. Organized by category:

### Core Development (7)

| Skill | Triggers On |
| ----- | ----------- |
| `go-coding-standards` | Writing Go code, reviewing implementations |
| `tdd-workflow` | Implementing features (TDD is mandatory) |
| `test-driven-development` | Before writing implementation code |
| `testing-patterns` | Writing tests, Mockery v3.5 patterns |
| `interface-design` | Defining interfaces, composition |
| `cli-development` | Cobra/Viper CLI commands |
| `concurrency-patterns` | Goroutines, channels, context |

### Code Review (4)

| Skill | Triggers On |
| ----- | ----------- |
| `code-review` | Reviewing code, pre-commit checks |
| `go-code-review` | Reviewing .go files for idiomatic patterns |
| `go-testing-code-review` | Reviewing *_test.go files |
| `review-verification-protocol` | Before reporting ANY code review findings |

### Documentation (6)

| Skill | Triggers On |
| ----- | ----------- |
| `docs-style` | Voice, tone, structure for technical docs |
| `tutorial-docs` | Learning-oriented tutorial guides |
| `howto-docs` | Task-oriented how-to guides |
| `explanation-docs` | Understanding-oriented conceptual guides |
| `reference-docs` | API and symbol reference documentation |
| `writing-clearly-and-concisely` | Any prose humans will read |

### Workflow (13)

| Skill | Triggers On |
| ----- | ----------- |
| `brainstorming` | Before any creative work or feature design |
| `writing-plans` | Multi-step task requiring a plan |
| `executing-plans` | Executing a written implementation plan |
| `subagent-driven-development` | Executing plans with independent tasks |
| `dispatching-parallel-agents` | 2+ independent tasks without shared state |
| `finishing-a-development-branch` | Implementation complete, deciding integration path |
| `receiving-code-review` | Receiving and implementing review feedback |
| `requesting-code-review` | Completing tasks, verifying quality |
| `verification-before-completion` | Before claiming work is complete |
| `using-git-worktrees` | Feature work needing workspace isolation |
| `using-superpowers` | Start of conversation, finding skills |
| `writing-skills` | Creating or editing skill files |
| `prp-workflow` | Working with PRP workflow patterns |

### Analysis & Patterns (5)

| Skill | Triggers On |
| ----- | ----------- |
| `systematic-debugging` | Bugs, test failures, root cause investigation |
| `audit-context-building` | Line-by-line code analysis for deep context |
| `continuous-learning` | Extracting reusable patterns from sessions |
| `recursive-decomposition` | Breaking complex problems into sub-problems |
| `coding-philosophy` | LEVER decisions + Karpathy execution guidelines |

### Specialized (1)

| Skill | Triggers On |
| ----- | ----------- |
| `commit` | Staging changes, creating atomic git commits |

## When to Use Agents Immediately

No user prompt needed — invoke proactively:

1. **Complex feature requests** → Use `Plan` subagent type (built-in)
2. **Code just written/modified** → Use `code-review-specialist` agent
3. **Bug fix or new feature** → Use `test-strategy-designer` agent
4. **Architectural decision** → Use `systems-architect` agent
5. **Security-sensitive code** → Use `security-threat-analyst` agent
6. **Build failures** → Use `code-analyzer-debugger` agent

## Parallel Task Execution

Always use parallel execution for independent operations. See the `dispatching-parallel-agents` skill for detailed patterns on when and how to dispatch.

```markdown
# ✅ DO: Launch in parallel
Launch 3 agents in parallel:
1. Agent 1: Security analysis of auth.go
2. Agent 2: Performance review of cache.go
3. Agent 3: Code review of handler.go

# ❌ DON'T: Run sequentially when unnecessary
First agent 1, then agent 2, then agent 3
```

## Multi-Perspective Analysis

For complex problems, use multiple perspectives:

- Factual reviewer
- Senior engineer
- Security expert
- Performance specialist
- Consistency checker

## Agent Selection Guidelines

| Task Type | Recommended |
| --------- | ----------- |
| New Go feature | `test-strategy-designer` → `code-review-specialist` |
| CLI command | `cli-development` skill → `cli-design-architect` agent |
| Performance issue | `performance-optimizer` agent |
| Security review | `security-threat-analyst` agent |
| Test failures | `systematic-debugging` skill → `code-analyzer-debugger` agent |
| Architecture decisions | `systems-architect` agent |
| Documentation | `technical-docs-writer` agent |
| Debugging bugs | `systematic-debugging` skill → `code-analyzer-debugger` agent |
| Large-scale analysis | `recursive-decomposition` skill |
| Code refactoring | `code-refactoring-expert` agent |

## Quick Reference

```bash
# Run pre-commit checks (invokes code review)
task pre-commit

# Run all quality checks
task check

# Quick commands
task q   # Quick build
task qt  # Quick test
task ql  # Quick lint
```

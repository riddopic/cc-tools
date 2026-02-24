# Agent Orchestration

Guidelines for using specialized agents and skills.

## Available Skills (Auto-Triggered)

Skills in `.claude/skills/` auto-trigger based on context. Load **2-3 most relevant** per task. Organized by category:

### Core Development (3)

| Skill | Triggers On |
| ----- | ----------- |
| `go-coding-standards` | Writing Go code, reviewing implementations, LEVER decisions |
| `tdd-workflow` | Implementing features, fixing bugs (TDD is mandatory) |
| `testing-patterns` | Writing tests, Mockery v3.5 patterns |

### Code Review (1)

| Skill | Triggers On |
| ----- | ----------- |
| `code-review` | Reviewing code, pre-commit checks, verification protocol |

### Documentation (6)

| Skill | Triggers On |
| ----- | ----------- |
| `docs-style` | Voice, tone, structure for technical docs |
| `tutorial-docs` | Learning-oriented tutorial guides |
| `howto-docs` | Task-oriented how-to guides |
| `explanation-docs` | Understanding-oriented conceptual guides |
| `reference-docs` | API and symbol reference documentation |
| `writing-clearly-and-concisely` | Any prose humans will read |

### Workflow (14)

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
| `search-first` | Before writing new code, adding dependencies |

### Analysis & Patterns (6)

| Skill | Triggers On |
| ----- | ----------- |
| `systematic-debugging` | Bugs, test failures, root cause investigation |
| `audit-context-building` | Line-by-line code analysis for deep context |
| `continuous-learning-v2` | Instinct management via `cc-tools instinct` commands |
| `recursive-decomposition` | Breaking complex problems into sub-problems |
| `reviewing-with-codex` | Second opinion on plans from Codex |
| `skill-audit` | Auditing skill quality, co-loading budget, periodic maintenance |

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
7. **Security audit / deep code analysis** → Use `function-analyzer` agent (with `audit-context-building` skill)

## Parallel Task Execution

Always use parallel execution for independent operations. See the `dispatching-parallel-agents` skill for detailed patterns on when and how to dispatch.

## Agent Selection Guidelines

| Task Type | Recommended |
| --------- | ----------- |
| New Go feature | `test-strategy-designer` → `code-review-specialist` |
| CLI command | `cli-design-architect` agent |
| Performance issue | `performance-optimizer` agent |
| Security review | `security-threat-analyst` agent |
| Test failures | `systematic-debugging` skill → `code-analyzer-debugger` agent |
| Architecture decisions | `systems-architect` agent |
| Documentation | `technical-docs-writer` agent |
| Debugging bugs | `systematic-debugging` skill → `code-analyzer-debugger` agent |
| Large-scale analysis | `recursive-decomposition` skill |
| Code refactoring | `code-refactoring-expert` agent |
| Security audit preparation | `audit-context-building` skill → `function-analyzer` agent |
| Skill quality audit | `skill-audit` skill |

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

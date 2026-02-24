# Skills and Slash Commands Reference

cc-tools includes two extension mechanisms for Claude Code: **skills** (auto-triggered context files) and **slash commands** (user-invoked actions). Skills load automatically based on task context. Slash commands are invoked explicitly with `/<command-name>`.

## Skills

Skills live in `.claude/skills/<name>/SKILL.md`. Claude Code loads the 2-3 most relevant skills per task based on context matching. Skills provide procedural guidance, not general knowledge.

### Core Development

| Skill | Triggers On |
|-------|-------------|
| `go-coding-standards` | Writing Go code, reviewing implementations, making architecture decisions |
| `tdd-workflow` | Implementing features, fixing bugs (TDD is mandatory in this project) |
| `testing-patterns` | Writing tests, setting up mocks, using Mockery v3.5 patterns |

### Code Review

| Skill | Triggers On |
|-------|-------------|
| `code-review` | Reviewing code, pre-commit checks, pull request verification |

### Documentation

| Skill | Triggers On |
|-------|-------------|
| `docs-style` | Voice, tone, and structure for all technical documentation |
| `tutorial-docs` | Writing learning-oriented tutorial guides |
| `howto-docs` | Writing task-oriented how-to guides |
| `explanation-docs` | Writing understanding-oriented conceptual guides |
| `reference-docs` | Writing API and symbol reference documentation |
| `writing-clearly-and-concisely` | Any prose that humans read |

### Workflow

| Skill | Triggers On |
|-------|-------------|
| `brainstorming` | Before creative work or feature design |
| `writing-plans` | Multi-step tasks requiring a plan |
| `executing-plans` | Executing a written implementation plan |
| `subagent-driven-development` | Executing plans with independent tasks |
| `dispatching-parallel-agents` | 2+ independent tasks without shared state |
| `finishing-a-development-branch` | Implementation complete, deciding integration path |
| `receiving-code-review` | Receiving and implementing review feedback |
| `requesting-code-review` | Completing tasks, verifying quality |
| `verification-before-completion` | Before claiming work is complete |
| `using-git-worktrees` | Feature work needing workspace isolation |
| `using-superpowers` | Start of conversation, finding relevant skills |
| `writing-skills` | Creating or editing skill files |
| `prp-workflow` | Working with Product Requirements Plan patterns |
| `search-first` | Before writing new code or adding dependencies |

### Analysis and Patterns

| Skill | Triggers On |
|-------|-------------|
| `systematic-debugging` | Bugs, test failures, root cause investigation |
| `audit-context-building` | Line-by-line code analysis for deep context |
| `continuous-learning-v2` | Instinct management via `cc-tools instinct` commands |
| `recursive-decomposition` | Breaking complex problems into sub-problems |
| `reviewing-with-codex` | Getting second opinions from Codex on plans |
| `skill-audit` | Auditing skill quality and periodic maintenance |

### Specialized

| Skill | Triggers On |
|-------|-------------|
| `commit` | Staging changes, creating atomic git commits |

## Slash Commands

Slash commands live in `.claude/commands/<name>.md`. Invoke them in Claude Code with `/<command-name>`.

### Development

| Command | Description |
|---------|-------------|
| `/tdd` | Enforce test-driven development workflow |
| `/fix-tests` | Fix failing tests using TDD workflow |
| `/fix-linting` | Fix linting errors using Go coding standards |
| `/test-coverage` | Analyze test coverage and generate missing tests |
| `/refactor-clean` | Identify and remove dead code with test verification |
| `/verify` | Run verification checks before completion |

### Documentation

| Command | Description |
|---------|-------------|
| `/draft-docs` | Generate first-draft documentation from code analysis |
| `/improve-doc` | Improve existing documentation using Diataxis principles |
| `/update-docs` | Sync documentation from source-of-truth files |
| `/update-codemaps` | Analyze codebase structure and update architecture docs |

### Planning and Execution

| Command | Description |
|---------|-------------|
| `/write-plan` | Create a multi-step implementation plan |
| `/execute-plan` | Execute a written implementation plan |
| `/brainstorm` | Explore ideas before implementation |
| `/ultrathink` | Deep thinking mode for complex tasks |
| `/ultrathink-with-research` | Deep thinking with web and documentation research |

### Code Review

| Command | Description |
|---------|-------------|
| `/review-staged-unstaged` | Review all staged and unstaged Go code changes |
| `/audit-context` | Build deep architectural context for analysis |
| `/audit-claude-md` | Audit CLAUDE.md against research findings |

### Git and Release

| Command | Description |
|---------|-------------|
| `/git-commit` | Stage changes and create conventional commits |
| `/release` | Prepare a release with checks, docs, and tagging |

### Instinct Management

| Command | Description |
|---------|-------------|
| `/instinct-status` | Show all learned instincts with confidence levels |
| `/instinct-export` | Export instincts for sharing |
| `/instinct-import` | Import instincts from external sources |
| `/evolve` | Cluster related instincts into skills, commands, or agents |
| `/learn` | Extract reusable patterns from current session |
| `/learn-eval` | Extract patterns with self-evaluation |

### Session Management

| Command | Description |
|---------|-------------|
| `/sessions` | Manage Claude Code sessions |

### Skills and Sync

| Command | Description |
|---------|-------------|
| `/sync-skills` | Sync skills from `.claude/skills/` to CLAUDE.md and agents.md |

### External Tools

| Command | Description |
|---------|-------------|
| `/consult` | Get second opinions from AI team (Gemini, Codex, Claude, Kimi) |

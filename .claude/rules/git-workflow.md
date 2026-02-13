# Git Workflow

Git conventions for the Quanta project.

## Commit Message Format

```text
<type>: <description>

<optional body>
```

**Types:** `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`

**Examples:**

```text
feat: add theme switcher command
fix: resolve race condition in metrics collector
refactor: extract validation logic to separate package
test: add table-driven tests for config loader
```

## Pre-Commit Requirements

Always run before committing. Apply `verification-before-completion` before every commit.

```bash
task pre-commit
```

This runs:

- `task fmt` - Format code (gofmt + goimports)
- `task lint` - Run golangci-lint
- `task test` - Run tests

## Pull Request Workflow

When creating PRs:

1. **Analyze full commit history** (not just latest commit)

   ```bash
   git diff main...HEAD
   git log main..HEAD --oneline
   ```

2. **Run all checks**

   ```bash
   task check
   task test-race
   ```

3. **Create PR with comprehensive summary**

   ```bash
   gh pr create --title "feat: add theme switcher" --body "$(cat <<'EOF'
   ## Summary
   - Add `theme switch` command for runtime theme changes
   - Support for all built-in themes
   - Persist selection to config file

   ## Test plan
   - [ ] Run `quanta theme switch powerline`
   - [ ] Verify theme persists after restart
   - [ ] Test invalid theme handling
   EOF
   )"
   ```

## Feature Implementation Flow

1. **Plan First**
   - Use planner agent for complex features
   - Identify dependencies and risks
   - Break into small, testable chunks

2. **TDD Approach**
   - Write test (RED)
   - Implement minimum code (GREEN)
   - Refactor if needed (REFACTOR)
   - Verify 80%+ coverage

3. **Code Review**
   - Use code-reviewer agent after writing code
   - Address CRITICAL and HIGH issues
   - Run `task pre-commit`

4. **Commit**
   - Detailed commit message
   - Follow conventional commits format
   - One logical change per commit

## Branch Naming

```text
feat/add-theme-switcher
fix/race-condition-metrics
refactor/extract-validation
test/config-loader-coverage
```

## Quick Commands

```bash
# Check status
git status

# Stage specific files (prefer over git add -A)
git add cmd/theme.go internal/theme/switcher.go

# Commit with conventional message
git commit -m "feat: add theme switcher command"

# Push and set upstream
git push -u origin feat/add-theme-switcher

# Create PR
gh pr create --fill
```

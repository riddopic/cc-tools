---
name: git-commit
description: Stage changes and create atomic git commits with conventional commit format. Use when the user asks to commit, after completing implementation work, or when invoking /git-commit.
---

# Git Commit

## Process

1. **Inspect all changes** — run `git status`, `git diff`, and `git diff --cached` to see staged, unstaged, and untracked files. Also run `git log --oneline -5` to match the repo's commit message style.

2. **Run pre-commit checks** — run `task pre-commit` (fmt + lint + test). If anything fails, fix it before committing. Do not skip this step.

3. **Group into atomic commits** — each commit should represent one logical change that is independently reversible. Separate unrelated changes into distinct commits. When in doubt, fewer larger commits are better than many micro-commits.

   Grouping criteria:
   - Same feature or bugfix = one commit
   - Refactoring separate from behavior changes = separate commits
   - Test additions alongside the code they test = one commit
   - Config/tooling changes unrelated to feature work = separate commit

4. **Write commit messages** — use conventional commit format. Summarize the "why" not the "what." Use imperative mood.

   ```text
   <type>: <description>
   ```

   Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`

5. **Stage specific files** — prefer `git add <file> ...` over `git add -A` or `git add .` to avoid accidentally including secrets or unrelated files. Never commit `.env`, credentials, or large binaries.

6. **Create commits** — use a HEREDOC for the message to ensure correct formatting:

   ```bash
   git commit -m "$(cat <<'EOF'
   feat: add theme switcher command
   EOF
   )"
   ```

7. **Verify** — run `git status` after each commit to confirm clean state.

## Rules

- Execute commits directly without asking "should I commit now?" — but always show what you're committing and why.
- Never amend a previous commit unless explicitly asked. Always create new commits.
- Never run `git push` unless the user explicitly asks.
- If pre-commit checks fail, fix the issues and re-run before committing.
- Do not commit files that likely contain secrets. Warn if the user asks to commit them.

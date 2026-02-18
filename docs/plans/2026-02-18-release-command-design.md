# Release Command Design

## Problem

The existing `.claude/commands/release.md` was written for a JavaScript/TypeScript project. It references `package.json`, `src/index.ts`, `bun test`, and `expect()` counts — none of which exist in cc-tools. The command needs a complete rewrite for this Go CLI project.

## Decisions

- **Version source:** ldflags via `git describe --tags`. No source file to edit at release time.
- **Output:** Commit + annotated tag. User pushes manually.
- **Gating:** Run `task check` before proceeding. Abort on failure.
- **Approach:** Single-pass linear. No dry-run phase — Claude Code already shows edits for review.

## Command Interface

```
/release           → patch bump (0.1.0 → 0.1.1)
/release minor     → minor bump (0.1.1 → 0.2.0)
/release major     → major bump (0.2.0 → 1.0.0)
```

First release (no tags): starts at `v0.1.0`.

## Steps

### 1. Pre-flight checks

- Run `task check` (fmt + lint + test-race). Abort on failure.
- Verify working tree is clean.
- Verify on `main` branch.

### 2. Determine version

- Find latest semver tag: `git tag -l 'v*' | sort -V | tail -1`
- No tags → treat as `v0.0.0` (first release becomes `v0.1.0`)
- Apply bump type, display `Bumping v0.1.0 → v0.2.0`

### 3. Analyze changes

- Collect commits: `git log --oneline <last-tag>..HEAD`
- Categorize by conventional commit prefix:
  - `feat:` → **Added**
  - `fix:` → **Fixed**
  - `refactor:`, `perf:` → **Changed**
  - `test:`, `docs:`, `chore:`, `ci:` → **Other**
- Parse test count from `task test` output

### 4. Create or update CHANGELOG.md

[Keep a Changelog](https://keepachangelog.com) format:

```markdown
## [0.2.0] - 2026-02-18

### Added
- Add package_manager.preferred config option

### Fixed
- Resolve gosec false positives for CLI taint analysis

### Changed
- Extract validation logic to separate package
```

### 5. Update README.md

- Update test count if mentioned
- Update version references
- Add new CLI commands or flags if applicable

### 6. Commit and tag

- Stage CHANGELOG.md and README.md
- Commit: `chore: release v0.2.0`
- Annotated tag: `git tag -a v0.2.0 -m "Release v0.2.0"`

### 7. Summary

- Display version bump (old → new)
- List categorized changes
- Remind: `Run 'git push && git push --tags' when ready`

## Taskfile Change

Add ldflags to the build task so tagged versions propagate to the binary:

```yaml
build:
  vars:
    VERSION:
      sh: git describe --tags --always --dirty 2>/dev/null || echo dev
  cmds:
    - mkdir -p {{.BIN_DIR}}
    - go build -ldflags "-X main.version={{.VERSION}}" -o {{.BINARY_PATH}} {{.MAIN_PATH}}
```

After tagging and building, `cc-tools --version` reports the tagged version.

## Files Modified

| File | Change |
|------|--------|
| `.claude/commands/release.md` | Rewrite for Go project |
| `Taskfile.yml` | Add ldflags to build task |
| `CHANGELOG.md` | Created on first release |

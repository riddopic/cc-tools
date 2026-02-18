# Release Command Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Rewrite the `/release` slash command for the Go cc-tools project and add version stamping via ldflags.

**Architecture:** Two changes: (1) add ldflags to the Taskfile build task so `git describe` stamps the binary version, and (2) rewrite `.claude/commands/release.md` as a Go-project-appropriate release workflow.

**Tech Stack:** Task (build), git tags (versioning), conventional commits (changelog generation)

---

### Task 1: Add ldflags to Taskfile build task

**Files:**
- Modify: `Taskfile.yml:71-82` (build task)

**Step 1: Edit the build task to include ldflags**

Replace the current build task:

```yaml
  build:
    desc: Build the binary
    aliases: [q]
    sources:
      - "**/*.go"
      - go.mod
      - go.sum
    generates:
      - "{{.BINARY_PATH}}"
    cmds:
      - mkdir -p {{.BIN_DIR}}
      - go build -o {{.BINARY_PATH}} {{.MAIN_PATH}}
```

With:

```yaml
  build:
    desc: Build the binary
    aliases: [q]
    sources:
      - "**/*.go"
      - go.mod
      - go.sum
    generates:
      - "{{.BINARY_PATH}}"
    vars:
      VERSION:
        sh: git describe --tags --always --dirty 2>/dev/null || echo dev
    cmds:
      - mkdir -p {{.BIN_DIR}}
      - go build -ldflags "-X main.version={{.VERSION}}" -o {{.BINARY_PATH}} {{.MAIN_PATH}}
```

**Step 2: Verify the build works**

Run: `task build`
Expected: Binary builds without error.

Run: `./bin/cc-tools --version`
Expected: Output includes `dev` or a commit hash (no tags exist yet).

**Step 3: Commit**

```bash
git add Taskfile.yml
git commit -m "feat: stamp binary version via ldflags from git describe"
```

---

### Task 2: Rewrite the release command

**Files:**
- Modify: `.claude/commands/release.md`

**Step 1: Replace the command file contents**

Write the following to `.claude/commands/release.md`:

````markdown
---
description: Prepare a release by running checks, updating docs, and tagging the version.
---

Prepare a release for cc-tools. The bump type is: $ARGUMENTS (default to `patch` if empty).

## Steps

### 1. Pre-flight checks

- Run `task check` (fmt + lint + test-race). **Stop immediately if anything fails.**
- Run `git status` to verify the working tree is clean. If there are uncommitted changes, stop and ask me to commit or stash them first.
- Run `git branch --show-current` to verify we are on `main`. If not, stop and ask me to switch.

### 2. Determine version

- Find the latest semver tag: `git tag -l 'v*' --sort=-version:refname | head -1`
- If no tags exist, treat the current version as `v0.0.0`
- Apply the bump type (`major`, `minor`, or `patch`) to calculate the new version
- Display: `Bumping vX.Y.Z → vA.B.C`

### 3. Analyze changes since last release

- Run `git log --oneline <last-tag>..HEAD` (or all commits if no tag exists)
- Categorize each commit by its conventional commit prefix:
  - `feat:` → **Added**
  - `fix:` → **Fixed**
  - `refactor:`, `perf:` → **Changed**
  - `test:`, `docs:`, `chore:`, `ci:` → **Other**
- Run `task test` and capture the test count from the output

### 4. Create or update CHANGELOG.md

- If `CHANGELOG.md` does not exist, create it with this header:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com),
and this project adheres to [Semantic Versioning](https://semver.org).

## [Unreleased]
```

- Add a new version section **below** `## [Unreleased]` with today's date
- Group commits into `### Added`, `### Fixed`, `### Changed`, `### Other` subsections
- Only include subsections that have entries
- Add comparison links at the bottom of the file:
  - `[Unreleased]` compares the new tag to HEAD
  - New version compares to the previous tag (or the initial commit if first release)

### 5. Update README.md

- If there are new CLI commands or flags, add them to the Commands section
- If there are architecture changes worth noting, update the relevant description
- Do not update test counts — these change too frequently

### 6. Commit and tag

- Stage only the files that were modified (CHANGELOG.md, README.md, and any other updated docs)
- Commit with message: `chore: release vA.B.C`
- Create an annotated tag: `git tag -a vA.B.C -m "Release vA.B.C"`

### 7. Present summary

Show:
- Version bump: `vX.Y.Z → vA.B.C`
- Number of commits included
- Categorized change summary (Added/Fixed/Changed/Other)
- Test count from the test run
- Reminder: `Run 'git push && git push --tags' when ready to publish`

Do NOT push to the remote. The user will push manually.
````

**Step 2: Verify the command renders correctly**

Run: `/release` in a Claude Code session to confirm it loads without syntax errors.
(Manual verification — this is a markdown file, not executable code.)

**Step 3: Commit**

```bash
git add .claude/commands/release.md
git commit -m "feat: rewrite release command for Go project"
```

---

### Task 3: Verify end-to-end

**Step 1: Build with ldflags**

Run: `task build && ./bin/cc-tools --version`
Expected: Version output shows commit hash (since no tags exist yet).

**Step 2: Verify install task still works**

Run: `task install`
Expected: Binary copies to `$GOPATH/bin` without error.

**Step 3: Final commit (if any fixups needed)**

If any adjustments were required, commit them as a `fix:` commit.

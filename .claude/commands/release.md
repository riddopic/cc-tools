---
description: Prepare a release by running checks, updating docs, and tagging the version.
---

Prepare a release for cc-tools. The bump type is: $ARGUMENTS (default to `patch` if empty).
If the bump type is not one of `major`, `minor`, or `patch`, ask me to clarify.

## Steps

### 1. Pre-flight checks

- Run `task check` (fmt + lint + test-race). **Stop immediately if anything fails.**
- Run `git status` to verify the working tree is clean. If there are uncommitted changes, stop and ask me to commit or stash them first.
- Run `git branch --show-current` to verify we are on `main`. If not, stop and ask me to switch.

### 2. Determine version

- Find the latest semver tag: `git tag -l 'v*' --sort=-version:refname | head -1`
- If no tags exist, treat the current version as `v0.0.0`
- Apply the bump type to calculate the new version:
  - `major`: increment first number, reset others to 0 (v1.2.3 → v2.0.0)
  - `minor`: increment second number, reset third to 0 (v1.2.3 → v1.3.0)
  - `patch`: increment third number (v1.2.3 → v1.2.4)
- Display: `Bumping vX.Y.Z → vA.B.C`

### 3. Analyze changes since last release

- Run `git log --oneline <last-tag>..HEAD` (or all commits if no tag exists)
- Categorize each commit by its conventional commit prefix:
  - `feat:` → **Added**
  - `fix:` → **Fixed**
  - `refactor:`, `perf:` → **Changed**
  - `test:`, `docs:`, `chore:`, `ci:` → **Other**
- Run `task test` and capture the test count from the output (`task test` uses `gotestsum`; parse the "N passed" count from its summary line)

### 4. Create or update CHANGELOG.md

- If `CHANGELOG.md` does not exist, create it with this header:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com),
and this project adheres to [Semantic Versioning](https://semver.org).

## [Unreleased]
```

- Add a new version section **below** `## [Unreleased]` with today's date (format: `YYYY-MM-DD`)
- Group commits into `### Added`, `### Fixed`, `### Changed`, `### Other` subsections
- Only include subsections that have entries
- Add comparison links at the bottom of the file:
  - `[Unreleased]` compares the new tag to HEAD
  - New version compares to the previous tag (or the initial commit if first release)
  - Example format:
    ```
    [Unreleased]: https://github.com/riddopic/cc-tools/compare/vA.B.C...HEAD
    [A.B.C]: https://github.com/riddopic/cc-tools/compare/vX.Y.Z...vA.B.C
    ```

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

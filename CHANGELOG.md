# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com),
and this project adheres to [Semantic Versioning](https://semver.org).

## [Unreleased]

## [0.1.2] - 2026-02-24

### Added

- Instinct command group (`cc-tools instinct`) with status, export, import, and evolve subcommands
- File-based instinct store with list, get, save, and delete operations
- Confidence scoring and decay for instincts
- Instinct types and YAML frontmatter parser
- YAML and JSON instinct export formats
- Instinct configuration keys and defaults
- Import, Evolve, and Export functions in instinct package
- ToolOutput and Error fields in observe events, passed through ObserveHandler
- Skill-audit infrastructure for evidence-based skill quality evaluation

### Fixed

- Stale skill references across commands and skills updated to current names

### Changed

- Consolidated skills and updated loading policy based on SkillsBench research
- Trimmed CLAUDE.md and rules based on AGENTS.md research findings
- Trimmed 15 skills from Comprehensive to Detailed/Compact format
- Merged coding-philosophy and review-verification-protocol into parent skills
- Simplified instinct command by delegating to internal/instinct package
- Rewrote codex-review skill addressing 24 review findings
- Updated instinct commands to use `cc-tools instinct` instead of Python/Bash scripts
- Removed redundant observe.sh hooks from settings.json
- Removed obsolete Python/Bash instinct scripts
- Removed obsolete PRP workflow commands

### Other

- Diataxis-structured documentation added with updated README
- Skill lifecycle checklist from audit learnings
- Instinct migration design and implementation plan
- Updated spellcheck dictionary and test alignment
- 1210 tests with race detector coverage

## [0.1.1] - 2026-02-23

### Added

- Drift detection handler for UserPromptSubmit events — tracks session intent from the first prompt, extracts keywords, and warns when subsequent prompts diverge below the configured overlap threshold
- Stop reminder handler for Stop events — counts responses per session and emits rotating reminders at configurable intervals with a strong wrap-up warning
- Six new configuration keys: `drift.enabled`, `drift.min_edits`, `drift.threshold`, `stop_reminder.enabled`, `stop_reminder.interval`, `stop_reminder.warn_at`
- `learn-eval` command for extracting reusable patterns with self-evaluation and location-aware saving
- `search-first` skill for research-before-coding workflows
- `function-analyzer` agent for security audit deep code analysis

### Fixed

- Hook integration JSON example in README now uses correct nested format matching the Claude Code settings schema
- Stop event in project settings now dispatches through cc-tools hook alongside existing evaluate-session script

### Changed

- Removed dead code from internal/shared/colors.go

### Other

- Raised cmd/cc-tools test coverage from 49.7% to 78.8%
- 1108 tests with race detector coverage
- Removed 19 stale plan and audit documents
- Updated skill cross-references for search-first and function-analyzer

## [0.1.0] - 2026-02-18

Initial release of cc-tools, a CLI companion for Claude Code.

### Added

- Hook dispatch system with handler registry routing all Claude Code hook events
- Parallel lint and test validation via `validate` command with cooldown locks
- Notification backends: audio playback, macOS desktop alerts, ntfy push notifications with quiet hours
- Session tracking with metadata storage, aliases, and search
- Tool use observation logging to filesystem with rotation
- Context compaction suggestions and log tracking
- MCP server enable/disable management
- Per-directory skip configuration for lint, test, or both
- Package manager detection (npm, pnpm, bun, yarn) with config override
- Styled terminal output via charmbracelet/lipgloss
- Cobra CLI with subcommands: hook, validate, session, config, skip, unskip, mcp, debug, version
- Build-time version stamping via ldflags from git describe
- `/release` slash command for version management

### Fixed

- Gosec false positives for CLI taint analysis
- Directory traversal rejection in skip command paths
- Temp file cleanup in debug logging
- Process existence check using syscall.Signal(0)

### Changed

- Migrated all hook scripts from JS/Python/Bash to unified Go binary
- Consolidated configuration to ~/.config/cc-tools/config.json
- Unified filesystem interfaces into shared package (HooksFS, RegistryFS, FS)

### Other

- 1001 tests with race detector coverage
- 8 hook parity test suites validating handler behavior
- Comprehensive linting with 60+ golangci-lint rules
- Mockery v3.5 mock generation for all interfaces
- Architecture design docs and implementation plans

[Unreleased]: https://github.com/riddopic/cc-tools/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/riddopic/cc-tools/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/riddopic/cc-tools/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/riddopic/cc-tools/commits/v0.1.0

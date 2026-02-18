# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com),
and this project adheres to [Semantic Versioning](https://semver.org).

## [Unreleased]

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

[Unreleased]: https://github.com/riddopic/cc-tools/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/riddopic/cc-tools/commits/v0.1.0

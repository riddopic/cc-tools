---
description: Sync documentation from source-of-truth files (go.mod, Makefile, config)
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Edit
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
skills:
  - go-coding-standards
---

# Update Documentation

Sync documentation from source-of-truth files.

## Source-of-Truth Files

| File | Contains |
|------|----------|
| `go.mod` | Module name, Go version, dependencies |
| `Makefile` | Available commands and targets |
| `~/.quanta/` | User configuration directory |
| `.quanta.yaml` | Project configuration |

## Required Standards

Follow the coding guidelines in `docs/CODING_GUIDELINES.md`:
- CLI patterns from `docs/examples/patterns/cli.md`
- Documentation standards from `docs/examples/standards/documentation.md`

## Execution Steps

### 1. Extract Commands from Makefile

```bash
# List all make targets with descriptions
make help
```

Generate a commands reference table:

```markdown
| Command | Description |
|---------|-------------|
| `make build` | Build binary with version info |
| `make test` | Run unit tests |
| ... | ... |
```

### 2. Extract Dependencies from go.mod

```bash
# Show module info
go list -m

# Show direct dependencies
go list -m -f '{{.Path}} {{.Version}}' all | head -20
```

Document key dependencies:

```markdown
## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/spf13/cobra | vX.X.X | CLI framework |
| github.com/spf13/viper | vX.X.X | Configuration |
| ... | ... | ... |
```

### 3. Document Configuration

Extract configuration options from Viper setup:

```markdown
## Configuration

### File Locations
- `~/.quanta/config.yaml` - User configuration
- `.quanta.yaml` - Project configuration

### Environment Variables
- `QUANTA_*` - Viper auto-binds env vars with QUANTA_ prefix

### Configuration Hierarchy
1. Command-line flags (highest priority)
2. Environment variables
3. Project config file (.quanta.yaml)
4. User config file (~/.quanta/config.yaml)
5. Defaults (lowest priority)
```

### 4. Update User Guide Documentation

Update command-specific guides in `docs/USER-GUIDE/`:

```markdown
# Command Documentation Structure

docs/USER-GUIDE/
├── INDEX.md                 # Documentation index
├── GETTING-STARTED.md       # Getting started guide
├── CONFIGURATION.md         # Configuration reference
├── ANALYZE.md               # quanta analyze command
├── CALIBRATE.md             # quanta calibrate command
├── DOCTOR.md                # quanta doctor command
├── EXPORT.md                # quanta export command
├── HISTORY.md               # quanta history command
├── REGRESSION.md            # quanta regression command
├── RUN.md                   # quanta run command
└── [COMMAND].md             # Other command guides
```

For each command guide, ensure:
- Synopsis matches current CLI help output
- Flags are current with implementation
- Examples are working and up-to-date

### 5. Update Getting Started Guide

Update `docs/USER-GUIDE/GETTING-STARTED.md` with current setup instructions:

```markdown
# Getting Started

## Installation

### Building from Source
\`\`\`bash
make build
make install
\`\`\`

## Quick Start

1. Run `quanta doctor` to verify environment
2. Configure API keys in `~/.quanta/config.yaml`
3. Run `quanta analyze <contract>` to analyze a contract

## Common Issues

### Issue: Build fails with missing dependencies
Solution: Run `go mod tidy`

### Issue: Tests fail with race conditions
Solution: Run `make test-race` to identify, fix concurrent access

## Next Steps

See `docs/USER-GUIDE/INDEX.md` for full documentation index.
```

### 6. Identify Obsolete Documentation

```bash
# Find USER-GUIDE docs not modified in 90+ days
find docs/USER-GUIDE/ -name "*.md" -mtime +90 -type f
```

List for manual review. Check if command still exists or documentation needs refresh.

### 7. Show Diff Summary

```bash
# Show changes to docs
git diff --stat docs/
```

### 8. Build Hugo Documentation

After updating markdown files, build the Hugo documentation site:

```bash
# Generate CLI documentation for Hugo
make docs-cli

# Build Hugo static site
make docs-build
```

This ensures the Hugo site reflects all documentation changes.

## Commands

| Command | Purpose |
|---------|---------|
| `make help` | List all available targets |
| `go list -m all` | List all dependencies |
| `make doctor` | Check development environment |
| `make docs-cli` | Generate CLI documentation for Hugo |
| `make docs-build` | Build Hugo static documentation site |

## Output Files

| File | Purpose |
|------|---------|
| `docs/USER-GUIDE/*.md` | Command-specific user guides |
| `docs/USER-GUIDE/INDEX.md` | Documentation index |
| `docs/USER-GUIDE/GETTING-STARTED.md` | Getting started guide |
| `docs/USER-GUIDE/CONFIGURATION.md` | Configuration reference |
| `README.md` | Project overview (update commands section) |

## Best Practices

- Keep documentation close to source-of-truth
- Automate extraction where possible
- Mark generated sections with comments
- Review obsolete docs quarterly

## Integration with Other Commands

- Run after adding new make targets
- Run after updating dependencies
- Run before releases
- After updating USER-GUIDE docs, always run Hugo build

---
description: Analyze codebase structure and update architecture documentation
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
skills:
  - go-coding-standards
  - interface-design
---

# Update Codemaps

Analyze the Go codebase structure and update architecture documentation.

## Required Standards

Follow the coding guidelines in `docs/CODING_GUIDELINES.md`:
- Go project structure: `cmd/`, `internal/`, `pkg/`
- Interface-first design
- Clear package boundaries

## Execution Steps

### 1. Analyze Package Structure

```bash
# List all packages
go list ./...

# Show package dependencies
go mod graph

# Show import graph for a specific package
go list -f '{{.ImportPath}} -> {{.Imports}}' ./internal/...
```

### 2. Scan Source Files

Analyze all source files for:
- Package declarations
- Interface definitions
- Struct definitions
- Function signatures
- Import dependencies

```bash
# Count lines per package
find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | sort -n

# Find interface definitions
rg "^type \w+ interface" --type go

# Find struct definitions
rg "^type \w+ struct" --type go
```

### 3. Generate Codemaps

Create token-lean codemaps in the following structure:

#### `docs/codemaps/architecture.md` - Overall Architecture

```markdown
# Quanta Architecture

## Overview
[High-level system description]

## Package Structure
- cmd/           - CLI entry points
- internal/      - Private packages
- pkg/           - Public packages (if any)

## Key Components
[Component diagram or list]

## Data Flow
[Request/response flow]
```

#### `docs/codemaps/packages.md` - Package Structure

```markdown
# Package Structure

## cmd/
[CLI commands and entry points]

## internal/
[Internal packages with descriptions]

## Dependencies
[Key external dependencies from go.mod]
```

#### `docs/codemaps/interfaces.md` - Interface Definitions

```markdown
# Interfaces

## Core Interfaces
[List of primary interfaces with methods]

## Service Interfaces
[Service layer interfaces]

## Repository Interfaces
[Data access interfaces]
```

#### `docs/codemaps/dependencies.md` - Module Dependencies

```markdown
# Dependencies

## Direct Dependencies
[From go.mod require block]

## Dependency Graph
[Key dependency relationships]
```

### 4. Calculate Diff Percentage

```bash
# Compare with previous version
git diff --stat docs/codemaps/
```

If changes > 30%, request user approval before updating.

### 5. Add Freshness Timestamp

Each codemap should include:

```markdown
---
Generated: YYYY-MM-DD HH:MM:SS
Quanta Version: vX.X.X
---
```

### 6. Save Diff Report

Save analysis report to `.reports/codemap-diff.txt`:

```
Codemap Update Report
=====================
Date: YYYY-MM-DD
Files Changed: N
Lines Added: +X
Lines Removed: -Y
Change Percentage: Z%

Changes by File:
- architecture.md: [summary]
- packages.md: [summary]
- interfaces.md: [summary]
- dependencies.md: [summary]
```

## Commands

| Command | Purpose |
|---------|---------|
| `go list ./...` | List all packages |
| `go mod graph` | Show dependency graph |
| `make build` | Verify build still works |

## Analysis Tools

```bash
# Package dependencies
go list -f '{{.ImportPath}}: {{.Imports}}' ./...

# Find circular dependencies
go list -f '{{.ImportPath}} {{.Imports}}' ./... | tsort

# Count exported vs unexported
rg "^func [A-Z]" --type go -c  # Exported
rg "^func [a-z]" --type go -c  # Unexported
```

## Output Structure

```
docs/codemaps/
├── architecture.md    # Overall architecture
├── packages.md        # Package structure (cmd/, internal/, pkg/)
├── interfaces.md      # Interface definitions
└── dependencies.md    # Module dependencies

.reports/
└── codemap-diff.txt   # Change report
```

## Best Practices

- Focus on high-level structure, not implementation details
- Keep codemaps token-lean for LLM consumption
- Update after significant architectural changes
- Version control the codemaps

## Integration with Other Commands

- Use after major refactoring
- Run before architecture reviews
- Update when adding new packages

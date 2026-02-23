---
name: search-first
description: Research-before-coding workflow. Search existing codebase, Go packages, and MCP tools before writing custom code.
user-invocable: true
---

# /search-first — Research Before You Code

Systematizes the "search for existing solutions before implementing" workflow.

## Trigger

Use this skill when:
- Starting a new feature that likely has existing solutions
- Adding a dependency or integration
- The user asks "add X functionality" and you're about to write code
- Before creating a new utility, helper, or abstraction

## Workflow

```
┌─────────────────────────────────────────────┐
│  1. NEED ANALYSIS                           │
│     Define what functionality is needed     │
│     Identify Go / project constraints       │
├─────────────────────────────────────────────┤
│  2. SEARCH (codebase first, then external)  │
│     ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│     │ internal/│ │ pkg.go.  │ │ MCP /    │  │
│     │ + stdlib │ │ dev      │ │ Skills   │  │
│     └──────────┘ └──────────┘ └──────────┘  │
├─────────────────────────────────────────────┤
│  3. EVALUATE                                │
│     Score candidates (functionality, maint, │
│     community, docs, license, deps)         │
├─────────────────────────────────────────────┤
│  4. DECIDE                                  │
│     ┌─────────┐  ┌──────────┐  ┌─────────┐  │
│     │  Adopt  │  │  Extend  │  │  Build  │  │
│     │ as-is   │  │  /Wrap   │  │  Custom │  │
│     └─────────┘  └──────────┘  └─────────┘  │
├─────────────────────────────────────────────┤
│  5. IMPLEMENT                               │
│     Use existing package / Configure MCP /  │
│     Write minimal custom code               │
└─────────────────────────────────────────────┘
```

## Decision Matrix

| Signal | Action |
|--------|--------|
| Exact match, well-maintained, MIT/Apache/BSD | **Adopt** — `go get` and use directly |
| Partial match, good foundation | **Extend** — `go get` + write thin wrapper |
| Multiple weak matches | **Compose** — combine 2-3 small packages |
| Nothing suitable found | **Build** — write custom, but informed by research |

## How to Use

### Quick Mode (inline)

Before writing a utility or adding functionality, run through:

1. Does the Go standard library cover this? → Check `pkg.go.dev/std`
2. Does an `internal/` package already do this? → `rg "FunctionName" internal/`
3. Is there a Go package for this? → Search `pkg.go.dev`
4. Is there an MCP server or skill for this? → Check available MCP tools and `.claude/skills/`
5. Is there a reference implementation? → Search GitHub

### Full Mode (agent)

For non-trivial functionality, launch a research agent:

```
Task(subagent_type="general-purpose", prompt="
  Research existing tools for: [DESCRIPTION]
  Language: Go
  Constraints: [ANY]

  Search order:
  1. Go standard library
  2. pkg.go.dev / awesome-go
  3. Existing internal/ packages in this repo
  4. MCP servers and Claude Code skills
  5. GitHub reference implementations

  Return: Structured comparison with recommendation
")
```

## Search Shortcuts by Category

### Go Standard Library (check first)

- HTTP → `net/http`, `net/http/httptest`
- JSON → `encoding/json`
- Concurrency → `sync`, `context`, `errgroup`
- Testing → `testing`, `testing/fstest`, `io/fs`
- CLI args → already using Cobra/Viper
- Crypto → `crypto/*`
- Text processing → `strings`, `text/template`, `regexp`
- I/O → `io`, `bufio`, `os`

### Go Ecosystem

- Linting → `golangci-lint` (already configured)
- Formatting → `gofmt`, `goimports` (already configured)
- Testing → `testify` (already a dependency), `go test`
- Mocking → `mockery` (already configured, v3.5)
- HTTP clients → `net/http` + `internal/proxy/factory.go` (Tor-aware)
- CLI → `cobra`, `viper` (already dependencies)
- TUI → `bubbletea`, `lipgloss`, `bubbles` (already dependencies)
- Logging → `zap` (already a dependency)

### AI/LLM Integration

- Claude SDK docs → Context7 MCP for latest docs
- Prompt management → Check existing `internal/llm/gates/prompts/`

### Project-Specific

- HTTP clients → **Must use** `cmd.CreateHTTPClient()` (Tor-aware, see networking rules)
- Config → Use `internal/config/` typed config, not raw Viper
- Styling → Use `internal/tui/styles/` centralized styles

## Integration Points

### With Plan agent

The Plan agent should invoke search-first before architecture decisions:
- Identify available tools and existing packages
- Incorporate them into the implementation plan
- Avoid reimplementing what already exists in `internal/`

### With Explore agent

Combine for progressive discovery:
- Cycle 1: Search codebase (`rg`, `mgrep`) and Go stdlib
- Cycle 2: Evaluate external candidates on pkg.go.dev
- Cycle 3: Test compatibility with project constraints

## Examples

### Example 1: "Add retry logic to HTTP calls"
```
Need: Resilient HTTP client with retries
Search: Go stdlib (net/http), internal/ packages
Found: internal/proxy/factory.go already creates Tor-aware clients
Also: hashicorp/go-retryablehttp on pkg.go.dev (7k+ stars)
Action: EXTEND — wrap existing factory with retry middleware
Result: Thin wrapper over existing infrastructure
```

### Example 2: "Add structured logging to a new package"
```
Need: Structured logging in a new internal/ package
Search: Existing project patterns
Found: Project uses zap throughout, internal/cli/ has logger setup
Action: ADOPT — use zap.Logger injection pattern from existing packages
Result: Zero new dependencies, consistent with codebase
```

### Example 3: "Add YAML config validation"
```
Need: Validate YAML configuration files against expected structure
Search: Go stdlib, pkg.go.dev "yaml schema validation"
Found: go-playground/validator (already indirect dep), mapstructure tags
Also: internal/config/ already uses Viper with struct tags
Action: EXTEND — add validate struct tags to existing config types
Result: No new dependencies, leverages existing Viper + validator
```

## Anti-Patterns

- **Jumping to code**: Writing a utility without checking if one exists in `internal/` or stdlib
- **Ignoring MCP**: Not checking if an MCP server already provides the capability
- **Ignoring project conventions**: Using `http.Client{}` directly instead of `cmd.CreateHTTPClient()`
- **Over-customizing**: Wrapping a library so heavily it loses its benefits
- **Dependency bloat**: Adding a large dependency for one small feature when stdlib suffices

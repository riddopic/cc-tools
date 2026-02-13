# cc-tools Coding Guidelines

> Comprehensive coding guidelines for the cc-tools project - a Claude Code integration toolkit providing hook validation, skip registries, debug logging, MCP server management, and configuration utilities.

## Table of Contents

1. [Project Philosophy](#project-philosophy)
2. [Project Structure](#project-structure)
3. [Go Language Standards](#go-language-standards)
4. [Naming Conventions](#naming-conventions)
5. [Error Handling](#error-handling)
6. [Documentation Standards](#documentation-standards)
7. [Testing Standards](#testing-standards)
8. [CLI Development](#cli-development)
9. [Performance Guidelines](#performance-guidelines)
10. [Security Practices](#security-practices)

## Project Philosophy

### Core Principles

- **Simplicity First**: Favor simple, obvious solutions over clever ones
- **Explicit Over Implicit**: Make intentions clear in code
- **Composition Over Inheritance**: Use interfaces and embedding
- **Early Returns**: Reduce nesting with guard clauses
- **Small Functions**: Keep functions focused and under 50 lines
- **Testability**: Design code to be easily testable

### Go Idioms We Follow

- Errors are values, not exceptions
- Don't communicate by sharing memory; share memory by communicating
- The zero value should be useful
- Accept interfaces, return concrete types
- Make the zero value useful
- A little copying is better than a little dependency

## Project Structure

```text
cc-tools/
├── go.mod                   # Module definition (github.com/riddopic/cc-tools)
├── go.sum                   # Dependency checksums
├── cmd/
│   └── cc-tools/            # CLI application entry point
│       ├── main.go          # Main CLI dispatcher
│       ├── config.go        # Configuration management command
│       ├── debug.go         # Debug logging utilities
│       ├── mcp.go           # MCP server management command
│       └── skip.go          # Skip registry commands
├── internal/                # Private application code
│   ├── config/              # Configuration management
│   │   ├── config.go        # Configuration structures
│   │   ├── manager.go       # Configuration loading/saving
│   │   └── *_test.go        # Tests
│   ├── hooks/               # Hook validation and execution
│   │   └── *.go             # Hook processing logic
│   ├── mcp/                 # MCP server integration
│   │   ├── mcp.go           # MCP server management
│   │   └── *_test.go        # Tests
│   ├── skipregistry/        # Skip tool registry
│   │   └── *.go             # Skip registry logic
│   ├── debug/               # Debug logging
│   │   └── *.go             # Debug utilities
│   ├── output/              # Terminal output formatting
│   │   ├── output.go        # Output abstraction
│   │   ├── table.go         # Table rendering
│   │   └── hook.go          # Hook output formatting
│   └── shared/              # Shared utilities
│       ├── fs.go            # Filesystem abstractions
│       ├── debug_paths.go   # Path resolution
│       └── mocks/           # Test mocks
├── docs/                    # Documentation
│   ├── CODING_GUIDELINES.md # This document
│   └── examples/            # Usage examples
├── testdata/                # Test fixtures
└── .github/                 # GitHub configuration
    └── workflows/           # CI/CD workflows
```

### Directory Guidelines

- **`/internal`**: Use for all private application code
- **`/cmd`**: Keep command files small, delegate to internal packages
- **`/pkg`**: Only create if you explicitly want to share code
- **`/testdata`**: Store all test fixtures here
- **Never create empty directories** - add them when needed

## Go Language Standards

### Language Version

- **Go 1.26** is the current version
- Enable Go modules: `GO111MODULE=on`

### Code Formatting

```bash
# Install required tools
task tools-install

# Format code
task fmt          # Runs gofmt and goimports
task polish       # Format, auto-fix lint issues, clean backup files
```

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"
    "io"

    // Third-party packages
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "go.uber.org/zap"

    // Internal packages
    "github.com/riddopic/cc-tools/internal/config"
    "github.com/riddopic/cc-tools/internal/hooks"
)
```

## Naming Conventions

### Packages

- Use singular nouns: `user` not `users`
- Short, lowercase: `config` not `configuration`
- No underscores or mixed caps

### Files

- Lowercase with underscores: `status_line.go`
- Test files: `status_line_test.go`
- Group related functionality

### Go Identifiers

```go
// Exported (public)
type StatusLine struct {}
func NewStatusLine() *StatusLine {}
const MaxRetries = 3

// Unexported (private)
type renderer struct {}
func parseConfig() error {}
const defaultTimeout = 30
```

### Interfaces

```go
// Good - behavior focused
type Reader interface {}
type Validator interface {}
type Renderer interface {}

// Bad - noun focused or Hungarian notation
type IReader interface {}
type ReaderInterface interface {}
type DataInterface interface {}
```

## Error Handling

### Error Creation and Wrapping

```go
// Define cc-tools errors
var (
    ErrConfigNotFound = errors.New("configuration file not found")
    ErrInvalidHook    = errors.New("invalid hook configuration")
)

// Wrap errors with context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &cfg, nil
}
```

### Error Checking

```go
// Check and handle immediately
resp, err := client.Get(url)
if err != nil {
    return fmt.Errorf("failed to fetch data: %w", err)
}
defer resp.Body.Close()

// Check specific errors
if errors.Is(err, ErrConfigNotFound) {
    // Handle missing config
}

// Check error types
var syntaxErr *json.SyntaxError
if errors.As(err, &syntaxErr) {
    // Handle syntax error
}
```

### Error Messages

- Start with lowercase (unless proper noun)
- No punctuation at the end
- Include context about what failed
- Be actionable when possible

```go
// Good
"failed to connect to database"
"invalid configuration: theme 'dark' not found"
"timeout waiting for response"

// Bad
"Error connecting to database."
"Failed"
"Something went wrong"
```

## Documentation Standards

### Package Documentation

```go
// Package hooks provides validation and execution utilities
// for Claude Code hook configurations.
//
// It supports PreToolUse, PostToolUse, and Stop hooks with
// pattern matching and parameter validation.
package hooks
```

### Function Documentation

```go
// ValidateHook validates a hook configuration against the registered schema.
// It checks pattern matching, required parameters, and execution permissions.
//
// The hook configuration must specify a valid matcher and command.
// If the configuration is invalid, an error is returned.
//
// Example:
//
//  hook := &Hook{Matcher: "Edit", Command: "prettier"}
//  err := ValidateHook(hook)
//  if err != nil {
//      log.Fatal(err)
//  }
func ValidateHook(hook *Hook) error {
    // Implementation
}
```

### Type Documentation

```go
// Hook represents a Claude Code hook configuration entry.
// It defines when and how a hook script should be executed.
//
// Hooks can be PreToolUse, PostToolUse, or Stop type, with
// pattern matchers for tool names or output content.
type Hook struct {
    Matcher string
    Command string
    Type    HookType
}
```

### Field Documentation

```go
type Config struct {
    // DebugEnabled controls whether debug logging is enabled.
    // When true, logs are written to ~/.claude/logs/
    DebugEnabled bool `json:"debug_enabled"`

    // MCPServers contains the list of configured MCP server endpoints.
    // Each server must specify a name and connection details.
    MCPServers []MCPServer `json:"mcp_servers"`

    // SkipRegistry defines tools that should skip permission prompts.
    // Tools are identified by name and can have optional glob patterns.
    SkipRegistry []string `json:"skip_registry"`
}
```

## Testing Standards

### Table-Driven Tests

```go
func TestRenderer_Render(t *testing.T) {
    tests := []struct {
        name    string
        input   *StatusData
        theme   Theme
        want    string
        wantErr bool
    }{
        {
            name: "default theme with basic data",
            input: &StatusData{
                SessionID: "abc123",
                Status:    "active",
            },
            theme: DefaultTheme,
            want:  "Session: abc123 | Status: active",
        },
        {
            name:    "nil input returns error",
            input:   nil,
            theme:   DefaultTheme,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := NewRenderer(tt.theme)
            got, err := r.Render(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Mock Setup

```go
func TestClaudeClient_GetSession(t *testing.T) {
    tests := []struct {
        name      string
        sessionID string
        setupMock func(*mocks.MockHTTPClient)
        want      *Session
        wantErr   bool
    }{
        {
            name:      "successful fetch",
            sessionID: "test-123",
            setupMock: func(m *mocks.MockHTTPClient) {
                response := &http.Response{
                    StatusCode: 200,
                    Body:       io.NopCloser(strings.NewReader(`{"id":"test-123","status":"active"}`)),
                }
                m.On("Get", "https://api.claude.ai/sessions/test-123").Return(response, nil)
            },
            want: &Session{
                ID:     "test-123",
                Status: "active",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(mocks.MockHTTPClient)
            tt.setupMock(mockClient)

            client := NewClaudeClient(mockClient)
            got, err := client.GetSession(tt.sessionID)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }

            mockClient.AssertExpectations(t)
        })
    }
}
```

### Test Organization

- Keep unit tests next to code
- Use `testdata/` for fixtures
- Group integration tests separately
- Run tests in parallel when possible
- Use `t.Helper()` for test utilities

### Test Isolation

Test isolation prevents state leakage between tests. Follow these patterns:

**File-Based Tests - Use Temporary Directories**
```go
func setupTestConfig(t *testing.T) string {
    t.Helper()
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.json")

    // Write test config
    data := []byte(`{"debug_enabled": true}`)
    err := os.WriteFile(configPath, data, 0644)
    require.NoError(t, err)

    return configPath
}

func TestConfigLoad(t *testing.T) {
    configPath := setupTestConfig(t)

    cfg, err := LoadConfig(configPath)
    require.NoError(t, err)
    assert.True(t, cfg.DebugEnabled)
}
```

**Rules**:
- Always use `t.Cleanup()` or `defer` for resource cleanup
- Use `t.TempDir()` for file-based tests (auto-cleanup)
- Never reference real user config files like `~/.cc-tools/config.json` in tests
- Use mocks for filesystem operations when appropriate

## CLI Development

### Simple Command Structure

cc-tools uses a simple command dispatcher pattern without Cobra:

```go
// cmd/cc-tools/main.go
func main() {
    out := output.NewTerminal(os.Stdout, os.Stderr)

    if len(os.Args) < minArgs {
        printUsage(out)
        os.Exit(1)
    }

    switch os.Args[1] {
    case "validate":
        runValidate()
    case "skip":
        runSkipCommand()
    case "config":
        runConfigCommand()
    case "mcp":
        runMCPCommand()
    case "debug":
        runDebugCommand()
    default:
        printUsage(out)
        os.Exit(1)
    }
}
```

### Configuration Management

cc-tools uses JSON-based configuration:

```go
// internal/config/config.go
type Config struct {
    DebugEnabled bool         `json:"debug_enabled"`
    MCPServers   []MCPServer  `json:"mcp_servers"`
    SkipRegistry []string     `json:"skip_registry"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &cfg, nil
}

// Environment variables
// CC_TOOLS_DEBUG=true enables debug logging
// CC_TOOLS_CONFIG=/custom/path sets config location
```

### User Interaction

```go
// Respect NO_COLOR environment variable
if os.Getenv("NO_COLOR") != "" {
    color.NoColor = true
}

// Provide helpful error messages
if err := runCommand(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    fmt.Fprintf(os.Stderr, "Run 'cc-tools help' for usage.\n")
    os.Exit(1)
}

// Use appropriate exit codes
const (
    ExitSuccess = 0
    ExitError   = 1
    ExitUsage   = 2
)
```

## Performance Guidelines

### Optimization Principles

1. **Measure First**: Profile before optimizing
2. **Optimize Hot Paths**: Focus on frequently executed code
3. **Reduce Allocations**: Minimize heap allocations
4. **Use Buffering**: Buffer I/O operations
5. **Concurrent When Beneficial**: Use goroutines for parallel work

### Common Optimizations

```go
// Pre-allocate slices
data := make([]string, 0, expectedSize)

// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

// Buffer channels appropriately
ch := make(chan Event, 100) // Buffered channel

// Use strings.Builder for string concatenation
var b strings.Builder
b.WriteString("Session: ")
b.WriteString(sessionID)
result := b.String()
```

### Profiling

```go
// Enable profiling in development
import _ "net/http/pprof"

func init() {
    if os.Getenv("ENABLE_PROFILING") == "true" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
}

// Profile with: go tool pprof http://localhost:6060/debug/pprof/profile
```

## Security Practices

### Input Validation

```go
// Validate all user input
func SetTheme(theme string) error {
    validThemes := []string{"default", "powerline", "minimal", "classic"}

    for _, valid := range validThemes {
        if theme == valid {
            return nil
        }
    }

    return fmt.Errorf("invalid theme: %s", theme)
}

// Sanitize file paths
func LoadConfigFile(path string) error {
    // Prevent directory traversal
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return errors.New("invalid path: directory traversal detected")
    }

    // Check file exists and is readable
    info, err := os.Stat(cleanPath)
    if err != nil {
        return fmt.Errorf("cannot access config file: %w", err)
    }

    if info.IsDir() {
        return errors.New("path is a directory, not a file")
    }

    return nil
}
```

### Sensitive Data

- Never log sensitive information
- Use environment variables for secrets
- Clear sensitive data from memory after use
- Don't commit secrets to version control

## Code Quality Checklist

### Before Committing

- [ ] Code passes formatting checks (`task fmt`)
- [ ] No linter warnings (`task lint`)
- [ ] All tests pass (`task test`)
- [ ] Race detector passes (`task test-race`)
- [ ] New code has appropriate test coverage (`task coverage`)
- [ ] Documentation is updated for public APIs
- [ ] Error messages are clear and actionable
- [ ] No commented-out code
- [ ] No TODO comments without issue references

**Quick check:** Run `task check` to verify all pre-commit checks pass

### Code Review Focus

- [ ] Follows project structure conventions
- [ ] Uses appropriate error handling
- [ ] Has comprehensive test coverage
- [ ] Documentation is clear and complete
- [ ] No unnecessary complexity
- [ ] Follows naming conventions
- [ ] No security vulnerabilities
- [ ] Performance is acceptable

## Quick Reference

### Common Commands

```bash
# Development
task fmt           # Format code (gofmt and goimports)
task lint          # Run golangci-lint
task test          # Run fast unit tests (-short)
task watch         # Auto-run tests on file changes (TDD essential!)
task test-race     # Run tests with race detector
task coverage      # Generate test coverage report
task build         # Build cc-tools binary

# Benchmarking
task bench         # Run benchmarks with memory stats

# Environment & Tools
task doctor        # Check development environment
task tools-install # Install all required tools

# Before commit
task check         # Run all checks (fmt + lint + test-race)

# For profiling (still use go commands directly)
go test -cpuprofile=cpu.prof ./...    # CPU profile
go test -memprofile=mem.prof ./...    # Memory profile
go tool pprof cpu.prof                # Analyze profile
```

### Useful Links

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)

## References

For detailed information on specific topics, see:

- [.claude/rules/](../.claude/rules/) - Project-specific coding rules
- [docs/examples/](docs/examples/) - Example code and patterns
